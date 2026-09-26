package column

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

type mockAuth struct {
	userID int64
}

func (m mockAuth) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := porte.WithIdentity(r.Context(), porte.Identity{UserID: m.userID, SessionID: 1})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m mockAuth) IdentityForUser(_ context.Context, userID int64) (int64, string, error) {
	return userID, "user@facile.studio", nil
}

func newTestService(t *testing.T) (*Service, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&schemas.User{},
		&schemas.Workspace{},
		&schemas.WorkspaceMember{},
		&schemas.Project{},
		&schemas.Column{},
		&schemas.Task{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return NewService(db), db
}

func seedProject(t *testing.T, db *gorm.DB) (int64, int64) {
	t.Helper()
	user := schemas.User{Email: "user@facile.studio", Name: "User"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	workspace := schemas.Workspace{Name: "Workspace", Slug: "ws", CreatedAt: time.Now()}
	if err := db.Create(&workspace).Error; err != nil {
		t.Fatalf("seed workspace: %v", err)
	}
	member := schemas.WorkspaceMember{
		WorkspaceID: workspace.ID,
		UserID:      user.ID,
		Role:        "admin",
		JoinedAt:    time.Now(),
	}
	if err := db.Create(&member).Error; err != nil {
		t.Fatalf("seed member: %v", err)
	}
	project := schemas.Project{
		WorkspaceID: workspace.ID,
		Name:        "Project",
		Slug:        "project",
		CreatedAt:   time.Now(),
	}
	if err := db.Create(&project).Error; err != nil {
		t.Fatalf("seed project: %v", err)
	}
	return user.ID, project.ID
}

func TestCreateColumn(t *testing.T) {
	service, db := newTestService(t)
	userID, projectID := seedProject(t, db)
	ctx := context.Background()

	col1, err := service.controller.create(ctx, userID, projectID, &CreateColumnRequest{Name: "Backlog"})
	if err != nil {
		t.Fatalf("create column 1: %v", err)
	}
	if col1.Position != 0 || col1.Slug != "backlog" {
		t.Fatalf("col1 unexpected position or slug: %+v", col1)
	}

	col2, err := service.controller.create(ctx, userID, projectID, &CreateColumnRequest{Name: "In Progress"})
	if err != nil {
		t.Fatalf("create column 2: %v", err)
	}
	if col2.Position != 1 || col2.Slug != "in-progress" {
		t.Fatalf("col2 unexpected position or slug: %+v", col2)
	}

	_, err = service.controller.create(ctx, userID+999, projectID, &CreateColumnRequest{Name: "Forbidden"})
	if err == nil || errors.Status(err) != 403 {
		t.Fatalf("expected forbidden for stranger, got %v", err)
	}
}

func TestCreateColumnValidation(t *testing.T) {
	service, db := newTestService(t)
	userID, projectID := seedProject(t, db)
	ctx := context.Background()

	_, err := service.controller.create(ctx, userID, projectID, &CreateColumnRequest{Name: "---"})
	if err == nil || errors.Status(err) != 400 {
		t.Fatalf("expected 400 for empty slug, got %v", err)
	}

	_, err = service.controller.create(ctx, userID, projectID, &CreateColumnRequest{Name: "Planned"})
	if err == nil || errors.Status(err) != 409 {
		t.Fatalf("expected 409 for virtual status, got %v", err)
	}

	_, err = service.controller.create(ctx, userID, projectID, &CreateColumnRequest{Name: "Done"})
	if err != nil {
		t.Fatalf("create done: %v", err)
	}
	_, err = service.controller.create(ctx, userID, projectID, &CreateColumnRequest{Name: "done"})
	if err == nil || errors.Status(err) != 409 {
		t.Fatalf("expected 409 for duplicate slug, got %v", err)
	}
}

func TestListColumns(t *testing.T) {
	service, db := newTestService(t)
	userID, projectID := seedProject(t, db)
	ctx := context.Background()

	service.controller.create(ctx, userID, projectID, &CreateColumnRequest{Name: "Col 1"})
	service.controller.create(ctx, userID, projectID, &CreateColumnRequest{Name: "Col 2"})

	cols, err := service.controller.list(ctx, userID, projectID)
	if err != nil {
		t.Fatalf("list columns: %v", err)
	}
	if len(cols) != 2 || cols[0].Position != 0 || cols[1].Position != 1 {
		t.Fatalf("unexpected columns list: %+v", cols)
	}

	_, err = service.controller.list(ctx, userID+999, projectID)
	if err == nil || errors.Status(err) != 403 {
		t.Fatalf("expected 403 for stranger list, got %v", err)
	}
}

func TestUpdateColumn(t *testing.T) {
	service, db := newTestService(t)
	userID, projectID := seedProject(t, db)
	ctx := context.Background()

	created, err := service.controller.create(ctx, userID, projectID, &CreateColumnRequest{Name: "Old Name"})
	if err != nil {
		t.Fatalf("create column: %v", err)
	}

	newName := "New Name"
	isFinal := true
	updated, err := service.controller.update(ctx, userID, created.ID, &UpdateColumnRequest{
		Name:    &newName,
		IsFinal: &isFinal,
	})
	if err != nil {
		t.Fatalf("update column: %v", err)
	}
	if updated.Name != "New Name" || !updated.IsFinal {
		t.Fatalf("unexpected updated column: %+v", updated)
	}

	_, err = service.controller.update(ctx, userID, 99999, &UpdateColumnRequest{Name: &newName})
	if err == nil || errors.Status(err) != 404 {
		t.Fatalf("expected 404 for missing column, got %v", err)
	}
}

func TestDeleteColumn(t *testing.T) {
	service, db := newTestService(t)
	userID, projectID := seedProject(t, db)
	ctx := context.Background()

	created, err := service.controller.create(ctx, userID, projectID, &CreateColumnRequest{Name: "To Delete"})
	if err != nil {
		t.Fatalf("create column: %v", err)
	}

	task := schemas.Task{
		ProjectID:   projectID,
		WorkspaceID: 1,
		ColumnID:    &created.ID,
		Title:       "Task",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := db.Create(&task).Error; err != nil {
		t.Fatalf("create task: %v", err)
	}

	_, err = service.controller.delete(ctx, userID, created.ID)
	if err == nil || errors.Status(err) != 409 {
		t.Fatalf("expected 409 deleting column with tasks, got %v", err)
	}

	if err := db.Delete(&task).Error; err != nil {
		t.Fatalf("delete task: %v", err)
	}

	deleted, err := service.controller.delete(ctx, userID, created.ID)
	if err != nil {
		t.Fatalf("delete column: %v", err)
	}
	if deleted.ID != created.ID {
		t.Fatalf("unexpected deleted id: %d", deleted.ID)
	}
}

func TestReorderColumns(t *testing.T) {
	service, db := newTestService(t)
	userID, projectID := seedProject(t, db)
	ctx := context.Background()

	c1, _ := service.controller.create(ctx, userID, projectID, &CreateColumnRequest{Name: "Col A"})
	c2, _ := service.controller.create(ctx, userID, projectID, &CreateColumnRequest{Name: "Col B"})

	reordered, err := service.controller.reorder(ctx, userID, projectID, &ReorderColumnsRequest{
		Columns: []ColumnOrderItem{
			{ID: c1.ID, Position: 1},
			{ID: c2.ID, Position: 0},
		},
	})
	if err != nil {
		t.Fatalf("reorder columns: %v", err)
	}
	if len(reordered) != 2 || reordered[0].ID != c2.ID || reordered[1].ID != c1.ID {
		t.Fatalf("unexpected reordered result: %+v", reordered)
	}

	_, err = service.controller.reorder(ctx, userID, projectID, &ReorderColumnsRequest{
		Columns: []ColumnOrderItem{{ID: 99999, Position: 0}},
	})
	if err == nil || errors.Status(err) != 400 {
		t.Fatalf("expected 400 for unknown column, got %v", err)
	}
}

func TestHTTPRoutes(t *testing.T) {
	service, db := newTestService(t)
	userID, projectID := seedProject(t, db)

	router := chi.NewRouter()
	RegisterRoutes(router, service, mockAuth{userID: userID})

	body, _ := json.Marshal(CreateColumnRequest{Name: "Sprint 1"})
	req := httptest.NewRequest(http.MethodPost, "/column/"+strconv.FormatInt(projectID, 10), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("POST /column/{projectId} status = %d, body = %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/column/"+strconv.FormatInt(projectID, 10), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /column/{projectId} status = %d, body = %s", w.Code, w.Body.String())
	}

	var cols []ColumnResponse
	if err := json.Unmarshal(w.Body.Bytes(), &cols); err != nil || len(cols) != 1 {
		t.Fatalf("failed to decode response: %v, body = %s", err, w.Body.String())
	}
}
