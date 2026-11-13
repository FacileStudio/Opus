package github

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/FacileStudio/Opus/apps/api/internal/githubapp"
	"github.com/FacileStudio/Opus/apps/api/internal/githubapp/githubapptest"
	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type repoFixture struct {
	db      *gorm.DB
	fake    *githubapptest.Fake
	app     *githubapp.Client
	project schemas.Project
	admin   schemas.User
	member  schemas.User
}

func setupRepoFixture(t *testing.T) *repoFixture {
	t.Helper()
	db := setupTestDB(t)
	adminGitHub, memberGitHub := "42", "77"
	admin := schemas.User{Email: "admin@facile.studio", Name: "Admin", GitHubID: &adminGitHub}
	member := schemas.User{Email: "member@facile.studio", Name: "Member", GitHubID: &memberGitHub}
	db.Create(&admin)
	db.Create(&member)
	ws := schemas.Workspace{Name: "W", Slug: "w"}
	db.Create(&ws)
	db.Create(&schemas.WorkspaceMember{WorkspaceID: ws.ID, UserID: admin.ID, Role: "admin"})
	db.Create(&schemas.WorkspaceMember{WorkspaceID: ws.ID, UserID: member.ID, Role: "member"})
	project := schemas.Project{WorkspaceID: ws.ID, Name: "P", Slug: "p"}
	db.Create(&project)
	fake := githubapptest.New(t)
	fake.Installed["acme/app"] = true
	fake.Logins["42"] = "alice"
	fake.Logins["77"] = "bob"
	fake.Permissions["alice"] = "write"
	fake.Permissions["bob"] = "write"
	return &repoFixture{db: db, fake: fake, app: fake.Client, project: project, admin: admin, member: member}
}

func (f *repoFixture) serve(t *testing.T, userID int64, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	router := chi.NewRouter()
	RegisterRepoRoutes(router, NewRepos(f.db, f.app), mockAuth{userID: userID})
	req := httptest.NewRequest(method, "/project/"+strconv.FormatInt(f.project.ID, 10)+path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestLinkingNeedsAWorkspaceAdminWithWriteOnGitHub(t *testing.T) {
	f := setupRepoFixture(t)
	link := `{"repo":"acme/app"}`
	if rec := f.serve(t, f.member.ID, http.MethodPost, "/github-repos", link); rec.Code != http.StatusForbidden {
		t.Fatalf("plain member: %d %s", rec.Code, rec.Body)
	}
	f.fake.Permissions["alice"] = "read"
	if rec := f.serve(t, f.admin.ID, http.MethodPost, "/github-repos", link); rec.Code != http.StatusForbidden {
		t.Fatalf("read-only on GitHub: %d %s", rec.Code, rec.Body)
	}
	f.fake.Logins["43"] = "carol"
	f.fake.Permissions["carol"] = "admin"
	f.db.Model(&schemas.User{}).Where("id = ?", f.admin.ID).Update("github_id", "43")
	rec := f.serve(t, f.admin.ID, http.MethodPost, "/github-repos", link)
	if rec.Code != http.StatusCreated {
		t.Fatalf("admin with GitHub admin: %d %s", rec.Code, rec.Body)
	}
	var linked RepoResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &linked); err != nil {
		t.Fatal(err)
	}
	if linked.FullName != "acme/app" || linked.DefaultBranch != "trunk" {
		t.Fatalf("linked = %+v", linked)
	}
	if rec := f.serve(t, f.admin.ID, http.MethodPost, "/github-repos", link); rec.Code != http.StatusConflict {
		t.Fatalf("second link: %d", rec.Code)
	}
}

func TestMissingPreconditionsAnswer412(t *testing.T) {
	f := setupRepoFixture(t)
	f.app = nil
	if rec := f.serve(t, f.member.ID, http.MethodGet, "/github-issues", ""); rec.Code != http.StatusPreconditionFailed {
		t.Fatalf("no App: %d", rec.Code)
	}
	f.app = f.fake.Client
	f.db.Model(&schemas.User{}).Where("id = ?", f.admin.ID).Update("github_id", nil)
	if rec := f.serve(t, f.admin.ID, http.MethodPost, "/github-repos", `{"repo":"acme/app"}`); rec.Code != http.StatusPreconditionFailed {
		t.Fatalf("no github_id: %d", rec.Code)
	}
	f.db.Model(&schemas.User{}).Where("id = ?", f.admin.ID).Update("github_id", "42")
	if rec := f.serve(t, f.admin.ID, http.MethodPost, "/github-repos", `{"repo":"acme/other"}`); rec.Code != http.StatusPreconditionFailed {
		t.Fatalf("App not installed: %d", rec.Code)
	}
	if rec := f.serve(t, f.admin.ID, http.MethodPost, "/github-repos", `{"repo":"not a repo"}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("bad name: %d", rec.Code)
	}
}

func TestMembersReadIssuesAndOutsidersDoNot(t *testing.T) {
	f := setupRepoFixture(t)
	f.db.Create(&schemas.ProjectGitHubRepo{ProjectID: f.project.ID, RepoFullName: "acme/app", DefaultBranch: "main"})
	f.db.Create(&schemas.ProjectGitHubRepo{ProjectID: f.project.ID, RepoFullName: "acme/gone", DefaultBranch: "main"})
	f.fake.Issues["acme/app"] = []map[string]any{
		{"number": 3, "title": "Crash", "state": "open", "html_url": "https://github.com/acme/app/issues/3"},
		{"number": 4, "title": "PR", "state": "open", "pull_request": map[string]any{}},
	}
	rec := f.serve(t, f.member.ID, http.MethodGet, "/github-issues", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("member: %d %s", rec.Code, rec.Body)
	}
	var out IssuesResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Issues) != 1 || out.Issues[0].Number != 3 || out.Issues[0].Repo != "acme/app" {
		t.Fatalf("issues = %+v", out.Issues)
	}
	if len(out.Unavailable) != 1 || out.Unavailable[0].Repo != "acme/gone" {
		t.Fatalf("unavailable = %+v", out.Unavailable)
	}
	outsider := schemas.User{Email: "out@facile.studio", Name: "Out"}
	f.db.Create(&outsider)
	if rec := f.serve(t, outsider.ID, http.MethodGet, "/github-issues", ""); rec.Code != http.StatusForbidden {
		t.Fatalf("outsider: %d", rec.Code)
	}
}

func TestUnlinkNeedsAnAdminButNotTheApp(t *testing.T) {
	f := setupRepoFixture(t)
	row := schemas.ProjectGitHubRepo{ProjectID: f.project.ID, RepoFullName: "acme/app", DefaultBranch: "main"}
	f.db.Create(&row)
	path := "/github-repos/" + strconv.FormatInt(row.ID, 10)
	f.app = nil
	if rec := f.serve(t, f.member.ID, http.MethodDelete, path, ""); rec.Code != http.StatusForbidden {
		t.Fatalf("member unlink: %d", rec.Code)
	}
	if rec := f.serve(t, f.admin.ID, http.MethodDelete, path, ""); rec.Code != http.StatusNoContent {
		t.Fatalf("admin unlink: %d %s", rec.Code, rec.Body)
	}
}
