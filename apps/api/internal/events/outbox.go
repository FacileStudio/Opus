package events

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
)

const (
	outboxInterval  = 2 * time.Second
	outboxBatchSize = 100
)

func (p *Publisher) enqueue(ctx context.Context, channel string, event any) {
	raw, err := json.Marshal(event)
	if err != nil {
		p.logger.Error("antenne: failed to serialize outbox event", slog.String("channel", channel), slog.Any("error", err))
		return
	}
	row := schemas.AntenneOutbox{Channel: channel, Payload: string(raw)}
	if err := p.db.WithContext(context.WithoutCancel(ctx)).Create(&row).Error; err != nil {
		p.logger.Error("antenne: failed to enqueue outbox event", slog.String("channel", channel), slog.Any("error", err))
	}
}

func (p *Publisher) runOutbox(ctx context.Context, every time.Duration) {
	ticker := time.NewTicker(every)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.drainOutbox(ctx)
		}
	}
}

// drainOutbox sends queued events in id order while the pool is connected. A
// failure stops the batch, since skipping a row would reorder the log; the
// row is retried on the next tick.
func (p *Publisher) drainOutbox(ctx context.Context) {
	if !p.transport.IsConnected() {
		return
	}
	var rows []schemas.AntenneOutbox
	if err := p.db.WithContext(ctx).Order("id ASC").Limit(outboxBatchSize).Find(&rows).Error; err != nil {
		p.logger.Error("antenne: failed to read outbox", slog.Any("error", err))
		return
	}
	for _, row := range rows {
		err := p.transport.EmitNow(row.Channel, json.RawMessage(row.Payload))
		p.recordResult(err)
		if err != nil {
			p.logger.Warn("antenne: outbox emit failed", slog.String("channel", row.Channel), slog.Any("error", err))
			p.db.WithContext(ctx).Model(&row).Updates(map[string]any{"attempts": row.Attempts + 1, "last_error": err.Error()})
			return
		}
		p.db.WithContext(ctx).Delete(&row)
	}
}
