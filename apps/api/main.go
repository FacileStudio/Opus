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
	"strconv"
	"syscall"
	"time"

	"github.com/FacileStudio/Journal/sdk/journal"
	"github.com/FacileStudio/Opus/apps/api/internal/env"
	"github.com/go-chi/chi/v5"

	documentation "github.com/FacileStudio/Opus/apps/api/internal/documentation"
	"github.com/FacileStudio/Opus/apps/api/internal/legacy"
	"github.com/FacileStudio/tronc/apiref"
	"github.com/FacileStudio/tronc/health"
	"github.com/FacileStudio/tronc/healthcheck"
	"github.com/FacileStudio/tronc/httpx"
	"github.com/FacileStudio/tronc/logger"
	troncmiddleware "github.com/FacileStudio/tronc/middleware"
	"github.com/FacileStudio/tronc/spa"
)

// referenceConfig describes the API for the reference page at /docs.
func referenceConfig() apiref.Config {
	return apiref.Config{
		Title:       "Opus API",
		Description: "Self-hosted project management for teams.",
		Servers:     []string{"/api"},
		Registry: documentation.Response{
			Modules: []documentation.Module{},
		},
	}
}

// buildRouter assembles the chi router with the suite middleware chain, health
// probes, the API reference and the SPA catch-all.
func buildRouter(appEnv *env.Config, appLogger *slog.Logger) chi.Router {
	router := httpx.NewRouter(httpx.Config{
		TrustedProxies: appEnv.TrustedProxies,
		CDNProxies:     appEnv.CDNProxies,
		CDNHeader:      appEnv.CDNHeader,
		Logger:         appLogger,
		CORS: troncmiddleware.CORSConfig{
			AllowedOrigins: appEnv.CORSAllowedOrigins,
		},
	})

	health.Mount(router)
	apiref.Mount(router, referenceConfig())

	router.Route("/api", func(api chi.Router) {
		if appEnv.LegacyAPI != nil {
			legacy.MountProxy(api, appEnv.LegacyAPI)
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
func createApiServer(appEnv *env.Config, appLogger *slog.Logger) (*http.Server, error) {
	router := buildRouter(appEnv, appLogger)
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

// run wires the environment, the logger and the server, then blocks until the
// process is told to stop.
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

	server, err := createApiServer(&appEnv, appLogger)
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
