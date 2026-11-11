package github

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/FacileStudio/Opus/apps/api/internal/database/dbtest"
	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/porte"
	"github.com/go-chi/chi/v5"
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
	return userID, "dev@facile.studio", nil
}

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := dbtest.Open(t)
	return db
}

func signPayload(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestExtractTaskRefs(t *testing.T) {
	tests := []struct {
		texts []string
		want  []TaskRef
	}{
		{
			texts: []string{"feat: opus-42 add github integration"},
			want:  []TaskRef{{ProjectSlug: "opus", TaskNumber: 42}},
		},
		{
			texts: []string{"heads/facile-10-fix", "Closes FACILE-10 and opus-99"},
			want: []TaskRef{
				{ProjectSlug: "facile", TaskNumber: 10},
				{ProjectSlug: "opus", TaskNumber: 99},
			},
		},
		{
			texts: []string{"no issue here", "random-branch"},
			want:  nil,
		},
	}

	for _, tt := range tests {
		got := extractTaskRefs(tt.texts...)
		if len(got) != len(tt.want) {
			t.Fatalf("extractTaskRefs(%v) = %v, want %v", tt.texts, got, tt.want)
		}
		for i, r := range got {
			if r.ProjectSlug != tt.want[i].ProjectSlug || r.TaskNumber != tt.want[i].TaskNumber {
				t.Errorf("ref[%d] = %+v, want %+v", i, r, tt.want[i])
			}
		}
	}
}

func TestVerifySignature(t *testing.T) {
	secret := "test-secret"
	body := []byte(`{"hello":"world"}`)
	sig := signPayload(body, secret)

	if !verifySignature(body, sig, secret) {
		t.Fatal("expected signature to be valid")
	}
	if verifySignature(body, "sha256=invalid", secret) {
		t.Fatal("expected invalid signature to fail")
	}
	if !verifySignature(body, "", "") {
		t.Fatal("expected empty secret to allow any signature")
	}
}

type testFixture struct {
	db      *gorm.DB
	svc     *Service
	task    schemas.Task
	user    schemas.User
	colProg schemas.Column
	colRev  schemas.Column
	colDone schemas.Column
}

func setupFixture(t *testing.T) *testFixture {
	t.Helper()
	db := setupTestDB(t)
	svc := NewService(db, "secret-123")
	ghID := "12345"
	user := schemas.User{Email: "yann@facile.studio", Name: "Yann", GitHubID: &ghID}
	db.Create(&user)
	ws := schemas.Workspace{Name: "Facile", Slug: "facile"}
	db.Create(&ws)
	proj := schemas.Project{WorkspaceID: ws.ID, Name: "Opus", Slug: "opus"}
	db.Create(&proj)
	colTodo := schemas.Column{ProjectID: proj.ID, Name: "To Do", Slug: "to-do", Position: 0}
	colProg := schemas.Column{ProjectID: proj.ID, Name: "In Progress", Slug: "in-progress", Position: 1}
	colRev := schemas.Column{ProjectID: proj.ID, Name: "In Review", Slug: "in-review", Position: 2}
	colDone := schemas.Column{ProjectID: proj.ID, Name: "Done", Slug: "done", Position: 3, IsFinal: true}
	db.Create(&colTodo)
	db.Create(&colProg)
	db.Create(&colRev)
	db.Create(&colDone)
	task := schemas.Task{
		ProjectID:   proj.ID,
		WorkspaceID: ws.ID,
		Number:      42,
		Title:       "Add GitHub webhook",
		Status:      colTodo.Slug,
		ColumnID:    &colTodo.ID,
	}
	db.Create(&task)
	return &testFixture{db: db, svc: svc, task: task, user: user, colProg: colProg, colRev: colRev, colDone: colDone}
}

func TestProcessWebhookPullRequestLifecycle(t *testing.T) {
	f := setupFixture(t)
	ctx := context.Background()

	payload := WebhookPayload{
		Action: "opened",
		Number: 101,
		PullRequest: &PullRequestPayload{
			ID:      1001,
			Number:  101,
			Title:   "feat: opus-42 webhook receiver",
			HTMLURL: "https://github.com/FacileStudio/Opus/pull/101",
			State:   "open",
			Draft:   true,
			Head:    BranchRef{Ref: "yann/opus-42-webhook"},
			Base:    BranchRef{Ref: "main"},
			User:    &SenderPayload{ID: 12345, Login: "saravenpi"},
		},
		Repository: &RepositoryPayload{
			ID:            555,
			FullName:      "FacileStudio/Opus",
			Name:          "Opus",
			DefaultBranch: "main",
		},
	}

	if err := f.svc.ProcessWebhook(ctx, &payload); err != nil {
		t.Fatalf("webhook open: %v", err)
	}

	var task schemas.Task
	f.db.First(&task, f.task.ID)
	if task.Status != "in-progress" || *task.ColumnID != f.colProg.ID || *task.ActorID != f.user.ID {
		t.Fatalf("task after open = %+v", task)
	}

	payload.Action = "ready_for_review"
	payload.PullRequest.Draft = false
	if err := f.svc.ProcessWebhook(ctx, &payload); err != nil {
		t.Fatalf("webhook review: %v", err)
	}
	f.db.First(&task, f.task.ID)
	if task.Status != "in-review" || *task.ColumnID != f.colRev.ID {
		t.Fatalf("task after review = %+v", task)
	}

	payload.Action = "closed"
	payload.PullRequest.State = "closed"
	payload.PullRequest.Merged = true
	if err := f.svc.ProcessWebhook(ctx, &payload); err != nil {
		t.Fatalf("webhook merge: %v", err)
	}
	f.db.First(&task, f.task.ID)
	if task.Status != "done" || *task.ColumnID != f.colDone.ID {
		t.Fatalf("task after merge = %+v", task)
	}
}

func TestRouterWebhookAndListEndpoints(t *testing.T) {
	db := setupTestDB(t)
	secret := "secret-xyz"
	svc := NewService(db, secret)

	ws := schemas.Workspace{Name: "W", Slug: "w"}
	db.Create(&ws)
	proj := schemas.Project{WorkspaceID: ws.ID, Name: "P", Slug: "p"}
	db.Create(&proj)
	task := schemas.Task{ProjectID: proj.ID, WorkspaceID: ws.ID, Number: 7, Title: "Test", Status: "to-do"}
	db.Create(&task)

	r := chi.NewRouter()
	RegisterRoutes(r, svc, mockAuth{userID: 1})

	payload := WebhookPayload{
		Action: "opened",
		Number: 12,
		PullRequest: &PullRequestPayload{
			ID:      99,
			Number:  12,
			Title:   "p-7 something",
			HTMLURL: "https://github.com/FacileStudio/p/pull/12",
			State:   "open",
			Head:    BranchRef{Ref: "p-7"},
			Base:    BranchRef{Ref: "main"},
		},
		Repository: &RepositoryPayload{FullName: "FacileStudio/p"},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/integrations/github/webhook", bytes.NewReader(body))
	req.Header.Set("X-Hub-Signature-256", signPayload(body, secret))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("webhook status = %d, want 200: %s", w.Code, w.Body.String())
	}

	getReq := httptest.NewRequest(http.MethodGet, "/project/1/task/"+strconv.FormatInt(task.ID, 10)+"/pull-requests/", nil)
	getW := httptest.NewRecorder()
	r.ServeHTTP(getW, getReq)

	if getW.Code != http.StatusOK {
		t.Fatalf("list PRs status = %d, want 200: %s", getW.Code, getW.Body.String())
	}
	var resps []PullRequestResponse
	if err := json.NewDecoder(getW.Body).Decode(&resps); err != nil {
		t.Fatalf("decode PR responses: %v", err)
	}
	if len(resps) != 1 || resps[0].PRNumber != 12 {
		t.Fatalf("responses = %+v, want pr 12", resps)
	}
}
