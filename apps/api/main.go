// Package main boots the Go Opus backend.
package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/FacileStudio/Journal/sdk/journal"
	"github.com/FacileStudio/Opus/apps/api/internal/database"
	"github.com/FacileStudio/Opus/apps/api/internal/env"
	"github.com/FacileStudio/Opus/apps/api/internal/events"
	"github.com/FacileStudio/Opus/apps/api/internal/poolsync"
	"github.com/FacileStudio/Opus/apps/api/internal/realtime"
	"github.com/FacileStudio/Opus/apps/api/internal/scheduler"
	"github.com/FacileStudio/Opus/apps/api/modules/auth"
	"github.com/FacileStudio/porte/local"
	"github.com/FacileStudio/porte/oidc"
	portepg "github.com/FacileStudio/porte/pg"
	"github.com/FacileStudio/porte/session"
	"github.com/FacileStudio/tronc/healthcheck"
	"github.com/FacileStudio/tronc/logger"

	"gorm.io/gorm"
)

type sqlPinger interface {
	PingContext(ctx context.Context) error
}

type authKits struct {
	sessions  *session.Manager
	passwords *local.Kit
	oidc      *oidc.Kit
}

func buildAuth(ctx context.Context, db *gorm.DB, appEnv *env.Config, appLogger *slog.Logger) (*authKits, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	store := portepg.New(sqlDB)
	users := auth.NewUserStore(db)
	cfg := appEnv.Porte()
	backfillPictures(ctx, users, cfg.Issuer, appLogger)

	sessions, err := session.New(cfg, session.Deps{Sessions: store.Sessions(), Logger: appLogger})
	if err != nil {
		return nil, err
	}
	kit, err := oidc.New(ctx, cfg, oidc.Deps{
		Users:      users,
		Identities: store.Identities(),
		Sessions:   sessions,
		Codes:      store.LoginCodes(),
		Logger:     appLogger,
	})
	if err != nil {
		return nil, err
	}
	passwords, err := local.New(local.Config{AllowRegistration: !appEnv.SSOOnly, MinPasswordLength: 12}, local.Deps{
		Users:      users,
		Identities: store.Identities(),
		Sessions:   sessions,
		Logger:     appLogger,
		Count:      users.CountUsers,
	})
	if err != nil {
		return nil, err
	}
	return &authKits{sessions: sessions, passwords: passwords, oidc: kit}, nil
}

// backfillPictures fills the Registre picture URL for accounts that have not
// signed in since Opus started storing it. A failure only costs pictures until
// the next sign-in, so it is logged and boot goes on.
func backfillPictures(ctx context.Context, users *auth.UserStore, issuer string, appLogger *slog.Logger) {
	if issuer == "" {
		return
	}
	updated, err := users.BackfillPictures(ctx, issuer)
	if err != nil {
		appLogger.Warn("picture backfill failed", slog.Any("error", err))
		return
	}
	appLogger.Info("picture backfill applied", slog.Int64("updated", updated))
}

type serverBundle struct {
	server  *http.Server
	cleanup func()
}

func initServices(db *gorm.DB, kits *authKits, appEnv *env.Config, appLogger *slog.Logger) (*realtime.Hub, events.Emitter, func(), error) {
	eventsCfg := events.LoadConfig()
	eventsPub := events.New(eventsCfg, events.WithLogger(appLogger))

	adapter := realtime.NewRedisAdapter(appEnv.RedisURL, appLogger)
	wsAuth := realtime.NewSessionAuthenticator(kits.sessions)
	hub, err := realtime.NewHub(realtime.Config{
		Adapter:       adapter,
		Authenticator: wsAuth,
		Authorizer:    realtime.NewMembershipAuthorizer(db),
		WorkspaceOf:   realtime.NewProjectWorkspaceResolver(db),
		Logger:        appLogger,
	})
	if err != nil {
		eventsPub.Close()
		return nil, nil, nil, err
	}

	attachPoolSync(db, eventsCfg, eventsPub, hub, appLogger)
	eventsPub.Start(context.Background())

	sched := scheduler.NewScheduler(db, appLogger, 5*time.Minute)
	schedCtx, cancelSched := context.WithCancel(context.Background())
	go sched.Start(schedCtx)

	cleanup := func() {
		cancelSched()
		if closeErr := hub.Close(); closeErr != nil {
			appLogger.Warn("realtime hub close error", slog.Any("error", closeErr))
		}
		eventsPub.Close()
	}
	return hub, eventsPub, cleanup, nil
}

func attachPoolSync(db *gorm.DB, cfg events.Config, pub *events.Publisher, hub *realtime.Hub, appLogger *slog.Logger) {
	if !cfg.Enabled() {
		return
	}
	if cfg.SyncWorkspaceID == 0 {
		appLogger.Warn("ANTENNE_SYNC_WORKSPACE_ID is not set, projects created in Sablier will not be imported")
	}
	syncer := poolsync.New(db, appLogger, cfg.SyncWorkspaceID, hub)
	syncer.Prune(context.Background())
	syncer.Register(pub)
}

func createApiServer(db *gorm.DB, sqlDB sqlPinger, appEnv *env.Config, appLogger *slog.Logger) (*serverBundle, error) {
	kits, err := buildAuth(context.Background(), db, appEnv, appLogger)
	if err != nil {
		return nil, err
	}
	hub, emitter, cleanup, err := initServices(db, kits, appEnv, appLogger)
	if err != nil {
		return nil, err
	}
	workspaces, spaces, stopSpaces := startWorkspaces(db, appEnv, appLogger)
	router := buildRouter(routerDeps{
		db:          db,
		sqlDB:       sqlDB,
		appEnv:      appEnv,
		appLogger:   appLogger,
		kits:        kits,
		realtimeHub: hub,
		emitter:     emitter,
		workspaces:  workspaces,
		directory:   spaces,
	})
	addr := ":" + strconv.Itoa(appEnv.Port)
	srv := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	return &serverBundle{server: srv, cleanup: func() { stopSpaces(); cleanup() }}, nil
}

func main() {
	if healthcheck.Handle(os.Args) {
		return
	}
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func initAppLogger(appEnv *env.Config) (*slog.Logger, func()) {
	var journalClient *journal.Client
	appLogger := logger.New(logger.Config{
		Level: appEnv.LogLevel,
		Wrap: func(handler slog.Handler) slog.Handler {
			if appEnv.JournalURL == "" || appEnv.JournalToken == "" {
				return handler
			}
			journalClient = journal.New(journal.Config{URL: appEnv.JournalURL, Token: appEnv.JournalToken})
			return journal.NewHandler(journalClient, handler)
		},
	})
	cleanup := func() {
		if journalClient != nil {
			journalClient.Close()
		}
	}
	return appLogger, cleanup
}

func initDatabaseAndStorage(appEnv *env.Config, appLogger *slog.Logger) (*gorm.DB, *sql.DB, error) {
	if err := database.EnsureDatabase(appEnv.DatabaseURL); err != nil {
		appLogger.Warn("database bootstrap skipped", slog.Any("error", err))
	}
	db, err := database.Open(appEnv.DatabaseURL)
	if err != nil {
		return nil, nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}
	if err := database.RunMigrations(sqlDB); err != nil {
		return nil, nil, err
	}
	appLogger.Info("database migrations applied")

	if err := os.MkdirAll(filepath.Join(appEnv.UploadDir, "workspaces"), 0o755); err != nil {
		return nil, nil, err
	}
	if err := os.MkdirAll(filepath.Join(appEnv.UploadDir, "assets"), 0o755); err != nil {
		return nil, nil, err
	}
	return db, sqlDB, nil
}

func listenAndServeWithGracefulShutdown(server *http.Server, appLogger *slog.Logger) error {
	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- server.ListenAndServe()
	}()
	appLogger.Info("API server started", slog.String("address", server.Addr))

	shutdownSignal, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErrCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-shutdownSignal.Done():
		appLogger.Info("server shutting down")
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownContext); err != nil {
			return err
		}
		appLogger.Info("server stopped")
	}
	return nil
}

func run() error {
	appEnv, err := env.Load()
	if err != nil {
		logger.New(logger.Config{}).Error("failed to load config", slog.Any("error", err))
		return err
	}
	appLogger, closeLogger := initAppLogger(&appEnv)
	defer closeLogger()

	db, sqlDB, err := initDatabaseAndStorage(&appEnv, appLogger)
	if err != nil {
		appLogger.Error("initialization failed", slog.Any("error", err))
		return err
	}
	defer sqlDB.Close()

	bundle, err := createApiServer(db, sqlDB, &appEnv, appLogger)
	if err != nil {
		appLogger.Error("failed to create server", slog.Any("error", err))
		return err
	}
	defer bundle.cleanup()
	defer bundle.server.Close()
	return listenAndServeWithGracefulShutdown(bundle.server, appLogger)
}
