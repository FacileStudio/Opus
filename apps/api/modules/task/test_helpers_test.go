package task

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/FacileStudio/Opus/apps/api/internal/database/dbtest"
	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/porte"
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
	db := dbtest.Open(t)
	return NewService(db), db
}

func seedProject(t *testing.T, db *gorm.DB) (int64, int64, schemas.Column) {
	t.Helper()
	user := schemas.User{Email: "user@facile.studio", Name: "Test User"}
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
	col := schemas.Column{
		ProjectID: project.ID,
		Name:      "To Do",
		Slug:      "to-do",
		Position:  0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := db.Create(&col).Error; err != nil {
		t.Fatalf("seed column: %v", err)
	}
	return user.ID, project.ID, col
}
