package comment

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/tronc/errors"
	"github.com/go-chi/chi/v5"
	"gorm.io/driver/sqlite"
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
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	err = db.AutoMigrate(
		&schemas.User{},
		&schemas.Workspace{},
		&schemas.WorkspaceMember{},
		&schemas.Project{},
		&schemas.Task{},
		&schemas.Comment{},
	)
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func seedCommentFixture(t *testing.T, db *gorm.DB) (int64, int64, int64) {
	t.Helper()
	user := schemas.User{Email: "user@facile.studio", Name: "User One", Image: "avatar.png"}
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
	task := schemas.Task{ProjectID: project.ID, WorkspaceID: workspace.ID, Title: "Task One", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := db.Create(&task).Error; err != nil {
		t.Fatalf("seed task: %v", err)
	}
	return user.ID, workspace.ID, task.ID
}

func TestCreateAndListComment(t *testing.T) {
	db := newTestDB(t)
	service := NewService(db)
	userID, _, taskID := seedCommentFixture(t, db)
	ctx := context.Background()

	comment, err := service.Create(ctx, userID, taskID, "first comment")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if comment.ID == 0 || comment.Content != "first comment" {
		t.Fatalf("unexpected comment: %+v", comment)
	}

	records, err := service.ListForTask(ctx, userID, taskID)
	if err != nil {
		t.Fatalf("ListForTask: %v", err)
	}
	if len(records) != 1 || records[0].Comment.ID != comment.ID {
		t.Fatalf("unexpected records count or id: %+v", records)
	}
	if records[0].AuthorName != "User One" || records[0].AuthorImage != "avatar.png" {
		t.Fatalf("unexpected author: name=%s, img=%s", records[0].AuthorName, records[0].AuthorImage)
	}
}

func TestTaskAccessChecks(t *testing.T) {
	db := newTestDB(t)
	service := NewService(db)
	userID, _, taskID := seedCommentFixture(t, db)
	ctx := context.Background()

	_, err := service.Create(ctx, userID+999, taskID, "forbidden")
	if err == nil || errors.Status(err) != 403 {
		t.Fatalf("want 403, got %v", err)
	}

	_, err = service.ListForTask(ctx, userID+999, taskID)
	if err == nil || errors.Status(err) != 403 {
		t.Fatalf("want 403, got %v", err)
	}

	_, err = service.Create(ctx, userID, taskID+999, "not found")
	if err == nil || errors.Status(err) != 404 {
		t.Fatalf("want 404, got %v", err)
	}

	_, err = service.ListForTask(ctx, userID, taskID+999)
	if err == nil || errors.Status(err) != 404 {
		t.Fatalf("want 404, got %v", err)
	}
}

func TestUpdateComment(t *testing.T) {
	db := newTestDB(t)
	service := NewService(db)
	userID, _, taskID := seedCommentFixture(t, db)
	ctx := context.Background()

	comment, err := service.Create(ctx, userID, taskID, "initial")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	updated, err := service.Update(ctx, userID, comment.ID, "updated")
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Content != "updated" {
		t.Fatalf("expected updated content, got %s", updated.Content)
	}

	_, err = service.Update(ctx, userID+999, comment.ID, "forbidden")
	if err == nil || errors.Status(err) != 403 {
		t.Fatalf("want 403, got %v", err)
	}

	_, err = service.Update(ctx, userID, comment.ID+999, "not found")
	if err == nil || errors.Status(err) != 404 {
		t.Fatalf("want 404, got %v", err)
	}
}

func TestDeleteComment(t *testing.T) {
	db := newTestDB(t)
	service := NewService(db)
	userID, _, taskID := seedCommentFixture(t, db)
	ctx := context.Background()

	comment, err := service.Create(ctx, userID, taskID, "to delete")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	_, err = service.Delete(ctx, userID+999, comment.ID)
	if err == nil || errors.Status(err) != 403 {
		t.Fatalf("want 403, got %v", err)
	}

	deleted, err := service.Delete(ctx, userID, comment.ID)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if deleted.ID != comment.ID {
		t.Fatalf("deleted id mismatch")
	}

	_, err = service.Delete(ctx, userID, comment.ID)
	if err == nil || errors.Status(err) != 404 {
		t.Fatalf("want 404 after deletion, got %v", err)
	}
}

func TestControllerValidation(t *testing.T) {
	db := newTestDB(t)
	service := NewService(db)
	userID, _, taskID := seedCommentFixture(t, db)
	ctx := context.Background()

	_, err := service.controller.list(ctx, userID, 0)
	if err == nil || errors.Status(err) != 400 {
		t.Fatalf("want 400 for invalid task id, got %v", err)
	}

	_, err = service.controller.create(ctx, userID, taskID, &CreateCommentRequest{Content: "   "})
	if err == nil || errors.Status(err) != 400 {
		t.Fatalf("want 400 for empty content, got %v", err)
	}

	_, err = service.controller.update(ctx, userID, 0, &UpdateCommentRequest{Content: "valid"})
	if err == nil || errors.Status(err) != 400 {
		t.Fatalf("want 400 for invalid comment id, got %v", err)
	}

	_, err = service.controller.update(ctx, userID, 1, &UpdateCommentRequest{Content: ""})
	if err == nil || errors.Status(err) != 400 {
		t.Fatalf("want 400 for empty content, got %v", err)
	}

	_, err = service.controller.delete(ctx, userID, 0)
	if err == nil || errors.Status(err) != 400 {
		t.Fatalf("want 400 for invalid comment id, got %v", err)
	}
}

func TestRouterEndpoints(t *testing.T) {
	db := newTestDB(t)
	service := NewService(db)
	userID, _, taskID := seedCommentFixture(t, db)
	auth := &testAuth{userID: userID, email: "user@facile.studio"}

	router := chi.NewRouter()
	RegisterRoutes(router, service, auth)

	body, err := json.Marshal(CreateCommentRequest{Content: "from http"})
	if err != nil {
		t.Fatalf("marshal create: %v", err)
	}
	req := httptest.NewRequest("POST", "/comment/"+strconv.FormatInt(taskID, 10), bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /comment/{taskId} = %d, want 200", rec.Code)
	}

	var created CommentResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal created: %v", err)
	}

	req = httptest.NewRequest("GET", "/comment/"+strconv.FormatInt(taskID, 10), nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /comment/{taskId} = %d, want 200", rec.Code)
	}

	upBody, err := json.Marshal(UpdateCommentRequest{Content: "edited http"})
	if err != nil {
		t.Fatalf("marshal update: %v", err)
	}
	req = httptest.NewRequest("PUT", "/comment/"+strconv.FormatInt(created.ID, 10), bytes.NewReader(upBody))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("PUT /comment/{id} = %d, want 200", rec.Code)
	}

	req = httptest.NewRequest("DELETE", "/comment/"+strconv.FormatInt(created.ID, 10), nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("DELETE /comment/{id} = %d, want 200", rec.Code)
	}
}
