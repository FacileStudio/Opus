package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/FacileStudio/Opus/apps/api/internal/authcontext"
	"github.com/FacileStudio/Opus/apps/api/internal/events"
	"github.com/FacileStudio/Opus/apps/api/internal/middleware"
	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

// reconcileJob is the Sablier backfill: it queues every syncable project and
// task for the pool. Failures are logged; the next run starts over.
func reconcileJob(publisher *events.Publisher, appLogger *slog.Logger) func(context.Context) {
	return func(ctx context.Context) {
		if err := publisher.Reconcile(ctx); err != nil {
			appLogger.Warn("sablier reconcile failed", slog.Any("error", err))
		}
	}
}

// mountReconcile adds POST /sync/reconcile, which re-runs the Sablier
// backfill in the background and answers 202. Only an owner or admin of a
// workspace may call it.
func mountReconcile(api chi.Router, db *gorm.DB, authService middleware.Authenticator, job func(context.Context)) {
	api.With(middleware.RequireAuth(authService)).Post("/sync/reconcile", func(w http.ResponseWriter, request *http.Request) {
		identity, _ := authcontext.IdentityFromContext(request.Context())
		var count int64
		err := db.WithContext(request.Context()).Model(&schemas.WorkspaceMember{}).
			Where("user_id = ? AND role IN ?", identity.UserID, []string{"owner", "admin"}).Count(&count).Error
		if err != nil {
			httpjson.WriteError(w, errors.Internal("failed to check the caller's role", err))
			return
		}
		if count == 0 {
			httpjson.WriteError(w, errors.Forbidden("Only a workspace owner or admin can reconcile with Sablier"))
			return
		}
		go job(context.WithoutCancel(request.Context()))
		httpjson.WriteJSON(w, http.StatusAccepted, map[string]string{"status": "queued"})
	})
}
