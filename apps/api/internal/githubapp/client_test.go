package githubapp_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/FacileStudio/Opus/apps/api/internal/githubapp"
	"github.com/FacileStudio/Opus/apps/api/internal/githubapp/githubapptest"
)

func TestNewIsOffWithoutEnvAndRejectsHalfAConfig(t *testing.T) {
	client, err := githubapp.New("", "")
	if client != nil || err != nil {
		t.Fatalf("empty env: got %v, %v", client, err)
	}
	if _, err := githubapp.New("Iv1.x", ""); err == nil {
		t.Fatal("a client id without a key must fail")
	}
	if _, err := githubapp.New("Iv1.x", "bm90IGEga2V5"); err == nil {
		t.Fatal("a key that is not PEM must fail")
	}
}

func TestTokensAreScopedToOneRepoAndOnePermissionAndCached(t *testing.T) {
	fake := githubapptest.New(t)
	fake.Installed["acme/app"] = true
	ctx := context.Background()
	for range 2 {
		if _, err := fake.Client.Repository(ctx, "acme/app"); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := fake.Client.Issues(ctx, "acme/app", "open"); err != nil {
		t.Fatal(err)
	}
	scopes := fake.Scopes()
	if len(scopes) != 2 {
		t.Fatalf("want one mint per permission, got %d", len(scopes))
	}
	repos, _ := scopes[0]["repositories"].([]any)
	perms, _ := scopes[0]["permissions"].(map[string]any)
	if len(repos) != 1 || repos[0] != "app" || len(perms) != 1 || perms["metadata"] != "read" {
		t.Fatalf("metadata token scope = %v", scopes[0])
	}
	if perms, _ := scopes[1]["permissions"].(map[string]any); len(perms) != 1 || perms["issues"] != "read" {
		t.Fatalf("issues token scope = %v", scopes[1])
	}
}

func TestAnUninstalledRepoAndABadNameAreTyped(t *testing.T) {
	fake := githubapptest.New(t)
	ctx := context.Background()
	if _, err := fake.Client.Repository(ctx, "acme/other"); !errors.Is(err, githubapp.ErrNotInstalled) {
		t.Fatalf("uninstalled: %v", err)
	}
	if _, err := fake.Client.Repository(ctx, "acme/app/../x"); !errors.Is(err, githubapp.ErrInvalidRepo) {
		t.Fatalf("bad name: %v", err)
	}
}

func TestPermissionResolvesTheLoginThenTheCollaboratorLevel(t *testing.T) {
	fake := githubapptest.New(t)
	fake.Installed["acme/app"] = true
	fake.Logins["42"] = "alice"
	fake.Permissions["alice"] = "write"
	ctx := context.Background()
	cases := map[string]string{"42": "write", "99": "none", "not-an-id": "none"}
	for id, want := range cases {
		got, err := fake.Client.Permission(ctx, "acme/app", id)
		if err != nil || got != want {
			t.Fatalf("Permission(%s) = %q, %v; want %q", id, got, err, want)
		}
	}
}

func TestIssuesDropPullRequestsAndServeFromCache(t *testing.T) {
	fake := githubapptest.New(t)
	fake.Installed["acme/app"] = true
	fake.Issues["acme/app"] = []map[string]any{
		{"number": 1, "title": "Bug", "state": "open", "labels": []map[string]any{{"name": "bug", "color": "d73a4a"}}, "user": map[string]any{"login": "bob"}},
		{"number": 2, "title": "A PR", "state": "open", "pull_request": map[string]any{"url": "x"}},
	}
	ctx := context.Background()
	for range 3 {
		issues, err := fake.Client.Issues(ctx, "acme/app", "open")
		if err != nil {
			t.Fatal(err)
		}
		if len(issues) != 1 || issues[0].Number != 1 || issues[0].Labels[0].Name != "bug" {
			t.Fatalf("issues = %+v", issues)
		}
	}
	if calls, _ := fake.IssueCalls(); calls != 1 {
		t.Fatalf("want one request inside the cache window, got %d", calls)
	}
}

func TestIssuesRevalidateWithTheETagAfterAMinute(t *testing.T) {
	fake := githubapptest.New(t)
	fake.Installed["acme/app"] = true
	fake.Issues["acme/app"] = []map[string]any{{"number": 5, "title": "Old", "state": "open"}}
	now := time.Now()
	githubapp.SetClock(fake.Client, func() time.Time { return now })
	ctx := context.Background()
	if _, err := fake.Client.Issues(ctx, "acme/app", "open"); err != nil {
		t.Fatal(err)
	}
	now = now.Add(61 * time.Second)
	issues, err := fake.Client.Issues(ctx, "acme/app", "open")
	if err != nil || len(issues) != 1 || issues[0].Number != 5 {
		t.Fatalf("revalidated issues = %+v, %v", issues, err)
	}
	if calls, notModified := fake.IssueCalls(); calls != 2 || notModified != 1 {
		t.Fatalf("calls = %d, 304s = %d", calls, notModified)
	}
}
