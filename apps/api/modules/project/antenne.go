package project

import (
	"context"

	"github.com/FacileStudio/Opus/apps/api/internal/events"
	"github.com/FacileStudio/Opus/apps/api/schemas"
)

// SetEmitter enables publishing project events to the Antenne bus.
func (s *Service) SetEmitter(emitter events.Emitter) {
	s.emitter = emitter
}

func (s *Service) emit(ctx context.Context, userID int64, channel string, project schemas.Project) {
	if s.emitter == nil {
		return
	}
	var actorEmail string
	s.orm.WithContext(ctx).Model(&schemas.User{}).Where("id = ?", userID).Limit(1).Pluck("email", &actorEmail)
	s.emitter.EmitProject(ctx, channel, events.NewProjectPayload(project, actorEmail))
}

func (s *Service) loadTasksForEvents(ctx context.Context, projectID int64) []schemas.Task {
	if s.emitter == nil {
		return nil
	}
	var tasks []schemas.Task
	s.orm.WithContext(ctx).Where("project_id = ?", projectID).Find(&tasks)
	return tasks
}

func (s *Service) emitTasksDeleted(ctx context.Context, project schemas.Project, tasks []schemas.Task) {
	if s.emitter == nil || len(tasks) == 0 {
		return
	}
	projectFacileID := events.NewProjectPayload(project, "").FacileID
	actorEmails := events.TaskActorEmails(ctx, s.orm, tasks)
	for _, task := range tasks {
		s.emitter.EmitTask(ctx, events.ChannelTaskDeleted, events.NewTaskPayload(task, projectFacileID, actorEmails[task.ID]))
	}
}
