package realtime

import (
	"context"
	"net/http"
	"strconv"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"gorm.io/gorm"
)

// Publisher dispatches a broadcast message to every client on a project channel.
type Publisher interface {
	BroadcastToProject(projectID string, msg BroadcastMessage, excludeInitiatorID string) error
}

type initiatorKey struct{}

// WithInitiator stores the id of the browser window that caused a change.
func WithInitiator(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, initiatorKey{}, id)
}

// InitiatorFromContext returns the window id stored by WithInitiator.
func InitiatorFromContext(ctx context.Context) string {
	id, _ := ctx.Value(initiatorKey{}).(string)
	return id
}

// InitiatorMiddleware records the request's window id so broadcasts skip the window that made the change.
func InitiatorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id := extractInitiatorID(r); id != "" {
			r = r.WithContext(WithInitiator(r.Context(), id))
		}
		next.ServeHTTP(w, r)
	})
}

// Notify broadcasts a change on a project and skips the window that caused it.
func Notify(ctx context.Context, publisher Publisher, msgType string, projectID, taskID int64) {
	if publisher == nil || projectID <= 0 {
		return
	}
	msg := BroadcastMessage{Type: msgType}
	if taskID > 0 {
		msg.TaskID = strconv.FormatInt(taskID, 10)
	}
	_ = publisher.BroadcastToProject(strconv.FormatInt(projectID, 10), msg, InitiatorFromContext(ctx))
}

// NotifyTasks broadcasts msg once on the project of each given task.
func NotifyTasks(ctx context.Context, db *gorm.DB, publisher Publisher, msg BroadcastMessage, taskIDs ...int64) {
	if publisher == nil || len(taskIDs) == 0 {
		return
	}
	var projectIDs []int64
	db.WithContext(ctx).Model(&schemas.Task{}).Where("id IN ?", taskIDs).Distinct().Pluck("project_id", &projectIDs)
	for _, projectID := range projectIDs {
		_ = publisher.BroadcastToProject(strconv.FormatInt(projectID, 10), msg, InitiatorFromContext(ctx))
	}
}
