package events

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	enveloppe "github.com/FacileStudio/enveloppe/go"
)

// IdempotencyKey builds an idempotency key from an entity ID and timestamp.
func IdempotencyKey(entityID string, t time.Time) string {
	return fmt.Sprintf("%s:%d", entityID, t.UTC().UnixMilli())
}

// ActionForChannel maps an event channel name to its canonical enveloppe action.
func ActionForChannel(channel string) enveloppe.Action {
	switch channel {
	case ChannelTaskCreated, ChannelProjectCreated, ChannelCommentCreated, ChannelWorkspaceCreated:
		return enveloppe.ActionCreated
	case ChannelTaskDeleted, ChannelProjectDeleted:
		return enveloppe.ActionDeleted
	default:
		return enveloppe.ActionUpdated
	}
}

// BuildEnvelope creates a populated enveloppe.Event with the Opus contract.
func BuildEnvelope[T any](channel string, entityID string, objType enveloppe.ObjectType, payload T, now time.Time) enveloppe.Event[T] {
	return enveloppe.Event[T]{
		Version:        enveloppe.EventVersion,
		App:            enveloppe.AppOpus,
		Object:         objType,
		Action:         ActionForChannel(channel),
		FacileID:       entityID,
		Payload:        payload,
		Timestamp:      now.UTC().Format(time.RFC3339),
		IdempotencyKey: IdempotencyKey(entityID, now),
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

func (p *Publisher) emit(channel string, entityID string, objType enveloppe.ObjectType, payload any) error {
	if p == nil || !p.config.Enabled() || p.transport == nil {
		if p != nil && p.logger != nil {
			p.logger.Debug("antenne disabled, skipping event", slog.String("channel", channel))
		}
		return nil
	}
	now := time.Now().UTC()
	event := BuildEnvelope(channel, entityID, objType, payload, now)
	err := p.transport.EmitNow(channel, event)
	p.recordResult(err)
	if err != nil {
		p.logger.Warn("antenne emit failed", slog.String("channel", channel), slog.Any("error", err))
	}
	return nil
}

// EmitTask publishes a task event to the given channel.
func (p *Publisher) EmitTask(_ context.Context, channel string, payload TaskPayload) error {
	return p.emit(channel, payload.entityID(), enveloppe.ObjectTask, payload)
}

// EmitProject publishes a project event to the given channel.
func (p *Publisher) EmitProject(_ context.Context, channel string, payload ProjectPayload) error {
	return p.emit(channel, payload.entityID(), enveloppe.ObjectProject, payload)
}

// EmitComment publishes a comment event to the given channel.
func (p *Publisher) EmitComment(_ context.Context, channel string, payload CommentPayload) error {
	return p.emit(channel, payload.entityID(), ObjectComment, payload)
}

// EmitWorkspace publishes a workspace event to the given channel.
func (p *Publisher) EmitWorkspace(_ context.Context, channel string, payload WorkspacePayload) error {
	return p.emit(channel, payload.entityID(), ObjectWorkspace, payload)
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
