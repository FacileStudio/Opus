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
	envelope "github.com/FacileStudio/antenne-client/go/envelope"

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
	issuer      string
	notifier    realtime.Publisher
	adopted     func(ctx context.Context, projectID int64)
}

type outcome struct {
	skipped   bool
	adopted   bool
	msgType   string
	projectID int64
	taskID    int64
}

// New builds a Syncer. workspaceID receives projects from old Sablier
// payloads that carry no space; zero skips them.
func New(db *gorm.DB, logger *slog.Logger, workspaceID int64, notifier realtime.Publisher) *Syncer {
	return &Syncer{db: db, logger: logger, workspaceID: workspaceID, notifier: notifier}
}

// OnAdopted runs fn after a project took over the facile_id of its Sablier
// twin, so its tasks can be published again under the new project id.
func (s *Syncer) OnAdopted(fn func(ctx context.Context, projectID int64)) {
	s.adopted = fn
}

// SetIssuer names the OIDC issuer whose porte identities resolve owner and
// user subjects. Empty matches by email only.
func (s *Syncer) SetIssuer(issuer string) {
	s.issuer = issuer
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
// published itself, receiving one from another app, receiving one older than
// the ledger window, or receiving one it cannot place leaves the database unchanged.
func (s *Syncer) Handle(ctx context.Context, channel string, raw json.RawMessage) error {
	evt, err := decode(raw)
	if err != nil {
		return err
	}
	if evt.app != string(envelope.AppSablier) {
		return nil
	}
	if evt.at.IsZero() || time.Since(evt.at) > ledgerRetention {
		s.logger.Warn("pool sync: stale or undated event dropped", slog.String("channel", channel))
		return nil
	}
	if evt.facileID == "" || isLegacyID(evt.facileID) {
		s.logger.Warn("pool sync: event without a usable facile_id, dropped", slog.String("channel", channel))
		return nil
	}
	object, action, _ := strings.Cut(channel, ".")
	applied, err := s.applyOnce(ctx, object, action, evt)
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		applied, err = s.applyOnce(ctx, object, action, evt)
	}
	if errors.Is(err, errDuplicate) || errors.Is(err, errSkipped) {
		return nil
	}
	if err != nil {
		return err
	}
	realtime.Notify(ctx, s.notifier, applied.msgType, applied.projectID, applied.taskID)
	if applied.adopted && s.adopted != nil {
		s.adopted(ctx, applied.projectID)
	}
	return nil
}

func (s *Syncer) applyOnce(ctx context.Context, object, action string, evt inbound) (outcome, error) {
	var applied outcome
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
	return applied, err
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
	case string(envelope.ObjectProject):
		return s.applyProject(tx, action, evt)
	case string(envelope.ObjectTask):
		return s.applyTask(tx, action, evt)
	case string(envelope.ObjectTimeEntry):
		return s.applyTimeEntry(tx, action, evt)
	default:
		return outcome{skipped: true}, nil
	}
}
