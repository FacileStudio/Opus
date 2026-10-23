package comment

import (
	"context"
	"strconv"
	"testing"

	"github.com/FacileStudio/Opus/apps/api/internal/realtime"
	"github.com/FacileStudio/Opus/apps/api/schemas"
)

type recordingPublisher struct {
	messages []realtime.BroadcastMessage
	projects []string
}

func (p *recordingPublisher) BroadcastToProject(projectID string, msg realtime.BroadcastMessage, _ string) error {
	p.projects = append(p.projects, projectID)
	p.messages = append(p.messages, msg)
	return nil
}

func TestCreateCommentBroadcastsOnTheTaskProject(t *testing.T) {
	db := newTestDB(t)
	service := NewService(db)
	publisher := &recordingPublisher{}
	service.SetPublisher(publisher)
	userID, _, taskID := seedCommentFixture(t, db)

	if _, err := service.controller.create(context.Background(), userID, taskID, &CreateCommentRequest{Content: "hi"}); err != nil {
		t.Fatalf("create: %v", err)
	}

	var task schemas.Task
	db.First(&task, taskID)
	if len(publisher.messages) != 1 ||
		publisher.messages[0].Type != realtime.TypeCommentUpdated ||
		publisher.messages[0].TaskID != strconv.FormatInt(taskID, 10) ||
		publisher.projects[0] != strconv.FormatInt(task.ProjectID, 10) {
		t.Fatalf("unexpected broadcast: %+v on %v", publisher.messages, publisher.projects)
	}
}
