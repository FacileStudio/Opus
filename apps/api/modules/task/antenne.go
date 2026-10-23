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
	projectFacileIDs := service.projectFacileIDs(ctx, tasks)
	for _, task := range tasks {
		payload := events.NewTaskPayload(task, projectFacileIDs[task.ProjectID], actorEmail)
		service.emitter.EmitTask(ctx, channel, payload)
	}
}

func (service *Service) projectFacileIDs(ctx context.Context, tasks []schemas.Task) map[int64]string {
	ids := make([]int64, 0, len(tasks))
	for _, task := range tasks {
		ids = append(ids, task.ProjectID)
	}
	var projects []schemas.Project
	service.db.WithContext(ctx).Select("id", "facile_id").Where("id IN ?", ids).Find(&projects)
	byID := make(map[int64]string, len(projects))
	for _, project := range projects {
		if project.FacileID != nil {
			byID[project.ID] = *project.FacileID
		}
	}
	return byID
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
