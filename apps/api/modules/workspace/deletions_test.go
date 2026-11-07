package workspace

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/FacileStudio/Opus/apps/api/internal/storage"
	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/porte/directory"

	"gorm.io/gorm"
)

type erasureFixture struct {
	alice, bob                int64
	shared, alicePersonal     int64
	bobPersonal, sharedTask   int64
	personalAsset, aliceTaskN int64
}

func seedErasure(t *testing.T, service *Service, db *gorm.DB) erasureFixture {
	t.Helper()
	f := erasureFixture{}
	f.alice = createUser(t, db, "alice@facile.studio", "Alice")
	addIdentity(t, db, f.alice, "sub-alice")
	f.bob = createUser(t, db, "bob@facile.studio", "Bob")
	addIdentity(t, db, f.bob, "sub-bob")
	for _, id := range []int64{f.alice, f.bob} {
		if err := service.ensurePersonal(context.Background(), id); err != nil {
			t.Fatalf("personal workspace: %v", err)
		}
	}
	db.Model(&schemas.Workspace{}).Where("personal_owner_id = ?", f.alice).Select("id").Scan(&f.alicePersonal)
	db.Model(&schemas.Workspace{}).Where("personal_owner_id = ?", f.bob).Select("id").Scan(&f.bobPersonal)
	f.shared = createWorkspace(t, db, "Studio")
	addMember(t, db, f.shared, f.alice, "member")
	addMember(t, db, f.shared, f.bob, "owner")
	seedErasureContent(t, db, &f)
	return f
}

func seedErasureContent(t *testing.T, db *gorm.DB, f *erasureFixture) {
	t.Helper()
	now := time.Now()
	project := func(workspaceID int64, slug string) int64 {
		p := schemas.Project{WorkspaceID: workspaceID, Slug: slug, Name: slug, CreatedAt: now, UpdatedAt: now}
		if err := db.Create(&p).Error; err != nil {
			t.Fatalf("create project: %v", err)
		}
		return p.ID
	}
	task := func(workspaceID, projectID int64, title string) int64 {
		k := schemas.Task{WorkspaceID: workspaceID, ProjectID: projectID, Title: title, CreatedAt: now, UpdatedAt: now}
		if err := db.Create(&k).Error; err != nil {
			t.Fatalf("create task: %v", err)
		}
		return k.ID
	}
	sharedProject := project(f.shared, "studio-board")
	f.sharedTask = task(f.shared, sharedProject, "Ship it")
	f.aliceTaskN = task(f.alicePersonal, project(f.alicePersonal, "alice-notes"), "Private note")
	task(f.bobPersonal, project(f.bobPersonal, "bob-notes"), "Bob private")
	for _, c := range []schemas.Comment{
		{TaskID: f.sharedTask, UserID: f.alice, Content: "from alice", CreatedAt: now, UpdatedAt: now},
		{TaskID: f.sharedTask, UserID: f.bob, Content: "from bob", CreatedAt: now, UpdatedAt: now},
	} {
		if err := db.Create(&c).Error; err != nil {
			t.Fatalf("create comment: %v", err)
		}
	}
	link := schemas.ShareLink{ProjectID: sharedProject, Token: "tok", CreatedBy: f.alice, IsActive: true, CreatedAt: now, UpdatedAt: now}
	if err := db.Create(&link).Error; err != nil {
		t.Fatalf("create share link: %v", err)
	}
	invite := schemas.Invitation{WorkspaceID: f.shared, Email: "Alice@Facile.Studio", Role: "member", Status: "pending", InviterID: f.bob, CreatedAt: now}
	if err := db.Create(&invite).Error; err != nil {
		t.Fatalf("create invitation: %v", err)
	}
}

func count(t *testing.T, db *gorm.DB, query string, args ...any) int64 {
	t.Helper()
	var n int64
	if err := db.Raw(query, args...).Scan(&n).Error; err != nil {
		t.Fatalf("count %q: %v", query, err)
	}
	return n
}

func TestErasePersonKeepsOthersAndIsIdempotent(t *testing.T) {
	service, db := newPostgresService(t)
	ctx := context.Background()
	f := seedErasure(t, service, db)
	assets := &fakeAssets{}
	deletions := NewDeletions(service, directoryClient(), func(context.Context, string, string) error { return nil }, assets, nopLogger())

	for range 2 {
		if err := deletions.ErasePerson(ctx, testIssuer, "sub-alice"); err != nil {
			t.Fatalf("ErasePerson: %v", err)
		}
	}

	gone := map[string]int64{
		"account":        count(t, db, "SELECT count(*) FROM users WHERE id = ?", f.alice),
		"identity":       count(t, db, "SELECT count(*) FROM porte_identities WHERE subject = 'sub-alice'"),
		"membership":     count(t, db, "SELECT count(*) FROM workspace_members WHERE user_id = ?", f.alice),
		"personal space": count(t, db, "SELECT count(*) FROM workspaces WHERE id = ?", f.alicePersonal),
		"personal task":  count(t, db, "SELECT count(*) FROM tasks WHERE id = ?", f.aliceTaskN),
		"invitation":     count(t, db, "SELECT count(*) FROM invitations WHERE lower(email) = 'alice@facile.studio'"),
	}
	for name, n := range gone {
		if n != 0 {
			t.Errorf("%s: %d rows left, want 0", name, n)
		}
	}
	checkErasureSurvivors(t, db, f)
}

func checkErasureSurvivors(t *testing.T, db *gorm.DB, f erasureFixture) {
	t.Helper()
	kept := map[string]int64{
		"bob account":        count(t, db, "SELECT count(*) FROM users WHERE id = ?", f.bob),
		"bob identity":       count(t, db, "SELECT count(*) FROM porte_identities WHERE subject = 'sub-bob'"),
		"bob personal space": count(t, db, "SELECT count(*) FROM workspaces WHERE id = ?", f.bobPersonal),
		"bob personal task":  count(t, db, "SELECT count(*) FROM tasks WHERE workspace_id = ?", f.bobPersonal),
		"shared workspace":   count(t, db, "SELECT count(*) FROM workspaces WHERE id = ?", f.shared),
		"shared task":        count(t, db, "SELECT count(*) FROM tasks WHERE id = ?", f.sharedTask),
		"both comments":      count(t, db, "SELECT count(*) FROM comments WHERE task_id = ?", f.sharedTask),
		"anonymous comment":  count(t, db, "SELECT count(*) FROM comments WHERE content = 'from alice' AND user_id IS NULL"),
		"bob comment":        count(t, db, "SELECT count(*) FROM comments WHERE content = 'from bob' AND user_id = ?", f.bob),
		"share link":         count(t, db, "SELECT count(*) FROM share_links WHERE token = 'tok' AND created_by IS NULL"),
		"bob membership":     count(t, db, "SELECT count(*) FROM workspace_members WHERE user_id = ?", f.bob),
	}
	want := map[string]int64{"both comments": 2, "bob membership": 2}
	for name, n := range kept {
		expected, ok := want[name]
		if !ok {
			expected = 1
		}
		if n != expected {
			t.Errorf("%s: %d rows, want %d", name, n, expected)
		}
	}
}

func TestErasePersonRemovesPersonalAssetFiles(t *testing.T) {
	service, db := newPostgresService(t)
	f := seedErasure(t, service, db)
	key := "ws/alice/file.png"
	path := filepath.Join(service.uploadDir, "assets", key)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatalf("write asset: %v", err)
	}
	err := db.Exec(`INSERT INTO asset (id, workspace_id, project_id, task_id, object_key, filename, mime_type, size, created_by)
		SELECT 'a1', ?, project_id, id, ?, 'file.png', 'image/png', 1, ? FROM tasks WHERE id = ?`,
		f.alicePersonal, key, f.alice, f.aliceTaskN).Error
	if err != nil {
		t.Fatalf("seed asset: %v", err)
	}
	assets := storage.NewAssetStorage(service.uploadDir, nil)
	deletions := NewDeletions(service, directoryClient(), func(context.Context, string, string) error { return nil }, assets, nopLogger())
	if err := deletions.ErasePerson(context.Background(), testIssuer, "sub-alice"); err != nil {
		t.Fatalf("ErasePerson: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("asset file still on disk: %v", err)
	}
}

func TestDeletionStateRoundTrips(t *testing.T) {
	_, db := newPostgresService(t)
	ctx := context.Background()
	state := deletionState{orm: db}

	if got, err := state.Load(ctx); err != nil || got != nil {
		t.Fatalf("Load before any save = %q, %v, want nil", got, err)
	}
	for _, blob := range []string{`{"a":1}`, `{"a":2}`} {
		if err := state.Save(ctx, []byte(blob)); err != nil {
			t.Fatalf("Save: %v", err)
		}
	}
	if got, err := state.Load(ctx); err != nil || string(got) != `{"a":2}` {
		t.Fatalf("Load = %q, %v, want the last save", got, err)
	}
}

func nopLogger() *slog.Logger { return slog.New(slog.DiscardHandler) }

func directoryClient() directory.Client { return directory.Client{Issuer: testIssuer} }

type fakeAssets struct {
	deleted []string
	fail    error
}

func (f *fakeAssets) Delete(_ context.Context, keys []string) error {
	if f.fail != nil {
		return f.fail
	}
	f.deleted = append(f.deleted, keys...)
	return nil
}

type assetSeed struct {
	id, key                     string
	workspaceID, taskID, author int64
}

func seedAsset(t *testing.T, db *gorm.DB, a assetSeed) {
	t.Helper()
	err := db.Exec(`INSERT INTO asset (id, workspace_id, project_id, task_id, object_key, filename, mime_type, size, created_by)
		SELECT ?, ?, project_id, id, ?, 'file.png', 'image/png', 1, ? FROM tasks WHERE id = ?`,
		a.id, a.workspaceID, a.key, a.author, a.taskID).Error
	if err != nil {
		t.Fatalf("seed asset: %v", err)
	}
}

func TestErasePersonDeletesPersonalObjectsOnly(t *testing.T) {
	service, db := newPostgresService(t)
	f := seedErasure(t, service, db)
	seedAsset(t, db, assetSeed{"a1", "workspace/personal/one.png", f.alicePersonal, f.aliceTaskN, f.alice})
	seedAsset(t, db, assetSeed{"a2", "workspace/personal/two.png", f.alicePersonal, f.aliceTaskN, f.alice})
	seedAsset(t, db, assetSeed{"a3", "workspace/shared/three.png", f.shared, f.sharedTask, f.alice})
	assets := &fakeAssets{}
	deletions := NewDeletions(service, directoryClient(), func(context.Context, string, string) error { return nil }, assets, nopLogger())

	for range 2 {
		if err := deletions.ErasePerson(context.Background(), testIssuer, "sub-alice"); err != nil {
			t.Fatalf("ErasePerson: %v", err)
		}
	}
	slices.Sort(assets.deleted)
	want := []string{"workspace/personal/one.png", "workspace/personal/two.png"}
	if !slices.Equal(assets.deleted, want) {
		t.Fatalf("deleted objects = %v, want %v", assets.deleted, want)
	}
	if n := count(t, db, "SELECT count(*) FROM asset WHERE id = 'a3'"); n != 1 {
		t.Fatalf("shared asset rows = %d, want 1", n)
	}
}

func TestErasePersonKeepsAccountWhenStorageFailsThenRetries(t *testing.T) {
	service, db := newPostgresService(t)
	f := seedErasure(t, service, db)
	seedAsset(t, db, assetSeed{"a1", "workspace/personal/one.png", f.alicePersonal, f.aliceTaskN, f.alice})
	assets := &fakeAssets{fail: errors.New("bucket unreachable")}
	deletions := NewDeletions(service, directoryClient(), func(context.Context, string, string) error { return nil }, assets, nopLogger())

	if err := deletions.ErasePerson(context.Background(), testIssuer, "sub-alice"); err == nil {
		t.Fatal("a storage failure must fail the erasure")
	}
	if n := count(t, db, "SELECT count(*) FROM users WHERE id = ?", f.alice); n != 1 {
		t.Fatalf("account rows after a failed erasure = %d, want 1", n)
	}
	assets.fail = nil
	if err := deletions.ErasePerson(context.Background(), testIssuer, "sub-alice"); err != nil {
		t.Fatalf("retry: %v", err)
	}
	if n := count(t, db, "SELECT count(*) FROM users WHERE id = ?", f.alice); n != 0 || len(assets.deleted) != 1 {
		t.Fatalf("after the retry: %d accounts, %d objects deleted", n, len(assets.deleted))
	}
}
