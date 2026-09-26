package main

import (
	"log/slog"

	documentation "github.com/FacileStudio/Opus/apps/api/internal/documentation"
	"github.com/FacileStudio/Opus/apps/api/internal/env"
	"github.com/FacileStudio/Opus/apps/api/internal/legacy"
	"github.com/FacileStudio/Opus/apps/api/internal/middleware"
	"github.com/FacileStudio/Opus/apps/api/modules/activity"
	"github.com/FacileStudio/Opus/apps/api/modules/auth"
	"github.com/FacileStudio/Opus/apps/api/modules/column"
	"github.com/FacileStudio/Opus/apps/api/modules/comment"
	"github.com/FacileStudio/Opus/apps/api/modules/label"
	"github.com/FacileStudio/Opus/apps/api/modules/project"
	"github.com/FacileStudio/Opus/apps/api/modules/task"
	"github.com/FacileStudio/Opus/apps/api/modules/taskrelation"
	"github.com/FacileStudio/Opus/apps/api/modules/workspace"
	"github.com/FacileStudio/tronc/apiref"
	"github.com/FacileStudio/tronc/health"
	"github.com/FacileStudio/tronc/httpx"
	troncmiddleware "github.com/FacileStudio/tronc/middleware"
	"github.com/FacileStudio/tronc/spa"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type routerDeps struct {
	db        *gorm.DB
	sqlDB     sqlPinger
	appEnv    *env.Config
	appLogger *slog.Logger
	kits      *authKits
}

func referenceConfig() apiref.Config {
	return apiref.Config{
		Title:       "Opus API",
		Description: "Self-hosted project management for teams.",
		Servers:     []string{"/api"},
		Registry: documentation.Response{
			Modules: []documentation.Module{
				auth.Documentation(),
				workspace.Documentation(),
				workspace.InvitationsDocumentation(),
				project.Documentation(),
				column.Documentation(),
				task.Documentation(),
				taskrelation.Documentation(),
				label.Documentation(),
				comment.Documentation(),
				activity.Documentation(),
			},
		},
	}
}

func registerDomainRoutes(api chi.Router, db *gorm.DB, authService middleware.Authenticator) {
	project.RegisterRoutes(api, project.NewService(db), authService)
	column.RegisterRoutes(api, column.NewService(db), authService)
	task.RegisterRoutes(api, task.NewService(db), authService)
	taskrelation.RegisterRoutes(api, taskrelation.NewService(db), authService)
	label.RegisterRoutes(api, label.NewService(db), authService)
	comment.RegisterRoutes(api, comment.NewService(db), authService)
	activity.RegisterRoutes(api, activity.NewService(db), authService)
}

func buildRouter(deps routerDeps) chi.Router {
	authService := auth.NewService(deps.db, deps.kits.sessions, deps.kits.passwords, deps.appLogger)
	workspaceService := workspace.NewService(deps.db, deps.appEnv.UploadDir)

	router := httpx.NewRouter(httpx.Config{
		TrustedProxies: deps.appEnv.TrustedProxies,
		CDNProxies:     deps.appEnv.CDNProxies,
		CDNHeader:      deps.appEnv.CDNHeader,
		Logger:         deps.appLogger,
		CORS: troncmiddleware.CORSConfig{
			AllowedOrigins:   deps.appEnv.CORSAllowedOrigins,
			AllowCredentials: true,
			AllowedHeaders: []string{
				"Accept", "Authorization", "Content-Type", "X-Facile-CSRF", "X-Opus-Window-Id",
			},
		},
	})

	health.Mount(router, deps.sqlDB.PingContext)
	apiref.Mount(router, referenceConfig())

	router.Route("/api", func(api chi.Router) {
		deps.kits.sessions.Mount(api)
		deps.kits.oidc.Mount(api)
		auth.RegisterRoutes(api, authService, *deps.appEnv)
		workspace.RegisterRoutes(api, workspaceService, authService)
		registerDomainRoutes(api, deps.db, authService)

		if deps.appEnv.LegacyAPI != nil {
			legacy.MountProxy(api, deps.appEnv.LegacyAPI, deps.kits.sessions, authService, deps.appEnv.AuthBridgeSecret)
		}
	})

	clientDir := spa.DirFromEnv()
	if spa.Available(clientDir) {
		router.Handle("/*", spa.Handler(spa.Config{Dir: clientDir}))
		deps.appLogger.Info("serving client", slog.String("dir", clientDir))
	}

	return router
}
