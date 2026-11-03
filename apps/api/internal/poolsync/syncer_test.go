package poolsync

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/FacileStudio/Opus/apps/api/internal/realtime"
	"github.com/FacileStudio/Opus/apps/api/schemas"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type notifyRecorder struct {
	types []string
}

func (n *notifyRecorder) BroadcastToProject(_ string, msg realtime.BroadcastMessage, _ string) error {
	n.types = append(n.types, msg.Type)
	return nil
}

func newSyncer(t *testing.T) (*Syncer, *gorm.DB, *notifyRecorder) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(&schemas.User{}, &schemas.Workspace{}, &schemas.WorkspaceMember{}, &schemas.Project{}, &schemas.Column{},
		&schemas.Task{}, &schemas.TimeEntry{}, &schemas.PoolProcessedEvent{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := db.Exec("CREATE TABLE porte_identities (user_id INTEGER, provider TEXT, subject TEXT)").Error; err != nil {
		t.Fatalf("identities: %v", err)
	}
	workspace := schemas.Workspace{Name: "WS", Slug: "ws", CreatedAt: time.Now()}
	if err := db.Create(&workspace).Error; err != nil {
		t.Fatalf("seed workspace: %v", err)
	}
	if err := db.Create(&schemas.User{Email: "Alice@Facile.studio", Name: "Alice"}).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	notifier := &notifyRecorder{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return New(db, logger, workspace.ID, notifier), db, notifier
}

func wire(t *testing.T, app, key string, payload map[string]any) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(map[string]any{
		"version": 1, "app": app, "idempotency_key": key, "facile_id": payload["facile_id"], "payload": payload,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return raw
}

func send(t *testing.T, s *Syncer, channel string, raw json.RawMessage) {
	t.Helper()
	if err := s.Handle(context.Background(), channel, raw); err != nil {
		t.Fatalf("handle %s: %v", channel, err)
	}
}

func count(t *testing.T, db *gorm.DB, model any) int64 {
	t.Helper()
	var n int64
	if err := db.Model(model).Count(&n).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	return n
}

func seedSablierProject(t *testing.T, s *Syncer) {
	t.Helper()
	send(t, s, "project.created", wire(t, "Sablier", "p1", map[string]any{
		"facile_id": "proj-1", "name": "Website", "description": "Redesign", "icon": "Magnifer",
	}))
}

func TestProjectLifecycleFromSablier(t *testing.T) {
	s, db, _ := newSyncer(t)
	seedSablierProject(t, s)

	var created schemas.Project
	if err := db.Where("facile_id = ?", "proj-1").First(&created).Error; err != nil {
		t.Fatalf("project not created: %v", err)
	}
	if created.Name != "Website" || created.Icon != "Magnifier" || created.WorkspaceID != s.workspaceID || created.Slug != "website" {
		t.Errorf("unexpected project: %+v", created)
	}
	if n := count(t, db, &schemas.Column{}); n != 4 {
		t.Errorf("default columns = %d, want 4", n)
	}

	send(t, s, "project.updated", wire(t, "Sablier", "p2", map[string]any{"facile_id": "proj-1", "name": "Website v2"}))
	if err := db.First(&created, created.ID).Error; err != nil || created.Name != "Website v2" || created.Description != "Redesign" {
		t.Errorf("update not applied or clobbered description: %+v %v", created, err)
	}

	send(t, s, "project.deleted", wire(t, "Sablier", "p3", map[string]any{"facile_id": "proj-1"}))
	if n := count(t, db, &schemas.Project{}); n != 0 {
		t.Errorf("project not deleted, %d left", n)
	}
	if n := count(t, db, &schemas.Column{}); n != 0 {
		t.Errorf("columns not deleted, %d left", n)
	}
}

func TestEventsFromOtherAppsAreIgnored(t *testing.T) {
	s, db, _ := newSyncer(t)
	raw, _ := json.Marshal(map[string]any{"facile_id": "bare-1", "name": "Bare"})
	send(t, s, "project.created", raw)
	send(t, s, "project.created", wire(t, "Agenda", "a1", map[string]any{"facile_id": "agenda-1", "name": "Agenda"}))
	if n := count(t, db, &schemas.Project{}); n != 0 {
		t.Fatalf("foreign events applied, %d projects", n)
	}
}

func TestStaleAndLegacyEventsAreDropped(t *testing.T) {
	s, db, _ := newSyncer(t)
	old, _ := json.Marshal(map[string]any{
		"app": "Sablier", "idempotency_key": "old", "timestamp": time.Now().Add(-40 * 24 * time.Hour).UTC().Format(time.RFC3339),
		"payload": map[string]any{"facile_id": "old-1", "name": "Old"},
	})
	undated, _ := json.Marshal(map[string]any{"app": "Sablier", "payload": map[string]any{"facile_id": "undated", "name": "Undated"}})
	send(t, s, "project.created", old)
	send(t, s, "project.created", undated)
	send(t, s, "project.created", wire(t, "Sablier", "n1", map[string]any{"facile_id": "42", "name": "Numeric"}))
	if n := count(t, db, &schemas.Project{}); n != 0 {
		t.Fatalf("stale or legacy events created %d projects", n)
	}
}

func TestStaleTaskUpdateDoesNotRevertLocalEdit(t *testing.T) {
	s, db, _ := newSyncer(t)
	seedSablierProject(t, s)
	stale, _ := json.Marshal(map[string]any{
		"app": "Sablier", "idempotency_key": "s1", "timestamp": time.Now().Add(-time.Hour).UTC().Format(time.RFC3339),
		"payload": map[string]any{"facile_id": "task-1", "project_facile_id": "proj-1", "name": "Original"},
	})
	send(t, s, "task.created", stale)
	db.Model(&schemas.Task{}).Where("facile_id = ?", "task-1").Updates(map[string]any{"title": "Edited locally", "updated_at": time.Now()})
	again, _ := json.Marshal(map[string]any{
		"app": "Sablier", "idempotency_key": "s2", "timestamp": time.Now().Add(-30 * time.Minute).UTC().Format(time.RFC3339),
		"payload": map[string]any{"facile_id": "task-1", "project_facile_id": "proj-1", "name": "Replayed"},
	})
	send(t, s, "task.updated", again)
	var task schemas.Task
	db.Where("facile_id = ?", "task-1").First(&task)
	if task.Title != "Edited locally" {
		t.Errorf("stale update reverted local edit: %q", task.Title)
	}
}

func TestActorMustBeWorkspaceMember(t *testing.T) {
	s, db, _ := newSyncer(t)
	seedSablierProject(t, s)
	send(t, s, "task.created", wire(t, "Sablier", "t1", map[string]any{
		"facile_id": "task-1", "project_facile_id": "proj-1", "name": "Outsider", "actor_email": "alice@facile.studio",
	}))
	var task schemas.Task
	db.Where("facile_id = ?", "task-1").First(&task)
	if task.ActorID != nil {
		t.Fatalf("assigned a non-member: %v", *task.ActorID)
	}
	var user schemas.User
	db.First(&user)
	db.Create(&schemas.WorkspaceMember{WorkspaceID: s.workspaceID, UserID: user.ID, Role: "member", JoinedAt: time.Now()})
	send(t, s, "task.updated", wire(t, "Sablier", "t2", map[string]any{
		"facile_id": "task-1", "project_facile_id": "proj-1", "name": "Outsider", "actor_email": "alice@facile.studio",
	}))
	db.Where("facile_id = ?", "task-1").First(&task)
	if task.ActorID == nil || *task.ActorID != user.ID {
		t.Fatalf("member not assigned: %+v", task.ActorID)
	}
}

func TestTaskLifecycleFromSablier(t *testing.T) {
	s, db, notifier := newSyncer(t)
	seedSablierProject(t, s)
	var alice schemas.User
	db.First(&alice)
	db.Create(&schemas.WorkspaceMember{WorkspaceID: s.workspaceID, UserID: alice.ID, Role: "member", JoinedAt: time.Now()})

	send(t, s, "task.created", wire(t, "Sablier", "t1", map[string]any{
		"facile_id": "task-1", "project_facile_id": "proj-1", "name": "Write copy",
		"status": "in-progress", "actor_email": "alice@facile.studio",
	}))
	var task schemas.Task
	if err := db.Where("facile_id = ?", "task-1").First(&task).Error; err != nil {
		t.Fatalf("task not created: %v", err)
	}
	if task.Title != "Write copy" || task.Status != "in-progress" || task.ColumnID == nil || task.ActorID == nil || task.Number != 1 {
		t.Errorf("unexpected task: %+v", task)
	}

	send(t, s, "task.updated", wire(t, "Sablier", "t2", map[string]any{
		"facile_id": "task-1", "project_facile_id": "proj-1", "name": "Write copy v2", "status": "done",
	}))
	db.First(&task, task.ID)
	if task.Title != "Write copy v2" || task.Status != "done" {
		t.Errorf("update not applied: %+v", task)
	}

	send(t, s, "task.deleted", wire(t, "Sablier", "t3", map[string]any{"facile_id": "task-1"}))
	if n := count(t, db, &schemas.Task{}); n != 0 {
		t.Errorf("task not deleted")
	}
	want := []string{realtime.TypeTaskCreated, realtime.TypeTaskUpdated, realtime.TypeTaskDeleted}
	if len(notifier.types) != 3 || notifier.types[0] != want[0] || notifier.types[1] != want[1] || notifier.types[2] != want[2] {
		t.Errorf("realtime notifications = %v, want %v", notifier.types, want)
	}
}

func TestUpdateOfUnknownTaskDoesNotCreateIt(t *testing.T) {
	s, db, _ := newSyncer(t)
	seedSablierProject(t, s)
	send(t, s, "task.updated", wire(t, "Sablier", "t1", map[string]any{
		"facile_id": "late-1", "project_facile_id": "proj-1", "name": "Late", "status": "to-do",
	}))
	if n := count(t, db, &schemas.Task{}); n != 0 {
		t.Fatalf("tasks = %d, want 0", n)
	}
	send(t, s, "project.updated", wire(t, "Sablier", "p9", map[string]any{"facile_id": "ghost", "name": "Ghost"}))
	if n := count(t, db, &schemas.Project{}); n != 1 {
		t.Fatalf("projects = %d, want 1", n)
	}
}

func TestReplayIsIdempotent(t *testing.T) {
	s, db, notifier := newSyncer(t)
	seedSablierProject(t, s)
	raw := wire(t, "Sablier", "t1", map[string]any{"facile_id": "task-1", "project_facile_id": "proj-1", "name": "Once"})
	send(t, s, "task.created", raw)
	db.Model(&schemas.Task{}).Where("facile_id = ?", "task-1").Update("title", "Edited locally")
	send(t, s, "task.created", raw)
	send(t, s, "task.updated", raw)

	var tasks []schemas.Task
	db.Find(&tasks)
	if len(tasks) != 1 || tasks[0].Title != "Edited locally" {
		t.Errorf("replay changed state: %+v", tasks)
	}
	if len(notifier.types) != 1 {
		t.Errorf("replay notified again: %v", notifier.types)
	}
	if n := count(t, db, &schemas.PoolProcessedEvent{}); n != 2 {
		t.Errorf("ledger rows = %d, want 2", n)
	}
}

func TestOwnEventsAreIgnored(t *testing.T) {
	s, db, _ := newSyncer(t)
	send(t, s, "project.created", wire(t, "Opus", "o1", map[string]any{"facile_id": "mine", "name": "Mine"}))
	if count(t, db, &schemas.Project{}) != 0 || count(t, db, &schemas.PoolProcessedEvent{}) != 0 {
		t.Error("own event was applied or recorded")
	}
}

func TestTaskForUnknownProjectIsDropped(t *testing.T) {
	s, db, _ := newSyncer(t)
	raw := wire(t, "Sablier", "t1", map[string]any{"facile_id": "orphan", "project_facile_id": "proj-1", "name": "Orphan"})
	send(t, s, "task.created", raw)
	if count(t, db, &schemas.Task{}) != 0 {
		t.Fatal("orphan task created")
	}
	if count(t, db, &schemas.PoolProcessedEvent{}) != 0 {
		t.Fatal("dropped event was recorded, a retry could never apply it")
	}
	seedSablierProject(t, s)
	send(t, s, "task.created", raw)
	if count(t, db, &schemas.Task{}) != 1 {
		t.Fatal("redelivery after the project arrived did not apply")
	}
}

func TestProjectNeedsTargetWorkspace(t *testing.T) {
	s, db, _ := newSyncer(t)
	s.workspaceID = 0
	seedSablierProject(t, s)
	if count(t, db, &schemas.Project{}) != 0 {
		t.Fatal("project created without a target workspace")
	}
}

func TestDeleteOfUnknownEntitiesIsHarmless(t *testing.T) {
	s, _, _ := newSyncer(t)
	send(t, s, "project.deleted", wire(t, "Sablier", "d1", map[string]any{"facile_id": "ghost"}))
	send(t, s, "task.deleted", wire(t, "Sablier", "d2", map[string]any{"facile_id": "ghost"}))
}
