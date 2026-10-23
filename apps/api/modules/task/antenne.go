package task

import (
	"context"

	"github.com/FacileStudio/Opus/apps/api/internal/events"
	"github.com/FacileStudio/Opus/apps/api/schemas"
)

// SetEmitter enables publishing task events to the Antenne bus.
func (service *Service) SetEmitter(emitter events.Emitter) {
	service.emitter = emitter
}

func (service *Service) emit(ctx context.Context, userID int64, channel string, tasks ...schemas.Task) {
	if service.emitter == nil || len(tasks) == 0 {
		return
	}
	var actorEmail string
	service.db.WithContext(ctx).Model(&schemas.User{}).Where("id = ?", userID).Limit(1).Pluck("email", &actorEmail)
	for _, task := range tasks {
		service.emitter.EmitTask(ctx, channel, events.NewTaskPayload(task, actorEmail))
	}
}

func (service *Service) emitReloaded(ctx context.Context, userID int64, channel string, taskIDs []int64) {
	if service.emitter == nil {
		return
	}
	var tasks []schemas.Task
	if err := service.db.WithContext(ctx).Where("id IN ?", taskIDs).Find(&tasks).Error; err != nil {
		return
	}
	service.emit(ctx, userID, channel, tasks...)
}
