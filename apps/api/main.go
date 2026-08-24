// Command api is the Opus backend, ported from the TypeScript Hono service to
// the Facile Go stack (tronc chassis, porte auth, Journal logging).
package main

import (
	"context"
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
	documentation "github.com/FacileStudio/Opus/apps/api/internal/documentation"
	"github.com/FacileStudio/Opus/apps/api/internal/env"
	"github.com/FacileStudio/Opus/apps/api/internal/legacy"
	"github.com/FacileStudio/Opus/apps/api/modules/auth"
	"github.com/FacileStudio/Opus/apps/api/modules/workspace"
	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/porte/local"
	"github.com/FacileStudio/porte/oidc"
	portepg "github.com/FacileStudio/porte/pg"
	"github.com/FacileStudio/porte/session"
	"github.com/FacileStudio/tronc/apiref"
	"github.com/FacileStudio/tronc/health"
	"github.com/FacileStudio/tronc/healthcheck"
	"github.com/FacileStudio/tronc/httpx"
	"github.com/FacileStudio/tronc/logger"
	troncmiddleware "github.com/FacileStudio/tronc/middleware"
	"github.com/FacileStudio/tronc/spa"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type sqlPinger interface {
	PingContext(ctx context.Context) error
}

// referenceConfig describes the API for the reference page at /docs.
func referenceConfig() apiref.Config {
	return apiref.Config{
		Title:       "Opus API",
		Description: "Self-hosted project management for teams.",
		Servers:     []string{"/api"},
		Registry: documentation.Response{
			Modules: []documentation.Module{
				auth.Documentation,
				workspace.Documentation,
				workspace.InvitationsDocumentation,
			},
		},
	}
}

// buildAuth constructs porte: one session manager, shared by the OIDC kit and
// the local login, over the identity tables.
//
// One manager and not two: they would each keep their own idea of the clock
// and of whether the cookie is Secure, and porte refuses a kit whose config
// disagrees with its manager's for exactly that reason. Discovery runs here,
// so an unreachable or half-configured issuer fails at boot rather than on
// somebody's first login.
func buildAuth(ctx context.Context, db *gorm.DB, appEnv *env.Config, appLogger *slog.Logger) (*session.Manager, *local.Kit, *oidc.Kit, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, nil, err
	}
	store := portepg.New(sqlDB)
	users := auth.NewUserStore(db)
	cfg := appEnv.Porte()

	sessions, err := session.New(cfg, session.Deps{Sessions: store.Sessions(), Logger: appLogger})
	if err != nil {
		return nil, nil, nil, err
	}
	kit, err := oidc.New(ctx, cfg, oidc.Deps{
		Users:      users,
		Identities: store.Identities(),
		Sessions:   sessions,
		Codes:      store.LoginCodes(),
		Logger:     appLogger,
	})
	if err != nil {
		return nil, nil, nil, err
	}
	passwords, err := local.New(local.Config{AllowRegistration: !appEnv.SSOOnly, MinPasswordLength: 12}, local.Deps{
		Users:      users,
		Identities: store.Identities(),
		Sessions:   sessions,
		Logger:     appLogger,
		Count:      users.CountUsers,
	})
	if err != nil {
		return nil, nil, nil, err
	}
	return sessions, passwords, kit, nil
}

// buildRouter assembles the chi router with the suite middleware chain, health
// probes, the API reference, the native modules, the legacy proxy for
// unported paths and the SPA catch-all.
func buildRouter(db *gorm.DB, sqlDB sqlPinger, appEnv *env.Config, appLogger *slog.Logger, sessions *session.Manager, passwords *local.Kit, kit *oidc.Kit) chi.Router {
	authService := auth.NewService(db, sessions, passwords, appLogger)
	workspaceService := workspace.NewService(db, appEnv.UploadDir)

	router := httpx.NewRouter(httpx.Config{
		TrustedProxies: appEnv.TrustedProxies,
		CDNProxies:     appEnv.CDNProxies,
		CDNHeader:      appEnv.CDNHeader,
		Logger:         appLogger,
		CORS: troncmiddleware.CORSConfig{
			AllowedOrigins: appEnv.CORSAllowedOrigins,
			// The web app fetches with credentials: include (the HttpOnly
			// session cookie is the transport), so every cross-origin response
			// must carry Access-Control-Allow-Credentials or the browser drops
			// it and the fetch rejects. X-Opus-Window-Id rides on every hono
			// client request, so the preflight must allow it.
			AllowCredentials: true,
			AllowedHeaders: []string{
				"Accept", "Authorization", "Content-Type", "X-Facile-CSRF", "X-Opus-Window-Id",
			},
		},
	})

	health.Mount(router, sqlDB.PingContext)
	apiref.Mount(router, referenceConfig())

	router.Route("/api", func(api chi.Router) {
		sessions.Mount(api)
		kit.Mount(api)
		auth.RegisterRoutes(api, authService, *appEnv)
		workspace.RegisterRoutes(api, workspaceService, authService)

		if appEnv.LegacyAPI != nil {
			legacy.MountProxy(api, appEnv.LegacyAPI, sessions, authService, appEnv.AuthBridgeSecret)
		}
	})

	clientDir := spa.DirFromEnv()
	if spa.Available(clientDir) {
		router.Handle("/*", spa.Handler(spa.Config{Dir: clientDir}))
		appLogger.Info("serving client", slog.String("dir", clientDir))
	}

	return router
}

// createApiServer builds the HTTP server the process listens on.
func createApiServer(db *gorm.DB, sqlDB sqlPinger, appEnv *env.Config, appLogger *slog.Logger) (*http.Server, error) {
	sessions, passwords, kit, err := buildAuth(context.Background(), db, appEnv, appLogger)
	if err != nil {
		return nil, err
	}
	router := buildRouter(db, sqlDB, appEnv, appLogger, sessions, passwords, kit)
	addr := ":" + strconv.Itoa(appEnv.Port)
	return &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}, nil
}

// main starts the API and shuts down cleanly on SIGINT and SIGTERM.
func main() {
	if healthcheck.Handle(os.Args) {
		return
	}
	if err := run(); err != nil {
		os.Exit(1)
	}
}

// run wires the environment, the logger, the database and the server, then
// blocks until the process is told to stop.
func run() error {
	appEnv, err := env.Load()
	appLogger := logger.New(logger.Config{})
	if err != nil {
		appLogger.Error("failed to load config", slog.Any("error", err))
		return err
	}

	var journalClient *journal.Client
	appLogger = logger.New(logger.Config{
		Level: appEnv.LogLevel,
		Wrap: func(handler slog.Handler) slog.Handler {
			if appEnv.JournalURL == "" || appEnv.JournalToken == "" {
				return handler
			}
			journalClient = journal.New(journal.Config{URL: appEnv.JournalURL, Token: appEnv.JournalToken})
			return journal.NewHandler(journalClient, handler)
		},
	})
	if journalClient != nil {
		defer journalClient.Close()
	}

	db, err := database.Open(appEnv.DatabaseURL)
	if err != nil {
		appLogger.Error("failed to open database", slog.Any("error", err))
		return err
	}
	if err := schemas.Migrate(db); err != nil {
		appLogger.Error("failed to run migrations", slog.Any("error", err))
		return err
	}
	appLogger.Info("database migrations applied")

	if err := os.MkdirAll(appEnv.UploadDir, 0o755); err != nil {
		appLogger.Error("failed to prepare upload directory", slog.Any("error", err))
		return err
	}
	if err := os.MkdirAll(filepath.Join(appEnv.UploadDir, "workspaces"), 0o755); err != nil {
		appLogger.Error("failed to prepare upload subdirectories", slog.Any("error", err))
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		appLogger.Error("failed to access database handle", slog.Any("error", err))
		return err
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			appLogger.Error("failed to close database", slog.Any("error", err))
		}
	}()

	server, err := createApiServer(db, sqlDB, &appEnv, appLogger)
	if err != nil {
		appLogger.Error("failed to create server", slog.Any("error", err))
		return err
	}

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
			appLogger.Error("server stopped", slog.Any("error", err))
			return err
		}
	case <-shutdownSignal.Done():
		appLogger.Info("server shutting down")
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownContext); err != nil {
			appLogger.Error("server shutdown failed", slog.Any("error", err))
			return err
		}
		appLogger.Info("server stopped")
	}

	return nil
}
