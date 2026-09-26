package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"gorm.io/gorm"
)

// Scheduler manages periodic background maintenance jobs.
type Scheduler struct {
	db       *gorm.DB
	logger   *slog.Logger
	interval time.Duration
}

// NewScheduler creates an initialized maintenance scheduler.
func NewScheduler(db *gorm.DB, logger *slog.Logger, interval time.Duration) *Scheduler {
	if interval <= 0 {
		interval = 5 * time.Minute
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Scheduler{
		db:       db,
		logger:   logger,
		interval: interval,
	}
}

// Start executes maintenance jobs immediately and on ticker ticks until ctx is cancelled.
func (s *Scheduler) Start(ctx context.Context) {
	s.runAll(ctx, time.Now())
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			s.runAll(ctx, t)
		}
	}
}

func (s *Scheduler) runAll(ctx context.Context, now time.Time) {
	if ctx.Err() != nil {
		return
	}
	if err := s.RunPruneInvitations(ctx, now); err != nil {
		s.logger.Error("prune invitations job failed", slog.Any("error", err))
	}
	if ctx.Err() != nil {
		return
	}
	if err := s.RunDueReminders(ctx, now); err != nil {
		s.logger.Error("due reminders job failed", slog.Any("error", err))
	}
}

// RunPruneInvitations deletes expired pending workspace invitations.
func (s *Scheduler) RunPruneInvitations(ctx context.Context, now time.Time) error {
	if now.IsZero() {
		now = time.Now()
	}
	return s.db.WithContext(ctx).
		Where("status = ? AND expires_at IS NOT NULL AND expires_at < ?", "pending", now).
		Delete(&schemas.Invitation{}).Error
}
