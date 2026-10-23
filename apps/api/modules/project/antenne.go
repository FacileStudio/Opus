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
