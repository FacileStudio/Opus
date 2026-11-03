package poolsync

import (
	"context"
	"testing"
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"gorm.io/gorm"
)

func projectByFacileID(t *testing.T, db *gorm.DB, facileID string) schemas.Project {
	t.Helper()
	var project schemas.Project
	if err := db.Where("facile_id = ?", facileID).First(&project).Error; err != nil {
		t.Fatalf("project %s: %v", facileID, err)
	}
	return project
}

func localProject(t *testing.T, db *gorm.DB, workspaceID int64, name, facileID string) {
	t.Helper()
	record := schemas.Project{WorkspaceID: workspaceID, Name: name, Slug: name, FacileRef: schemas.FacileRef{FacileID: &facileID}}
	if err := db.Create(&record).Error; err != nil {
		t.Fatalf("seed project: %v", err)
	}
}

func TestProjectSpaceResolution(t *testing.T) {
	s, db, _ := newSyncer(t)
	s.SetIssuer("https://sso")
	var alice schemas.User
	db.First(&alice)
	linked := "space-1"
	team := schemas.Workspace{Name: "Team", Slug: "team", CreatedAt: time.Now(), RegistreSpaceID: &linked}
	personal := schemas.Workspace{Name: "Personal", Slug: "personal", CreatedAt: time.Now(), PersonalOwnerID: &alice.ID}
	db.Create(&team)
	db.Create(&personal)
	db.Exec("INSERT INTO porte_identities (user_id, provider, subject) VALUES (?, ?, ?)", alice.ID, "https://sso", "sub-alice")
	cases := []struct {
		facileID string
		space    map[string]any
		want     int64
	}{
		{"p-team", map[string]any{"registre_space_id": "space-1", "personal": false}, team.ID},
		{"p-subject", map[string]any{"registre_space_id": "", "personal": true, "owner_subject": "sub-alice"}, personal.ID},
		{"p-email", map[string]any{"personal": true, "owner_subject": "other", "owner_email": "ALICE@facile.studio"}, personal.ID},
	}
	for _, tc := range cases {
		send(t, s, "project.created", wire(t, "Sablier", tc.facileID, map[string]any{
			"facile_id": tc.facileID, "name": tc.facileID, "space": tc.space,
		}))
		if got := projectByFacileID(t, db, tc.facileID).WorkspaceID; got != tc.want {
			t.Errorf("%s landed in workspace %d, want %d", tc.facileID, got, tc.want)
		}
	}
}

func TestUnresolvableSpaceIsNotMarkedProcessed(t *testing.T) {
	s, db, _ := newSyncer(t)
	raw := wire(t, "Sablier", "u1", map[string]any{
		"facile_id": "p-1", "name": "Later", "space": map[string]any{"registre_space_id": "space-9"},
	})
	send(t, s, "project.created", raw)
	send(t, s, "project.created", wire(t, "Sablier", "u2", map[string]any{
		"facile_id": "p-2", "name": "Unlinked", "space": map[string]any{"personal": false},
	}))
	if count(t, db, &schemas.Project{}) != 0 || count(t, db, &schemas.PoolProcessedEvent{}) != 0 {
		t.Fatal("unresolvable project applied or recorded")
	}
	linked := "space-9"
	team := schemas.Workspace{Name: "Team", Slug: "team", CreatedAt: time.Now(), RegistreSpaceID: &linked}
	db.Create(&team)
	send(t, s, "project.created", raw)
	if projectByFacileID(t, db, "p-1").WorkspaceID != team.ID {
		t.Fatal("redelivery after the space was linked did not apply")
	}
}

func TestProjectLastWriterWinsAndSpaceMove(t *testing.T) {
	s, db, _ := newSyncer(t)
	seedSablierProject(t, s)
	send(t, s, "task.created", wire(t, "Sablier", "t1", map[string]any{"facile_id": "task-1", "project_facile_id": "proj-1", "name": "Copy"}))
	db.Model(&schemas.Project{}).Where("facile_id = ?", "proj-1").Update("updated_at", time.Now())
	send(t, s, "project.updated", wire(t, "Sablier", "p2", map[string]any{
		"facile_id": "proj-1", "name": "Stale", "updated_at": time.Now().Add(-time.Hour).Format(time.RFC3339Nano),
	}))
	if projectByFacileID(t, db, "proj-1").Name != "Website" {
		t.Fatal("older update overwrote a newer local edit")
	}
	linked := "space-1"
	team := schemas.Workspace{Name: "Team", Slug: "team", CreatedAt: time.Now(), RegistreSpaceID: &linked}
	db.Create(&team)
	send(t, s, "project.updated", wire(t, "Sablier", "p3", map[string]any{
		"facile_id": "proj-1", "name": "Fresh", "updated_at": time.Now().Add(time.Minute).Format(time.RFC3339Nano),
		"space": map[string]any{"registre_space_id": "space-1"},
	}))
	moved := projectByFacileID(t, db, "proj-1")
	var task schemas.Task
	db.Where("facile_id = ?", "task-1").First(&task)
	if moved.Name != "Fresh" || moved.WorkspaceID != team.ID || task.WorkspaceID != team.ID {
		t.Fatalf("newer update with a new space not applied: %+v, task workspace %d", moved, task.WorkspaceID)
	}
}

func TestTwinMergeKeepsTheSmallerFacileID(t *testing.T) {
	s, db, _ := newSyncer(t)
	localProject(t, db, s.workspaceID, "Website", "m-local")
	localProject(t, db, s.workspaceID, "Docs", "b-local")
	var adopted []int64
	s.OnAdopted(func(_ context.Context, projectID int64) { adopted = append(adopted, projectID) })
	send(t, s, "project.created", wire(t, "Sablier", "k1", map[string]any{"facile_id": "a-in", "name": "website"}))
	send(t, s, "project.created", wire(t, "Sablier", "k2", map[string]any{"facile_id": "z-in", "name": "DOCS"}))
	if projectByFacileID(t, db, "a-in").Name != "Website" || projectByFacileID(t, db, "b-local").Name != "Docs" {
		t.Fatal("tie-break did not keep the smaller facile_id on each twin")
	}
	if len(adopted) != 1 || adopted[0] != projectByFacileID(t, db, "a-in").ID {
		t.Fatalf("adopting a facile_id must re-emit that project's tasks once, got %v", adopted)
	}
	if count(t, db, &schemas.Project{}) != 2 || count(t, db, &schemas.PoolProcessedEvent{}) != 2 {
		t.Fatal("twin events created a project or were left unprocessed")
	}
	send(t, s, "project.created", wire(t, "Sablier", "k3", map[string]any{"facile_id": "0-in", "name": "Website"}))
	if count(t, db, &schemas.Project{}) != 3 {
		t.Fatal("a twin already linked to Sablier was merged again")
	}
	website := projectByFacileID(t, db, "a-in")
	local := "t-m"
	db.Create(&schemas.Task{ProjectID: website.ID, WorkspaceID: website.WorkspaceID, Title: "Copy", FacileRef: schemas.FacileRef{FacileID: &local}})
	send(t, s, "task.created", wire(t, "Sablier", "k4", map[string]any{"facile_id": "t-a", "project_facile_id": "a-in", "name": "copy"}))
	if count(t, db, &schemas.Task{}) != 1 || db.Where("facile_id = ?", "t-a").First(&schemas.Task{}).Error != nil {
		t.Fatal("task twin did not adopt the smaller incoming facile_id")
	}
}

func TestTaskUpdateKeepsAbsentFieldsAndMoves(t *testing.T) {
	s, db, _ := newSyncer(t)
	seedSablierProject(t, s)
	send(t, s, "project.created", wire(t, "Sablier", "p2", map[string]any{"facile_id": "proj-2", "name": "Docs"}))
	send(t, s, "task.created", wire(t, "Sablier", "t1", map[string]any{
		"facile_id": "task-1", "project_facile_id": "proj-1", "name": "Spec",
		"description": "Write it", "priority": "high", "due_date": "2026-12-01",
	}))
	send(t, s, "task.updated", wire(t, "Sablier", "t2", map[string]any{"facile_id": "task-1", "project_facile_id": "proj-2", "name": "Spec"}))
	var task schemas.Task
	db.Where("facile_id = ?", "task-1").First(&task)
	if task.ProjectID != projectByFacileID(t, db, "proj-2").ID || task.Description != "Write it" || task.Priority != "high" ||
		task.DueDate == nil || task.DueDate.Format(time.DateOnly) != "2026-12-01" {
		t.Fatalf("move lost fields or did not move: %+v", task)
	}
	send(t, s, "task.updated", wire(t, "Sablier", "t3", map[string]any{"facile_id": "task-1", "due_date": nil, "priority": "none"}))
	var cleared schemas.Task
	db.First(&cleared, task.ID)
	if cleared.DueDate != nil || cleared.Priority != "no-priority" {
		t.Fatalf("explicit clear not applied: %+v", cleared)
	}
}

func TestTimeEntryMirror(t *testing.T) {
	s, db, _ := newSyncer(t)
	seedSablierProject(t, s)
	send(t, s, "task.created", wire(t, "Sablier", "t1", map[string]any{"facile_id": "task-1", "project_facile_id": "proj-1", "name": "Copy"}))
	start := time.Now().Add(-2 * time.Hour).UTC().Truncate(time.Second)
	entry := map[string]any{
		"facile_id": "te-1", "task_facile_id": "task-1", "project_facile_id": "proj-1", "user_email": "ALICE@facile.studio",
		"started_at": start.Format(time.RFC3339), "stopped_at": nil, "paused_duration_ms": 0, "description": "Draft",
	}
	send(t, s, "time_entry.created", wire(t, "Sablier", "e1", entry))
	var mirror schemas.TimeEntry
	if err := db.Where("facile_id = ?", "te-1").First(&mirror).Error; err != nil {
		t.Fatalf("mirror not created: %v", err)
	}
	if mirror.Source != schemas.TimeEntrySourceSablier || mirror.EndTime != nil || mirror.Duration != 0 || mirror.UserID == nil {
		t.Fatalf("running entry mirrored wrong: %+v", mirror)
	}
	entry["stopped_at"] = start.Add(time.Hour).Format(time.RFC3339)
	entry["paused_duration_ms"] = 60000
	send(t, s, "time_entry.updated", wire(t, "Sablier", "e2", entry))
	db.First(&mirror, mirror.ID)
	if mirror.EndTime == nil || mirror.Duration != 3540 {
		t.Fatalf("stopped entry mirrored wrong: %+v", mirror)
	}
	send(t, s, "time_entry.deleted", wire(t, "Sablier", "e3", map[string]any{"facile_id": "te-1"}))
	entry["facile_id"], entry["task_facile_id"] = "te-2", "ghost"
	send(t, s, "time_entry.created", wire(t, "Sablier", "e4", entry))
	if count(t, db, &schemas.TimeEntry{}) != 0 || db.First(&schemas.PoolProcessedEvent{}, "idempotency_key = ?", "e4").Error == nil {
		t.Fatal("deleted mirror kept, or entry on an unknown task applied or recorded")
	}
}
