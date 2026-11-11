package activity

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/FacileStudio/Opus/apps/api/internal/database/dbtest"
	"github.com/FacileStudio/Opus/apps/api/internal/wireid"
	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/tronc/errors"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type testAuth struct {
	userID int64
	email  string
}

func (a *testAuth) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := porte.WithIdentity(r.Context(), porte.Identity{UserID: a.userID, SessionID: 1})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *testAuth) IdentityForUser(_ context.Context, userID int64) (int64, string, error) {
	return userID, a.email, nil
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := dbtest.Open(t)
	return db
}

func seedActivityFixture(t *testing.T, db *gorm.DB) (int64, int64, int64) {
	t.Helper()
	user := schemas.User{Email: "user@facile.studio", Name: "User One"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	workspace := schemas.Workspace{Name: "Workspace", Slug: "workspace", CreatedAt: time.Now()}
	if err := db.Create(&workspace).Error; err != nil {
		t.Fatalf("seed workspace: %v", err)
	}
	member := schemas.WorkspaceMember{WorkspaceID: workspace.ID, UserID: user.ID, Role: "member", JoinedAt: time.Now()}
	if err := db.Create(&member).Error; err != nil {
		t.Fatalf("seed member: %v", err)
	}
	project := schemas.Project{WorkspaceID: workspace.ID, Name: "Project", Slug: "proj", CreatedAt: time.Now()}
	if err := db.Create(&project).Error; err != nil {
		t.Fatalf("seed project: %v", err)
	}
	task := schemas.Task{ProjectID: project.ID, WorkspaceID: workspace.ID, Title: "Task", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := db.Create(&task).Error; err != nil {
		t.Fatalf("seed task: %v", err)
	}
	return user.ID, workspace.ID, task.ID
}

func TestCreateAndListActivity(t *testing.T) {
	db := newTestDB(t)
	service := NewService(db)
	userID, _, taskID := seedActivityFixture(t, db)
	ctx := context.Background()

	act, err := service.Create(ctx, userID, ActivityParams{
		TaskID:    taskID,
		UserID:    &userID,
		Type:      "status_changed",
		EventData: `{"oldStatus":"todo","newStatus":"done"}`,
	})
	if err != nil || act.ID == 0 {
		t.Fatalf("Create activity failed: %v", err)
	}

	list, err := service.ListForTask(ctx, userID, taskID)
	if err != nil || len(list) != 1 {
		t.Fatalf("ListForTask failed: %v, len=%d", err, len(list))
	}
	if list[0].Type != "status_changed" {
		t.Fatalf("unexpected activity type: %s", list[0].Type)
	}
}

func TestCommentActivityFlow(t *testing.T) {
	db := newTestDB(t)
	service := NewService(db)
	userID, _, taskID := seedActivityFixture(t, db)
	ctx := context.Background()

	cmt, err := service.CreateComment(ctx, userID, taskID, "first comment")
	if err != nil || cmt.ID == 0 {
		t.Fatalf("CreateComment failed: %v", err)
	}

	up, err := service.UpdateComment(ctx, userID, cmt.ID, "updated comment")
	if err != nil || up.Content != "updated comment" {
		t.Fatalf("UpdateComment failed: %v", err)
	}

	_, err = service.UpdateComment(ctx, userID+999, cmt.ID, "hacked")
	if err == nil || errors.Status(err) != 404 {
		t.Fatalf("want 404 for non-author update, got %v", err)
	}

	_, err = service.DeleteComment(ctx, userID+999, cmt.ID)
	if err == nil || errors.Status(err) != 404 {
		t.Fatalf("want 404 for non-author delete, got %v", err)
	}

	del, err := service.DeleteComment(ctx, userID, cmt.ID)
	if err != nil || del.ID != cmt.ID {
		t.Fatalf("DeleteComment failed: %v", err)
	}

	_, err = service.DeleteComment(ctx, userID, cmt.ID)
	if err == nil || errors.Status(err) != 404 {
		t.Fatalf("want 404 after delete, got %v", err)
	}
}

func TestAccessChecks(t *testing.T) {
	db := newTestDB(t)
	service := NewService(db)
	userID, _, taskID := seedActivityFixture(t, db)
	ctx := context.Background()

	_, err := service.ListForTask(ctx, userID+999, taskID)
	if err == nil || errors.Status(err) != 403 {
		t.Fatalf("want 403 for unauthorized list, got %v", err)
	}

	_, err = service.CreateComment(ctx, userID+999, taskID, "forbidden")
	if err == nil || errors.Status(err) != 403 {
		t.Fatalf("want 403 for unauthorized create, got %v", err)
	}

	_, err = service.ListForTask(ctx, userID, taskID+999)
	if err == nil || errors.Status(err) != 404 {
		t.Fatalf("want 404 for missing task list, got %v", err)
	}

	_, err = service.CreateComment(ctx, userID, taskID+999, "not found")
	if err == nil || errors.Status(err) != 404 {
		t.Fatalf("want 404 for missing task comment, got %v", err)
	}
}

func TestControllerValidation(t *testing.T) {
	db := newTestDB(t)
	service := NewService(db)
	userID, _, taskID := seedActivityFixture(t, db)
	ctx := context.Background()

	_, err := service.controller.list(ctx, userID, 0)
	if err == nil || errors.Status(err) != 400 {
		t.Fatalf("want 400 for list invalid task id, got %v", err)
	}

	_, err = service.controller.create(ctx, userID, &CreateActivityRequest{TaskID: 0})
	if err == nil || errors.Status(err) != 400 {
		t.Fatalf("want 400 for create invalid task id, got %v", err)
	}

	_, err = service.controller.create(ctx, userID, &CreateActivityRequest{TaskID: wireid.ID(taskID), Type: " "})
	if err == nil || errors.Status(err) != 400 {
		t.Fatalf("want 400 for create empty type, got %v", err)
	}

	_, err = service.controller.createComment(ctx, userID, &CommentActivityRequest{TaskID: wireid.ID(taskID), Comment: " "})
	if err == nil || errors.Status(err) != 400 {
		t.Fatalf("want 400 for create empty comment, got %v", err)
	}

	_, err = service.controller.updateComment(ctx, userID, &UpdateCommentActivityRequest{ActivityID: 0, Comment: "ok"})
	if err == nil || errors.Status(err) != 400 {
		t.Fatalf("want 400 for update invalid activity id, got %v", err)
	}

	_, err = service.controller.deleteComment(ctx, userID, &DeleteCommentActivityRequest{ActivityID: 0})
	if err == nil || errors.Status(err) != 400 {
		t.Fatalf("want 400 for delete invalid activity id, got %v", err)
	}
}

func executeJSONRequest(handler http.Handler, method, path string, payload any) *httptest.ResponseRecorder {
	var body *bytes.Reader
	if payload != nil {
		data, _ := json.Marshal(payload)
		body = bytes.NewReader(data)
	} else {
		body = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, body)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestRouterEndpoints(t *testing.T) {
	db := newTestDB(t)
	service := NewService(db)
	userID, _, taskID := seedActivityFixture(t, db)
	auth := &testAuth{userID: userID, email: "user@facile.studio"}

	router := chi.NewRouter()
	RegisterRoutes(router, service, auth)

	msg := "hello activity\n\n\nworld"
	rec := executeJSONRequest(router, "POST", "/activity/create", CreateActivityRequest{
		TaskID:  wireid.ID(taskID),
		Message: &msg,
		Type:    "task",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /activity/create = %d, want 200", rec.Code)
	}

	rec = executeJSONRequest(router, "GET", "/activity/"+strconv.FormatInt(taskID, 10), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /activity/{taskId} = %d, want 200", rec.Code)
	}

	rec = executeJSONRequest(router, "POST", "/activity/comment", CommentActivityRequest{
		TaskID:  wireid.ID(taskID),
		Comment: "http comment",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /activity/comment = %d, want 200", rec.Code)
	}
	var cmt ActivityResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &cmt); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	rec = executeJSONRequest(router, "PUT", "/activity/comment", UpdateCommentActivityRequest{
		ActivityID: wireid.ID(cmt.ID),
		Comment:    "edited comment",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("PUT /activity/comment = %d, want 200", rec.Code)
	}

	rec = executeJSONRequest(router, "DELETE", "/activity/comment", DeleteCommentActivityRequest{
		ActivityID: wireid.ID(cmt.ID),
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("DELETE /activity/comment = %d, want 200", rec.Code)
	}
}
