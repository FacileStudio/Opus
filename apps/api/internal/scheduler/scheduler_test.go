package scheduler

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/FacileStudio/Opus/apps/api/internal/database/dbtest"
	"github.com/FacileStudio/Opus/apps/api/schemas"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := dbtest.Open(t)
	now := time.Now()
	seeds := []any{
		&schemas.User{ID: 42, Email: "dedup@facile.studio", CreatedAt: now, UpdatedAt: now},
		&schemas.User{ID: 101, Email: "actor@facile.studio", CreatedAt: now, UpdatedAt: now},
		&schemas.Workspace{ID: 1, Name: "WS", Slug: "ws", CreatedAt: now},
		&schemas.Project{ID: 1, WorkspaceID: 1, Name: "P", Slug: "p", CreatedAt: now, UpdatedAt: now},
	}
	for _, seed := range seeds {
		if err := db.Create(seed).Error; err != nil {
			t.Fatalf("seed %T: %v", seed, err)
		}
	}
	return db
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestRunPruneInvitations(t *testing.T) {
	db := newTestDB(t)
	s := NewScheduler(db, discardLogger(), 5*time.Minute)
	ctx := context.Background()
	now := time.Now()

	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)

	inv1 := schemas.Invitation{WorkspaceID: 1, InviterID: 42, Email: "exp@facile.studio", Status: "pending", ExpiresAt: &past, CreatedAt: past}
	inv2 := schemas.Invitation{WorkspaceID: 1, InviterID: 42, Email: "valid@facile.studio", Status: "pending", ExpiresAt: &future, CreatedAt: now}
	inv3 := schemas.Invitation{WorkspaceID: 1, InviterID: 42, Email: "noexp@facile.studio", Status: "pending", ExpiresAt: nil, CreatedAt: now}
	inv4 := schemas.Invitation{WorkspaceID: 1, InviterID: 42, Email: "done@facile.studio", Status: "accepted", ExpiresAt: &past, CreatedAt: past}
	db.Create(&[]schemas.Invitation{inv1, inv2, inv3, inv4})

	if err := s.RunPruneInvitations(ctx, now); err != nil {
		t.Fatalf("prune invitations error: %v", err)
	}

	var remaining []schemas.Invitation
	if err := db.Order("id asc").Find(&remaining).Error; err != nil {
		t.Fatalf("query remaining: %v", err)
	}
	if len(remaining) != 3 {
		t.Fatalf("expected 3 remaining invitations, got %d", len(remaining))
	}
	for _, inv := range remaining {
		if inv.Email == "exp@facile.studio" {
			t.Fatalf("expected expired pending invitation to be pruned")
		}
	}
}

func createDueReminderTasks(db *gorm.DB, now time.Time, actorID int64, nonFinalCol, finalCol schemas.Column) {
	t1Due := now.Add(24 * time.Hour)
	t2Due := now.Add(60 * time.Minute)
	t3Due := now.Add(-5 * time.Minute)
	t4Due := now.Add(-5 * time.Minute)
	t5Due := now.Add(24 * time.Hour)
	t6Due := now.Add(5 * time.Hour)

	db.Create(&schemas.Task{WorkspaceID: 1, ProjectID: 1, ID: 1, Title: "Task 1", ActorID: &actorID, ColumnID: &nonFinalCol.ID, DueDate: &t1Due})
	db.Create(&schemas.Task{WorkspaceID: 1, ProjectID: 1, ID: 2, Title: "Task 2", ActorID: &actorID, ColumnID: nil, DueDate: &t2Due})
	db.Create(&schemas.Task{WorkspaceID: 1, ProjectID: 1, ID: 3, Title: "Task 3", ActorID: &actorID, ColumnID: &nonFinalCol.ID, DueDate: &t3Due})
	db.Create(&schemas.Task{WorkspaceID: 1, ProjectID: 1, ID: 4, Title: "Task 4", ActorID: &actorID, ColumnID: &finalCol.ID, DueDate: &t4Due})
	db.Create(&schemas.Task{WorkspaceID: 1, ProjectID: 1, ID: 5, Title: "Task 5", ActorID: nil, ColumnID: &nonFinalCol.ID, DueDate: &t5Due})
	db.Create(&schemas.Task{WorkspaceID: 1, ProjectID: 1, ID: 6, Title: "Task 6", ActorID: &actorID, ColumnID: &nonFinalCol.ID, DueDate: &t6Due})
}

func verifyReminderNotifications(t *testing.T, notifications []schemas.Notification) {
	t.Helper()
	expectedTypes := map[int64]string{
		1: "due_date_reminder",
		2: "due_date_reminder",
		3: "task_overdue",
	}
	for _, n := range notifications {
		if n.ResourceID == nil {
			t.Fatalf("expected resource id on notification")
		}
		var taskID int64
		switch *n.ResourceID {
		case "1":
			taskID = 1
		case "2":
			taskID = 2
		case "3":
			taskID = 3
		}
		if exp, ok := expectedTypes[taskID]; !ok || n.Type != exp {
			t.Fatalf("task %d unexpected notification type %s, want %s", taskID, n.Type, exp)
		}
	}
}

func TestRunDueReminders(t *testing.T) {
	db := newTestDB(t)
	s := NewScheduler(db, discardLogger(), 5*time.Minute)
	ctx := context.Background()
	now := time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)

	actorID := int64(101)
	nonFinalCol := schemas.Column{ProjectID: 1, Name: "In Progress", Slug: "in-prog", IsFinal: false}
	finalCol := schemas.Column{ProjectID: 1, Name: "Done", Slug: "done", IsFinal: true}
	db.Create(&nonFinalCol)
	db.Create(&finalCol)

	createDueReminderTasks(db, now, actorID, nonFinalCol, finalCol)

	if err := s.RunDueReminders(ctx, now); err != nil {
		t.Fatalf("run due reminders failed: %v", err)
	}

	var notifications []schemas.Notification
	db.Order("id asc").Find(&notifications)
	if len(notifications) != 3 {
		t.Fatalf("expected 3 notifications, got %d", len(notifications))
	}

	var sents []schemas.TaskReminderSent
	db.Order("id asc").Find(&sents)
	if len(sents) != 3 {
		t.Fatalf("expected 3 reminder sents, got %d", len(sents))
	}

	verifyReminderNotifications(t, notifications)
}

func TestRunDueRemindersDeduplication(t *testing.T) {
	db := newTestDB(t)
	s := NewScheduler(db, discardLogger(), 5*time.Minute)
	ctx := context.Background()
	now := time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)

	actorID := int64(42)
	due := now.Add(24 * time.Hour)
	db.Create(&schemas.Task{WorkspaceID: 1, ProjectID: 1, ID: 10, Title: "Dedup Task", ActorID: &actorID, DueDate: &due})

	if err := s.RunDueReminders(ctx, now); err != nil {
		t.Fatalf("first run failed: %v", err)
	}
	if err := s.RunDueReminders(ctx, now); err != nil {
		t.Fatalf("second run failed: %v", err)
	}

	var notifCount int64
	db.Model(&schemas.Notification{}).Count(&notifCount)
	if notifCount != 1 {
		t.Fatalf("expected 1 notification after second run, got %d", notifCount)
	}

	var sentCount int64
	db.Model(&schemas.TaskReminderSent{}).Count(&sentCount)
	if sentCount != 1 {
		t.Fatalf("expected 1 sent record, got %d", sentCount)
	}
}

func TestSchedulerStartLifecycle(t *testing.T) {
	db := newTestDB(t)
	s := NewScheduler(db, discardLogger(), 20*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		s.Start(ctx)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("scheduler did not stop after context cancel")
	}
}
