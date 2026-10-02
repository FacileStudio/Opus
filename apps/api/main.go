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
	"github.com/FacileStudio/Opus/apps/api/internal/realtime"
	"github.com/FacileStudio/Opus/apps/api/internal/scheduler"
	"github.com/FacileStudio/Opus/apps/api/modules/auth"
	"github.com/FacileStudio/Opus/apps/api/schemas"
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

type serverBundle struct {
	server  *http.Server
	cleanup func()
}

func initServices(db *gorm.DB, kits *authKits, appEnv *env.Config, appLogger *slog.Logger) (*realtime.Hub, func(), error) {
	eventsPub := events.New(events.LoadConfig(), events.WithLogger(appLogger))
	eventsPub.Start(context.Background())

	adapter := realtime.NewRedisAdapter(appEnv.RedisURL, appLogger)
	wsAuth := realtime.NewSessionAuthenticator(kits.sessions)
	hub, err := realtime.NewHub(realtime.Config{
		Adapter:       adapter,
		Authenticator: wsAuth,
		Logger:        appLogger,
	})
	if err != nil {
		eventsPub.Close()
		return nil, nil, err
	}

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
	return hub, cleanup, nil
}

func createApiServer(db *gorm.DB, sqlDB sqlPinger, appEnv *env.Config, appLogger *slog.Logger) (*serverBundle, error) {
	kits, err := buildAuth(context.Background(), db, appEnv, appLogger)
	if err != nil {
		return nil, err
	}
	hub, cleanup, err := initServices(db, kits, appEnv, appLogger)
	if err != nil {
		return nil, err
	}
	router := buildRouter(routerDeps{
		db:          db,
		sqlDB:       sqlDB,
		appEnv:      appEnv,
		appLogger:   appLogger,
		kits:        kits,
		realtimeHub: hub,
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
	return &serverBundle{server: srv, cleanup: cleanup}, nil
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
		appLogger.Warn("goose migrations failed, using GORM schema migration", slog.Any("error", err))
		if err := schemas.Migrate(db); err != nil {
			return nil, nil, err
		}
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
