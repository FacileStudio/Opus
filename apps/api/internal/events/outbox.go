package events

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	antenne "github.com/FacileStudio/antenne-client/go"
	"github.com/FacileStudio/antenne-client/go/catalog"
	envelope "github.com/FacileStudio/antenne-client/go/envelope"
	"github.com/FacileStudio/antenne-client/go/outbox"
)

const outboxMaxBackoff = 5 * time.Second

func (p *Publisher) newDrainer(db *sql.DB) *outbox.Drainer {
	return outbox.NewDrainer(db, p.sendQueued,
		outbox.WithMaxBackoff(outboxMaxBackoff),
		outbox.WithOnError(func(err error) {
			if !errors.Is(err, antenne.ErrNotConnected) {
				p.logger.Warn("antenne: outbox pass failed", slog.Any("error", err))
			}
		}),
	)
}

func (p *Publisher) sendQueued(channel string, payload any) error {
	if !p.transport.IsConnected() {
		return antenne.ErrNotConnected
	}
	raw, ok := payload.(json.RawMessage)
	if !ok {
		return fmt.Errorf("antenne: outbox row on %s is not raw JSON", channel)
	}
	event, err := canonical(channel, raw)
	if err != nil {
		return fmt.Errorf("antenne: decode outbox row on %s: %w", channel, err)
	}
	err = p.transport.EmitNow(channel, event)
	p.recordResult(err)
	return err
}

// canonical gives a queued event back the bytes Opus marshalled. The outbox
// keeps payloads as jsonb, which reorders the keys, so the row is decoded into
// the envelope it came from and encoded again.
func canonical(channel string, raw json.RawMessage) (json.RawMessage, error) {
	object, _, _ := strings.Cut(channel, ".")
	switch envelope.ObjectType(object) {
	case envelope.ObjectTask:
		return reencode[TaskPayload](raw)
	case envelope.ObjectProject:
		return reencode[ProjectPayload](raw)
	case ObjectComment:
		return reencode[CommentPayload](raw)
	case ObjectWorkspace:
		return reencode[WorkspacePayload](raw)
	}
	return raw, nil
}

func reencode[T any](raw json.RawMessage) (json.RawMessage, error) {
	var event envelope.Event[T]
	if err := json.Unmarshal(raw, &event); err != nil {
		return nil, err
	}
	return json.Marshal(event)
}

func (p *Publisher) enqueue(ctx context.Context, channel, key string, event any) {
	ctx = context.WithoutCancel(ctx)
	tx, err := p.sqlDB.BeginTx(ctx, nil)
	if err == nil {
		err = outbox.Enqueue(ctx, tx, catalog.Channel(channel), event, key)
		if err == nil || errors.Is(err, outbox.ErrDuplicate) {
			err = tx.Commit()
		} else {
			_ = tx.Rollback()
		}
	}
	if err != nil {
		p.logger.Error("antenne: failed to enqueue outbox event", slog.String("channel", channel), slog.Any("error", err))
	}
}
