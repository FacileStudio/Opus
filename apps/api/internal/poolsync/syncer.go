// Package poolsync applies Sablier project and task events from the Antenne
// pool to the Opus database.
package poolsync

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"time"

	antenne "github.com/FacileStudio/antenne-client/go"
	enveloppe "github.com/FacileStudio/enveloppe/go"

	"github.com/FacileStudio/Opus/apps/api/internal/events"
	"github.com/FacileStudio/Opus/apps/api/internal/realtime"
	"github.com/FacileStudio/Opus/apps/api/schemas"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const ledgerRetention = 35 * 24 * time.Hour

var (
	errDuplicate = errors.New("poolsync: event already processed")
	errSkipped   = errors.New("poolsync: event skipped")
)

// Listener registers inbound channel handlers on the pool client.
type Listener interface {
	Listen(channel string, handler antenne.EventHandler)
}

// Syncer applies inbound pool events. It writes through the database only, so
// a change it applies is never published back to the pool.
type Syncer struct {
	db          *gorm.DB
	logger      *slog.Logger
	workspaceID int64
	notifier    realtime.Publisher
}

type outcome struct {
	skipped   bool
	msgType   string
	projectID int64
	taskID    int64
}

// New builds a Syncer. workspaceID receives projects created by Sablier; zero
// disables project creation from the pool.
func New(db *gorm.DB, logger *slog.Logger, workspaceID int64, notifier realtime.Publisher) *Syncer {
	return &Syncer{db: db, logger: logger, workspaceID: workspaceID, notifier: notifier}
}

// Register subscribes the syncer to every sync channel.
func (s *Syncer) Register(listener Listener) {
	for _, channel := range events.SyncChannels() {
		listener.Listen(channel, func(payload json.RawMessage, meta antenne.EventMeta) {
			if err := s.Handle(context.Background(), channel, payload); err != nil {
				s.logger.Error("pool sync: apply failed",
					slog.String("channel", channel), slog.Int64("offset", meta.Offset), slog.Any("error", err))
			}
		})
	}
}

// Prune drops ledger rows older than the retention window.
func (s *Syncer) Prune(ctx context.Context) {
	cutoff := time.Now().Add(-ledgerRetention)
	err := s.db.WithContext(ctx).Where("processed_at < ?", cutoff).Delete(&schemas.PoolProcessedEvent{}).Error
	if err != nil {
		s.logger.Warn("pool sync: ledger prune failed", slog.Any("error", err))
	}
}

// Handle applies one inbound event. Replaying an event, receiving one Opus
// published itself, or receiving one it cannot place leaves the database unchanged.
func (s *Syncer) Handle(ctx context.Context, channel string, raw json.RawMessage) error {
	evt, err := decode(raw)
	if err != nil {
		return err
	}
	if evt.app == string(enveloppe.AppOpus) {
		return nil
	}
	if evt.facileID == "" {
		s.logger.Warn("pool sync: event without facile_id, dropped", slog.String("channel", channel))
		return nil
	}
	object, action, _ := strings.Cut(channel, ".")
	var applied outcome
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := claim(tx, evt.key); err != nil {
			return err
		}
		out, err := s.dispatch(tx, object, action, evt)
		if err != nil {
			return err
		}
		if out.skipped {
			return errSkipped
		}
		applied = out
		return nil
	})
	if errors.Is(err, errDuplicate) || errors.Is(err, errSkipped) {
		return nil
	}
	if err != nil {
		return err
	}
	realtime.Notify(ctx, s.notifier, applied.msgType, applied.projectID, applied.taskID)
	return nil
}

func claim(tx *gorm.DB, key string) error {
	if key == "" {
		return nil
	}
	res := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&schemas.PoolProcessedEvent{IdempotencyKey: key})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errDuplicate
	}
	return nil
}

func (s *Syncer) dispatch(tx *gorm.DB, object, action string, evt inbound) (outcome, error) {
	switch object {
	case string(enveloppe.ObjectProject):
		return s.applyProject(tx, action, evt)
	case string(enveloppe.ObjectTask):
		return s.applyTask(tx, action, evt)
	default:
		return outcome{skipped: true}, nil
	}
}
