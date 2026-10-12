package taskrelation

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/FacileStudio/Opus/apps/api/internal/wireid"
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
		&schemas.TaskRelation{},
	)
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

type fixture struct {
	userID      int64
	workspaceID int64
	taskID1     int64
	taskID2     int64
}

func seedFixture(t *testing.T, db *gorm.DB) fixture {
	t.Helper()
	user := schemas.User{Email: "test@facile.studio", Name: "Test User"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	ws := schemas.Workspace{Name: "Engineering", Slug: "engineering"}
	if err := db.Create(&ws).Error; err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	member := schemas.WorkspaceMember{WorkspaceID: ws.ID, UserID: user.ID, Role: "admin"}
	if err := db.Create(&member).Error; err != nil {
		t.Fatalf("create member: %v", err)
	}
	project := schemas.Project{WorkspaceID: ws.ID, Name: "Core", Slug: "core"}
	if err := db.Create(&project).Error; err != nil {
		t.Fatalf("create project: %v", err)
	}
	t1 := schemas.Task{ProjectID: project.ID, WorkspaceID: ws.ID, Title: "Task 1", ActorID: &user.ID}
	if err := db.Create(&t1).Error; err != nil {
		t.Fatalf("create task 1: %v", err)
	}
	t2 := schemas.Task{ProjectID: project.ID, WorkspaceID: ws.ID, Title: "Task 2"}
	if err := db.Create(&t2).Error; err != nil {
		t.Fatalf("create task 2: %v", err)
	}
	return fixture{userID: user.ID, workspaceID: ws.ID, taskID1: t1.ID, taskID2: t2.ID}
}

func TestRequireMember(t *testing.T) {
	db := newTestDB(t)
	fix := seedFixture(t, db)
	service := NewService(db)
	ctx := context.Background()

	if err := service.RequireMember(ctx, fix.userID, fix.workspaceID); err != nil {
		t.Fatalf("RequireMember for member: %v", err)
	}
	err := service.RequireMember(ctx, fix.userID+999, fix.workspaceID)
	if err == nil {
		t.Fatal("RequireMember for stranger succeeded, want Forbidden")
	}
	if errors.Status(err) != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", errors.Status(err))
	}
}

func TestCreateRelation(t *testing.T) {
	db := newTestDB(t)
	fix := seedFixture(t, db)
	service := NewService(db)
	ctx := context.Background()

	req := CreateTaskRelationRequest{
		SourceTaskID: wireid.ID(fix.taskID1),
		TargetTaskID: wireid.ID(fix.taskID2),
		RelationType: "subtask",
	}
	rel, err := service.CreateRelation(ctx, fix.userID, req)
	if err != nil {
		t.Fatalf("CreateRelation: %v", err)
	}
	if rel.ID == 0 || rel.RelationType != "subtask" {
		t.Fatalf("unexpected relation: %+v", rel)
	}

	_, err = service.CreateRelation(ctx, fix.userID, req)
	if err == nil || errors.Status(err) != http.StatusConflict {
		t.Fatalf("duplicate relation err = %v, want 409", err)
	}

	revReq := CreateTaskRelationRequest{
		SourceTaskID: wireid.ID(fix.taskID2),
		TargetTaskID: wireid.ID(fix.taskID1),
		RelationType: "subtask",
	}
	_, err = service.CreateRelation(ctx, fix.userID, revReq)
	if err == nil || errors.Status(err) != http.StatusConflict {
		t.Fatalf("reverse duplicate relation err = %v, want 409", err)
	}

	_, err = service.CreateRelation(ctx, fix.userID+999, req)
	if err == nil || errors.Status(err) != http.StatusForbidden {
		t.Fatalf("stranger create err = %v, want 403", err)
	}
}

func TestListAndTaskAccess(t *testing.T) {
	db := newTestDB(t)
	fix := seedFixture(t, db)
	service := NewService(db)
	ctx := context.Background()

	req := CreateTaskRelationRequest{
		SourceTaskID: wireid.ID(fix.taskID1),
		TargetTaskID: wireid.ID(fix.taskID2),
		RelationType: "blocks",
	}
	if _, err := service.CreateRelation(ctx, fix.userID, req); err != nil {
		t.Fatalf("CreateRelation: %v", err)
	}

	records, err := service.ListForTask(ctx, fix.userID, fix.taskID1)
	if err != nil {
		t.Fatalf("ListForTask: %v", err)
	}
	if len(records) != 1 || records[0].SourceTask == nil || records[0].TargetTask == nil {
		t.Fatalf("unexpected records: %+v", records)
	}
	if records[0].SourceTask.Title != "Task 1" || records[0].TargetTask.Title != "Task 2" {
		t.Fatalf("unexpected titles in records: %+v", records)
	}

	records2, err := service.ListForTask(ctx, fix.userID, fix.taskID2)
	if err != nil || len(records2) != 1 {
		t.Fatalf("ListForTask target: %v, len = %d", err, len(records2))
	}

	_, err = service.ListForTask(ctx, fix.userID, fix.taskID1+999)
	if err == nil || errors.Status(err) != http.StatusNotFound {
		t.Fatalf("ListForTask missing task err = %v, want 404", err)
	}

	_, err = service.ListForTask(ctx, fix.userID+999, fix.taskID1)
	if err == nil || errors.Status(err) != http.StatusForbidden {
		t.Fatalf("ListForTask stranger err = %v, want 403", err)
	}
}

func TestDeleteRelation(t *testing.T) {
	db := newTestDB(t)
	fix := seedFixture(t, db)
	service := NewService(db)
	ctx := context.Background()

	req := CreateTaskRelationRequest{
		SourceTaskID: wireid.ID(fix.taskID1),
		TargetTaskID: wireid.ID(fix.taskID2),
		RelationType: "related",
	}
	rel, err := service.CreateRelation(ctx, fix.userID, req)
	if err != nil {
		t.Fatalf("CreateRelation: %v", err)
	}

	deleted, err := service.DeleteRelation(ctx, fix.userID, rel.ID)
	if err != nil || deleted.ID != rel.ID {
		t.Fatalf("DeleteRelation: %v", err)
	}

	_, err = service.DeleteRelation(ctx, fix.userID, rel.ID)
	if err == nil || errors.Status(err) != http.StatusNotFound {
		t.Fatalf("DeleteRelation repeated err = %v, want 404", err)
	}
}

func TestControllerValidation(t *testing.T) {
	db := newTestDB(t)
	fix := seedFixture(t, db)
	service := NewService(db)
	ctx := context.Background()

	_, err := service.controller.create(ctx, fix.userID, nil)
	if err == nil || errors.Status(err) != http.StatusBadRequest {
		t.Fatalf("create nil req err = %v, want 400", err)
	}

	selfReq := &CreateTaskRelationRequest{
		SourceTaskID: wireid.ID(fix.taskID1),
		TargetTaskID: wireid.ID(fix.taskID1),
		RelationType: "subtask",
	}
	_, err = service.controller.create(ctx, fix.userID, selfReq)
	if err == nil || errors.Status(err) != http.StatusBadRequest {
		t.Fatalf("create self relation err = %v, want 400", err)
	}

	invalidTypeReq := &CreateTaskRelationRequest{
		SourceTaskID: wireid.ID(fix.taskID1),
		TargetTaskID: wireid.ID(fix.taskID2),
		RelationType: "unknown",
	}
	_, err = service.controller.create(ctx, fix.userID, invalidTypeReq)
	if err == nil || errors.Status(err) != http.StatusBadRequest {
		t.Fatalf("create invalid type err = %v, want 400", err)
	}
}

func TestRouterEndpoints(t *testing.T) {
	db := newTestDB(t)
	fix := seedFixture(t, db)
	service := NewService(db)
	auth := &testAuth{userID: fix.userID, email: "test@facile.studio"}

	router := chi.NewRouter()
	RegisterRoutes(router, service, auth)

	body, _ := json.Marshal(CreateTaskRelationRequest{
		SourceTaskID: wireid.ID(fix.taskID1),
		TargetTaskID: wireid.ID(fix.taskID2),
		RelationType: "subtask",
	})
	req := httptest.NewRequest(http.MethodPost, "/task-relation", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("POST /task-relation status = %d: %s", w.Code, w.Body.String())
	}

	var created TaskRelationResponse
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal created relation: %v", err)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/task-relation/"+strconv.FormatInt(fix.taskID1, 10), nil)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)
	if getW.Code != http.StatusOK {
		t.Fatalf("GET /task-relation status = %d: %s", getW.Code, getW.Body.String())
	}

	delReq := httptest.NewRequest(http.MethodDelete, "/task-relation/"+strconv.FormatInt(created.ID, 10), nil)
	delW := httptest.NewRecorder()
	router.ServeHTTP(delW, delReq)
	if delW.Code != http.StatusOK {
		t.Fatalf("DELETE /task-relation status = %d: %s", delW.Code, delW.Body.String())
	}
}
