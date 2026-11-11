package events

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	antenne "github.com/FacileStudio/antenne-client/go"
	"github.com/FacileStudio/antenne-client/go/catalog"
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
	err := p.transport.EmitNow(channel, payload)
	p.recordResult(err)
	return err
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
