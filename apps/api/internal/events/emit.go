package events

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	envelope "github.com/FacileStudio/antenne-client/go/envelope"

	"github.com/FacileStudio/Opus/apps/api/schemas"
)

// IdempotencyKey builds the contract key <object>.<action>:<facile_id>:<unixnano>,
// where the channel already is <object>.<action>.
func IdempotencyKey(channel, entityID string, t time.Time) string {
	return fmt.Sprintf("%s:%s:%d", channel, entityID, t.UTC().UnixNano())
}

// ActionForChannel maps an event channel name to its canonical envelope action.
func ActionForChannel(channel string) envelope.Action {
	switch channel {
	case ChannelTaskCreated, ChannelProjectCreated, ChannelCommentCreated, ChannelWorkspaceCreated:
		return envelope.ActionCreated
	case ChannelTaskDeleted, ChannelProjectDeleted:
		return envelope.ActionDeleted
	default:
		return envelope.ActionUpdated
	}
}

// BuildEnvelope creates a populated envelope.Event with the Opus contract.
func BuildEnvelope[T any](channel string, entityID string, objType envelope.ObjectType, payload T, now time.Time) envelope.Event[T] {
	return envelope.Event[T]{
		Version:        envelope.EventVersion,
		App:            envelope.AppOpus,
		Object:         objType,
		Action:         ActionForChannel(channel),
		FacileID:       entityID,
		Payload:        payload,
		Timestamp:      now.UTC().Format(time.RFC3339),
		IdempotencyKey: IdempotencyKey(channel, entityID, now),
	}
}

func (p *Publisher) recordResult(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err == nil {
		p.published++
		return
	}
	p.dropped++
	p.lastError = err.Error()
}

func (p *Publisher) active() bool {
	return p != nil && p.config.Enabled() && p.transport != nil
}

func (p *Publisher) emit(ctx context.Context, channel string, entityID string, objType envelope.ObjectType, payload any) error {
	if !p.active() {
		if p != nil && p.logger != nil {
			p.logger.Debug("antenne disabled, skipping event", slog.String("channel", channel))
		}
		return nil
	}
	event := BuildEnvelope(channel, entityID, objType, payload, time.Now().UTC())
	if p.drainer != nil {
		p.enqueue(ctx, channel, event.IdempotencyKey, event)
		return nil
	}
	err := p.transport.EmitNow(channel, event)
	p.recordResult(err)
	if err != nil {
		p.logger.Warn("antenne emit failed", slog.String("channel", channel), slog.Any("error", err))
	}
	return nil
}

// EmitTask publishes a task event to the given channel. A task whose
// workspace Sablier cannot place is not published. A published task counts as
// linked to Sablier, so it is never merged with a twin.
func (p *Publisher) EmitTask(ctx context.Context, channel string, payload TaskPayload) error {
	return p.emitTask(ctx, channel, payload, true)
}

func (p *Publisher) emitTask(ctx context.Context, channel string, payload TaskPayload, link bool) error {
	if !p.active() || !p.prepareTask(ctx, &payload) {
		return nil
	}
	if link {
		p.link(ctx, &schemas.Task{}, payload.ID)
	}
	return p.emit(ctx, channel, payload.entityID(), envelope.ObjectTask, payload)
}

// EmitProject publishes a project event to the given channel. A project in
// a workspace that is neither linked to Registre nor personal is not
// published. A published project counts as linked to Sablier.
func (p *Publisher) EmitProject(ctx context.Context, channel string, payload ProjectPayload) error {
	return p.emitProject(ctx, channel, payload, true)
}

func (p *Publisher) emitProject(ctx context.Context, channel string, payload ProjectPayload, link bool) error {
	if !p.active() || !p.prepareProject(ctx, &payload) {
		return nil
	}
	if link {
		p.link(ctx, &schemas.Project{}, payload.ID)
	}
	return p.emit(ctx, channel, payload.entityID(), envelope.ObjectProject, payload)
}

// EmitComment publishes a comment event to the given channel.
func (p *Publisher) EmitComment(ctx context.Context, channel string, payload CommentPayload) error {
	return p.emit(ctx, channel, payload.entityID(), ObjectComment, payload)
}

// EmitWorkspace publishes a workspace event to the given channel.
func (p *Publisher) EmitWorkspace(ctx context.Context, channel string, payload WorkspacePayload) error {
	return p.emit(ctx, channel, payload.entityID(), ObjectWorkspace, payload)
}

// EmitTaskCreated emits a task.created event.
func (p *Publisher) EmitTaskCreated(ctx context.Context, payload TaskPayload) error {
	return p.EmitTask(ctx, ChannelTaskCreated, payload)
}

// EmitTaskUpdated emits a task.updated event.
func (p *Publisher) EmitTaskUpdated(ctx context.Context, payload TaskPayload) error {
	return p.EmitTask(ctx, ChannelTaskUpdated, payload)
}

// EmitTaskDeleted emits a task.deleted event.
func (p *Publisher) EmitTaskDeleted(ctx context.Context, payload TaskPayload) error {
	return p.EmitTask(ctx, ChannelTaskDeleted, payload)
}

// EmitTaskStatusChanged emits a task.status_changed event.
func (p *Publisher) EmitTaskStatusChanged(ctx context.Context, payload TaskPayload) error {
	return p.EmitTask(ctx, ChannelTaskStatusChanged, payload)
}

// EmitTaskActorChanged emits a task.actor_changed event.
func (p *Publisher) EmitTaskActorChanged(ctx context.Context, payload TaskPayload) error {
	return p.EmitTask(ctx, ChannelTaskActorChanged, payload)
}

// EmitProjectCreated emits a project.created event.
func (p *Publisher) EmitProjectCreated(ctx context.Context, payload ProjectPayload) error {
	return p.EmitProject(ctx, ChannelProjectCreated, payload)
}

// EmitProjectUpdated emits a project.updated event.
func (p *Publisher) EmitProjectUpdated(ctx context.Context, payload ProjectPayload) error {
	return p.EmitProject(ctx, ChannelProjectUpdated, payload)
}

// EmitProjectDeleted emits a project.deleted event.
func (p *Publisher) EmitProjectDeleted(ctx context.Context, payload ProjectPayload) error {
	return p.EmitProject(ctx, ChannelProjectDeleted, payload)
}

// EmitCommentCreated emits a comment.created event.
func (p *Publisher) EmitCommentCreated(ctx context.Context, payload CommentPayload) error {
	return p.EmitComment(ctx, ChannelCommentCreated, payload)
}

// EmitWorkspaceCreated emits a workspace.created event.
func (p *Publisher) EmitWorkspaceCreated(ctx context.Context, payload WorkspacePayload) error {
	return p.EmitWorkspace(ctx, ChannelWorkspaceCreated, payload)
}
