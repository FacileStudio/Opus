package main

import (
	"context"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	documentation "github.com/FacileStudio/Opus/apps/api/internal/documentation"
	"github.com/FacileStudio/Opus/apps/api/internal/env"
	"github.com/FacileStudio/Opus/apps/api/internal/events"
	"github.com/FacileStudio/Opus/apps/api/internal/middleware"
	"github.com/FacileStudio/Opus/apps/api/internal/realtime"
	"github.com/FacileStudio/Opus/apps/api/internal/storage"
	"github.com/FacileStudio/Opus/apps/api/modules/activity"
	"github.com/FacileStudio/Opus/apps/api/modules/asset"
	"github.com/FacileStudio/Opus/apps/api/modules/auth"
	"github.com/FacileStudio/Opus/apps/api/modules/column"
	"github.com/FacileStudio/Opus/apps/api/modules/comment"
	"github.com/FacileStudio/Opus/apps/api/modules/config"
	"github.com/FacileStudio/Opus/apps/api/modules/label"
	"github.com/FacileStudio/Opus/apps/api/modules/notification"
	"github.com/FacileStudio/Opus/apps/api/modules/project"
	"github.com/FacileStudio/Opus/apps/api/modules/search"
	"github.com/FacileStudio/Opus/apps/api/modules/sharelink"
	"github.com/FacileStudio/Opus/apps/api/modules/task"
	"github.com/FacileStudio/Opus/apps/api/modules/taskrelation"
	"github.com/FacileStudio/Opus/apps/api/modules/timeentry"
	"github.com/FacileStudio/Opus/apps/api/modules/workflowrule"
	"github.com/FacileStudio/Opus/apps/api/modules/workspace"
	"github.com/FacileStudio/porte/directory"
	"github.com/FacileStudio/tronc/apiref"
	"github.com/FacileStudio/tronc/health"
	"github.com/FacileStudio/tronc/httpx"
	troncmiddleware "github.com/FacileStudio/tronc/middleware"
	"github.com/FacileStudio/tronc/spa"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type routerDeps struct {
	db          *gorm.DB
	sqlDB       sqlPinger
	appEnv      *env.Config
	appLogger   *slog.Logger
	kits        *authKits
	realtimeHub *realtime.Hub
	emitter     events.Emitter
	reconcile   func(context.Context)
	workspaces  *workspace.Service
	registre    registreFollowers
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
				sharelink.Documentation(),
				timeentry.Documentation(),
				workflowrule.Documentation(),
				notification.Documentation(),
				search.Documentation(),
				asset.Documentation(),
				config.Documentation(),
			},
		},
	}
}

func registerDomainRoutes(api chi.Router, authService middleware.Authenticator, deps routerDeps) {
	db, appEnv, hub := deps.db, deps.appEnv, deps.realtimeHub
	configService := config.NewService(appEnv)
	configService.SetSpacesURL(deps.workspaces.SpacesURL)
	config.RegisterRoutes(api, configService)
	projectService := project.NewService(db)
	projectService.SetEmitter(deps.emitter)
	project.RegisterRoutes(api, projectService, authService)
	column.RegisterRoutes(api, column.NewService(db), authService)
	taskService := task.NewService(db)
	if hub != nil {
		taskService.SetPublisher(hub)
	}
	taskService.SetEmitter(deps.emitter)
	task.RegisterRoutes(api, taskService, authService)
	task.RegisterPublicRoutes(api, taskService)
	relationService := taskrelation.NewService(db)
	labelService := label.NewService(db)
	commentService := comment.NewService(db)
	if hub != nil {
		relationService.SetPublisher(hub)
		labelService.SetPublisher(hub)
		commentService.SetPublisher(hub)
	}
	taskrelation.RegisterRoutes(api, relationService, authService)
	label.RegisterRoutes(api, labelService, authService)
	comment.RegisterRoutes(api, commentService, authService)
	activity.RegisterRoutes(api, activity.NewService(db), authService)
	sharelink.RegisterRoutes(api, sharelink.NewService(db), authService)
	timeentry.RegisterRoutes(api, timeentry.NewService(db), authService)
	workflowrule.RegisterRoutes(api, workflowrule.NewService(db), authService)
	notification.RegisterRoutes(api, notification.NewService(db), authService)
	search.RegisterRoutes(api, search.NewService(db), authService)
	assetStorage := storage.NewAssetStorage(appEnv.UploadDir, storage.LoadS3Config())
	asset.RegisterRoutes(api, asset.NewService(db, assetStorage), authService)
}

func buildRouter(deps routerDeps) chi.Router {
	authService := auth.NewService(deps.db, deps.kits.sessions, deps.appLogger)

	clientDir := spa.DirFromEnv()
	router := httpx.NewRouter(httpx.Config{
		TrustedProxies: deps.appEnv.TrustedProxies,
		CDNProxies:     deps.appEnv.CDNProxies,
		CDNHeader:      deps.appEnv.CDNHeader,
		Logger:         deps.appLogger,
		CORS: troncmiddleware.CORSConfig{
			AllowedOrigins:   deps.appEnv.CORSAllowedOrigins,
			AllowCredentials: true,
			AllowedHeaders: append(slices.Clone(troncmiddleware.DefaultAllowedHeaders),
				"X-Facile-CSRF", "X-Opus-Window-Id"),
		},
		Security: securityConfig(clientDir, storage.LoadS3Config(), deps.appLogger),
	})

	health.Mount(router, deps.sqlDB.PingContext)
	router.Get("/version", versionHandler)
	router.Head("/health", health.Live())
	router.Head("/api/health", health.Live())
	apiref.Mount(router, referenceConfig())

	if deps.realtimeHub != nil {
		router.Get("/ws", deps.realtimeHub.ServeWS)
		router.Get("/ws/{projectId}", deps.realtimeHub.ServeWS)
	}

	router.Route("/api", func(api chi.Router) {
		api.Use(realtime.InitiatorMiddleware)
		health.Mount(api, deps.sqlDB.PingContext)
		api.Head("/health", health.Live())
		deps.kits.sessions.Mount(api)
		deps.kits.oidc.Mount(api)
		auth.RegisterRoutes(api, authService)
		workspace.RegisterRoutes(api, deps.workspaces, authService)
		if deps.registre.spaces != nil && deps.registre.deletions != nil {
			api.Handle(strings.TrimPrefix(directory.SyncPath, "/api"), deps.registre.spaces.Handler())
			api.Method(http.MethodPost, strings.TrimPrefix(directory.DeletionsPath, "/api"), deps.registre.deletions.Handler())
		}
		registerDomainRoutes(api, authService, deps)
		if deps.reconcile != nil {
			mountReconcile(api, deps.db, authService, deps.reconcile)
		}

		if deps.realtimeHub != nil {
			api.Get("/ws", deps.realtimeHub.ServeWS)
			api.Get("/ws/{projectId}", deps.realtimeHub.ServeWS)
		}
	})

	if spa.Available(clientDir) {
		router.Handle("/*", spa.Handler(spa.Config{Dir: clientDir}))
		deps.appLogger.Info("serving client", slog.String("dir", clientDir))
	}

	return router
}
