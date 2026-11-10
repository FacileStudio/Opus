package main

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/Opus/apps/api/internal/env"
	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/directory"
	"github.com/FacileStudio/porte/exchange"
	"github.com/FacileStudio/porte/forfait"
	"github.com/go-chi/chi/v5"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type testAuth struct{}

func (testAuth) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.Header.Get("X-Test-User"), 10, 64)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(porte.WithIdentity(r.Context(), porte.Identity{UserID: id})))
	})
}

func (testAuth) IdentityForUser(_ context.Context, userID int64) (int64, string, error) {
	return userID, "member@facile.studio", nil
}

type noIdentities struct{}

func (noIdentities) ListByUser(context.Context, int64) ([]porte.StoredIdentity, error) {
	return nil, nil
}
func (noIdentities) Save(context.Context, porte.StoredIdentity) error { return nil }

func newTestIdeas(t *testing.T) (*ideas, *gorm.DB, chi.Router) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&schemas.User{}, &schemas.Workspace{}, &schemas.WorkspaceMember{}, &schemas.ForfaitState{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	exchanger, err := exchange.New(exchange.Config{Issuer: "https://sso.test", ClientID: "c", ClientSecret: "s", Identities: noIdentities{}})
	if err != nil {
		t.Fatalf("exchange: %v", err)
	}
	i := &ideas{
		db:         db,
		env:        &env.Config{BulleURL: "https://bulle.test", OIDC: &env.OIDCConfig{Issuer: "https://sso.test"}},
		logger:     slog.New(slog.NewTextHandler(io.Discard, nil)),
		plans:      forfait.NewStore(forfait.Feature{Key: aiAccessKey, Kind: forfait.Bool, Default: []byte("false")}),
		exchanger:  exchanger,
		configured: true,
	}
	router := chi.NewRouter()
	router.Route("/api", func(api chi.Router) { i.mount(api, testAuth{}) })
	return i, db, router
}

func seedSharedWorkspace(t *testing.T, db *gorm.DB, space string) (int64, int64) {
	t.Helper()
	user := schemas.User{Email: "member@facile.studio", Name: "Member"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	workspace := schemas.Workspace{Name: "Studio", Slug: "studio", CreatedAt: time.Now(), RegistreSpaceID: &space}
	if err := db.Create(&workspace).Error; err != nil {
		t.Fatal(err)
	}
	member := schemas.WorkspaceMember{WorkspaceID: workspace.ID, UserID: user.ID, Role: "admin", JoinedAt: time.Now()}
	if err := db.Create(&member).Error; err != nil {
		t.Fatal(err)
	}
	return user.ID, workspace.ID
}

func serve(router http.Handler, method, target string, userID int64) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, nil)
	if userID != 0 {
		req.Header.Set("X-Test-User", strconv.FormatInt(userID, 10))
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestIdeasAreBehindAuth(t *testing.T) {
	_, _, router := newTestIdeas(t)
	for _, route := range []struct{ method, path string }{
		{http.MethodGet, "/api/ideas"},
		{http.MethodPost, "/api/ideas/generate"},
		{http.MethodPost, "/api/ideas/abc/run"},
		{http.MethodGet, "/api/ideas/runs/abc"},
	} {
		if rec := serve(router, route.method, route.path, 0); rec.Code != http.StatusUnauthorized {
			t.Errorf("%s %s = %d, want 401", route.method, route.path, rec.Code)
		}
	}
}

func TestIdeasPassTheActiveSpace(t *testing.T) {
	i, db, _ := newTestIdeas(t)
	userID, workspaceID := seedSharedWorkspace(t, db, "spc_team")
	query := "/api/ideas?workspaceId=" + strconv.FormatInt(workspaceID, 10)

	member := httptest.NewRequest(http.MethodGet, query, nil)
	member = member.WithContext(porte.WithIdentity(member.Context(), porte.Identity{UserID: userID}))
	if got := i.space(member); got != "spc_team" {
		t.Fatalf("space for a member = %q, want spc_team", got)
	}
	stranger := httptest.NewRequest(http.MethodGet, query, nil)
	stranger = stranger.WithContext(porte.WithIdentity(stranger.Context(), porte.Identity{UserID: userID + 99}))
	if got := i.space(stranger); got != "" {
		t.Fatalf("space for a non-member = %q, want personal", got)
	}
	personal := httptest.NewRequest(http.MethodGet, "/api/ideas", nil)
	personal = personal.WithContext(porte.WithIdentity(personal.Context(), porte.Identity{UserID: userID}))
	if got := i.space(personal); got != "" {
		t.Fatalf("space without a workspace = %q, want personal", got)
	}
}

func TestFreePlanGetsPaymentRequired(t *testing.T) {
	i, db, router := newTestIdeas(t)
	userID, workspaceID := seedSharedWorkspace(t, db, "spc_team")
	i.apply(directory.Snapshot{
		EntitlementsRevision: 1,
		Spaces: []directory.Space{{
			ID:           "spc_team",
			Entitlements: map[string]json.RawMessage{aiAccessKey: json.RawMessage("false")},
		}},
	})
	target := "/api/ideas/generate?workspaceId=" + strconv.FormatInt(workspaceID, 10)
	rec := serve(router, http.MethodPost, target, userID)
	if rec.Code != http.StatusPaymentRequired {
		t.Fatalf("generate on a free plan = %d %s, want 402", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "https://sso.test/spaces") {
		t.Fatalf("402 body %q lacks the upgrade url", rec.Body.String())
	}
	list := serve(router, http.MethodGet, "/api/ideas?workspaceId="+strconv.FormatInt(workspaceID, 10), userID)
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), `"ai_access":false`) {
		t.Fatalf("list on a free plan = %d %s, want ai_access false", list.Code, list.Body.String())
	}
}

func paidSpace(revision int64) directory.Snapshot {
	return directory.Snapshot{
		EntitlementsRevision: revision,
		Spaces: []directory.Space{{
			ID:           "spc_team",
			Entitlements: map[string]json.RawMessage{aiAccessKey: json.RawMessage("true")},
		}},
	}
}

func spaceRequest(userID, workspaceID int64) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/api/ideas?workspaceId="+strconv.FormatInt(workspaceID, 10), nil)
	return req.WithContext(porte.WithIdentity(req.Context(), porte.Identity{UserID: userID}))
}

func TestAIAccessFailsClosedBeforeEntitlements(t *testing.T) {
	i, db, _ := newTestIdeas(t)
	userID, workspaceID := seedSharedWorkspace(t, db, "spc_team")
	if i.aiAccess(spaceRequest(userID, workspaceID)) {
		t.Fatal("ai.access must be off before Registre delivers entitlements")
	}
	i.apply(paidSpace(0))
	if i.aiAccess(spaceRequest(userID, workspaceID)) {
		t.Fatal("a snapshot without an entitlements revision must not turn ai.access on")
	}
	i.apply(paidSpace(1))
	if !i.aiAccess(spaceRequest(userID, workspaceID)) {
		t.Fatal("a paid space must have ai.access once entitlements arrived")
	}
}

func TestPlansSurviveARestart(t *testing.T) {
	i, db, _ := newTestIdeas(t)
	userID, workspaceID := seedSharedWorkspace(t, db, "spc_team")
	i.apply(paidSpace(3))
	restarted := &ideas{
		db:     db,
		env:    i.env,
		logger: i.logger,
		plans:  forfait.NewStore(forfait.Feature{Key: aiAccessKey, Kind: forfait.Bool, Default: []byte("false")}),
	}
	restarted.loadPlans(context.Background())
	if !restarted.aiAccess(spaceRequest(userID, workspaceID)) {
		t.Fatal("the saved plans must apply after a restart, before Registre answers")
	}
	restarted.apply(paidSpace(2))
	if !restarted.aiAccess(spaceRequest(userID, workspaceID)) {
		t.Fatal("an older snapshot must not replace the restored plans")
	}
}
