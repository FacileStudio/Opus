package events

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
)

func drainedTask(t *testing.T, workspace func(*schemas.User) schemas.Workspace) map[string]any {
	t.Helper()
	pub, fake, db := outboxPublisher(t)
	pub.config.ClientURL = "https://opus.example.test"
	owner := schemas.User{Email: "Owner@Facile.studio"}
	db.Create(&owner)
	ws := workspace(&owner)
	db.Create(&ws)
	project := schemas.Project{WorkspaceID: ws.ID, Name: "Launch", Slug: "launch"}
	db.Create(&project)
	task := schemas.Task{ProjectID: project.ID, WorkspaceID: ws.ID, Title: "Ship it", Status: "todo"}
	db.Create(&task)

	ctx := context.Background()
	pub.EmitTaskUpdated(ctx, NewTaskPayload(task, "", ""))
	if _, err := pub.drainer.Drain(ctx); err != nil {
		t.Fatalf("drain: %v", err)
	}
	msgs := fake.getMessages()
	if len(msgs) != 1 {
		t.Fatalf("want one task.updated, got %d", len(msgs))
	}
	var sent struct {
		Payload map[string]any `json:"payload"`
	}
	if err := json.Unmarshal(msgs[0].payload.(json.RawMessage), &sent); err != nil {
		t.Fatalf("decode: %v", err)
	}
	want := "https://opus.example.test/dashboard/workspace/" + strconv.FormatInt(ws.ID, 10) +
		"/project/" + strconv.FormatInt(project.ID, 10) + "/task/" + strconv.FormatInt(task.ID, 10)
	if sent.Payload["url"] != want {
		t.Errorf("url = %v, want %s", sent.Payload["url"], want)
	}
	return sent.Payload
}

func TestTaskPayloadNamesItsSpace(t *testing.T) {
	spaceID := "space-1"
	payload := drainedTask(t, func(*schemas.User) schemas.Workspace {
		return schemas.Workspace{Name: "Studio", Slug: "studio", CreatedAt: time.Now(), RegistreSpaceID: &spaceID}
	})
	if payload["registre_space_id"] != spaceID {
		t.Errorf("registre_space_id = %v, want %s", payload["registre_space_id"], spaceID)
	}
	if _, ok := payload["personal"]; ok {
		t.Errorf("a space task must not say personal: %v", payload)
	}
}

func TestPersonalTaskPayloadNamesItsOwner(t *testing.T) {
	payload := drainedTask(t, func(owner *schemas.User) schemas.Workspace {
		return schemas.Workspace{Name: "Personal", Slug: "personal", CreatedAt: time.Now(), PersonalOwnerID: &owner.ID}
	})
	if payload["personal"] != true || payload["owner_email"] != "owner@facile.studio" {
		t.Errorf("personal task not described: %v", payload)
	}
	if _, ok := payload["registre_space_id"]; ok {
		t.Errorf("a personal task must not carry a space id: %v", payload)
	}
}
