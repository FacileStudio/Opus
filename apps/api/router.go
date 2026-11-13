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
	"github.com/FacileStudio/Opus/apps/api/modules/github"
	"github.com/FacileStudio/Opus/apps/api/modules/label"
	"github.com/FacileStudio/Opus/apps/api/modules/mcp"
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
	"github.com/FacileStudio/porte/kit"
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
	ideas       *ideas
	domain      domain
	mcp         *mcp.Service
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
				github.Documentation(),
				mcp.Documentation(),
			},
		},
	}
}

// domain holds the services the HTTP routes and the MCP server share, so a
// write through either broadcasts and syncs the same way.
type domain struct {
	projects   *project.Service
	columns    *column.Service
	tasks      *task.Service
	relations  *taskrelation.Service
	labels     *label.Service
	comments   *comment.Service
	activities *activity.Service
}

// newDomain builds the shared services with their realtime publisher and
// Antenne emitter.
func newDomain(db *gorm.DB, hub *realtime.Hub, emitter events.Emitter) domain {
	d := domain{
		projects:   project.NewService(db),
		columns:    column.NewService(db),
		tasks:      task.NewService(db),
		relations:  taskrelation.NewService(db),
		labels:     label.NewService(db),
		comments:   comment.NewService(db),
		activities: activity.NewService(db),
	}
	d.projects.SetEmitter(emitter)
	d.tasks.SetEmitter(emitter)
	if hub != nil {
		d.tasks.SetPublisher(hub)
		d.relations.SetPublisher(hub)
		d.labels.SetPublisher(hub)
		d.comments.SetPublisher(hub)
	}
	return d
}

func registerDomainRoutes(api chi.Router, authService middleware.Authenticator, deps routerDeps) {
	db, appEnv, d := deps.db, deps.appEnv, deps.domain
	configService := config.NewService(appEnv)
	configService.SetSpacesURL(deps.workspaces.SpacesURL)
	config.RegisterRoutes(api, configService)
	project.RegisterRoutes(api, d.projects, authService)
	column.RegisterRoutes(api, d.columns, authService)
	task.RegisterRoutes(api, d.tasks, authService)
	task.RegisterPublicRoutes(api, d.tasks)
	taskrelation.RegisterRoutes(api, d.relations, authService)
	label.RegisterRoutes(api, d.labels, authService)
	comment.RegisterRoutes(api, d.comments, authService)
	activity.RegisterRoutes(api, d.activities, authService)
	sharelink.RegisterRoutes(api, sharelink.NewService(db), authService)
	timeentry.RegisterRoutes(api, timeentry.NewService(db), authService)
	workflowrule.RegisterRoutes(api, workflowrule.NewService(db), authService)
	notification.RegisterRoutes(api, notification.NewService(db), authService)
	search.RegisterRoutes(api, search.NewService(db), authService)
	assetStorage := storage.NewAssetStorage(appEnv.UploadDir, storage.LoadS3Config())
	assetStorage.SetUploadKey(appEnv.UploadKey())
	asset.RegisterRoutes(api, asset.NewService(db, assetStorage), authService)
	github.RegisterRoutes(api, github.NewService(db, appEnv.GitHubWebhookSecret), authService)
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
	if deps.mcp != nil {
		mcp.RegisterRoutes(router, deps.mcp)
	}

	if deps.realtimeHub != nil {
		router.Get("/ws", deps.realtimeHub.ServeWS)
		router.Get("/ws/{projectId}", deps.realtimeHub.ServeWS)
	}

	router.Route("/api", func(api chi.Router) {
		api.Use(realtime.InitiatorMiddleware)
		health.Mount(api, deps.sqlDB.PingContext)
		api.Head("/health", health.Live())
		if err := kit.Mount(api, kit.Config{Sessions: deps.kits.sessions, OIDC: deps.kits.oidc}); err != nil {
			deps.appLogger.Error("auth routes not mounted", slog.Any("error", err))
		}
		auth.RegisterRoutes(api, authService)
		workspace.RegisterRoutes(api, deps.workspaces, authService)
		if deps.registre.spaces != nil && deps.registre.deletions != nil {
			api.Handle(strings.TrimPrefix(directory.SyncPath, "/api"), deps.registre.spaces.Handler())
			api.Method(http.MethodPost, strings.TrimPrefix(directory.DeletionsPath, "/api"), deps.registre.deletions.Handler())
		}
		registerDomainRoutes(api, authService, deps)
		deps.ideas.mount(api, authService)
		if deps.mcp != nil {
			mcp.RegisterPendingRoutes(api, deps.mcp, authService)
		}
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
