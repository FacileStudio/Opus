package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/porte"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestUserStore(t *testing.T) *UserStore {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&schemas.User{}); err != nil {
		t.Fatalf("migrate user: %v", err)
	}
	if err := db.Exec("CREATE TABLE porte_identities (user_id integer, provider text, subject text)").Error; err != nil {
		t.Fatalf("create porte_identities: %v", err)
	}
	return NewUserStore(db)
}

func TestUserStoreCreateAndFind(t *testing.T) {
	store := newTestUserStore(t)
	ctx := context.Background()

	id, err := store.CreateFromPassword(ctx, "camille@facile.studio", "Camille")
	if err != nil {
		t.Fatalf("CreateFromPassword: %v", err)
	}
	if id == 0 {
		t.Fatal("CreateFromPassword returned id 0")
	}

	got, err := store.FindByEmail(ctx, "camille@facile.studio")
	if err != nil {
		t.Fatalf("FindByEmail: %v", err)
	}
	if got != id {
		t.Fatalf("FindByEmail = %d, want %d", got, id)
	}
}

func TestUserStoreFindByEmailMissing(t *testing.T) {
	store := newTestUserStore(t)
	_, err := store.FindByEmail(context.Background(), "nobody@facile.studio")
	if !errors.Is(err, porte.ErrNotFound) {
		t.Fatalf("FindByEmail on missing account = %v, want porte.ErrNotFound", err)
	}
}

func TestUserStoreUpsertFromOIDC(t *testing.T) {
	store := newTestUserStore(t)
	ctx := context.Background()

	id, err := store.UpsertFromOIDC(ctx, porte.Claims{
		Provider:      "authentik",
		Subject:       "user-123",
		Email:         "paul@facile.studio",
		EmailVerified: true,
		Name:          "Paul",
	})
	if err != nil {
		t.Fatalf("UpsertFromOIDC: %v", err)
	}

	again, err := store.UpsertFromOIDC(ctx, porte.Claims{
		Provider:      "authentik",
		Subject:       "user-123",
		Email:         "paul@facile.studio",
		EmailVerified: true,
		Name:          "Paul",
	})
	if err != nil {
		t.Fatalf("UpsertFromOIDC second call: %v", err)
	}
	if again != id {
		t.Fatalf("upsert changed the user id: %d -> %d", id, again)
	}

	if err := store.orm.Exec("INSERT INTO porte_identities (user_id, provider, subject) VALUES (?, 'authentik', 'user-123')", id).Error; err != nil {
		t.Fatalf("seed identity: %v", err)
	}
	direct, err := store.UpsertFromOIDC(ctx, porte.Claims{
		Provider:      "authentik",
		Subject:       "user-123",
		Email:         "paul@facile.studio",
		EmailVerified: true,
	})
	if err != nil {
		t.Fatalf("UpsertFromOIDC via identity: %v", err)
	}
	if direct != id {
		t.Fatalf("upsert via identity = %d, want %d", direct, id)
	}
}
func TestUserStoreStoresPictureClaim(t *testing.T) {
	store := newTestUserStore(t)
	ctx := context.Background()
	const picture = "https://sso.facile.studio/avatars/sub-1"

	cases := []struct {
		name  string
		claim string
		want  string
	}{
		{"registre picture url", picture, picture},
		{"data placeholder is not a picture", "data:image/svg+xml;base64,PHN2Zz4=", ""},
		{"plain http is dropped", "http://sso.facile.studio/avatars/sub-1", ""},
		{"a later sign-in restores it", picture, picture},
	}
	for _, tc := range cases {
		id, err := store.UpsertFromOIDC(ctx, porte.Claims{
			Provider:      "https://sso.facile.studio",
			Subject:       "sub-1",
			Email:         "lea@facile.studio",
			EmailVerified: true,
			Name:          "Lea",
			Picture:       tc.claim,
		})
		if err != nil {
			t.Fatalf("%s: UpsertFromOIDC: %v", tc.name, err)
		}
		var user schemas.User
		if err := store.orm.First(&user, id).Error; err != nil {
			t.Fatalf("%s: load user: %v", tc.name, err)
		}
		if user.Image != tc.want {
			t.Fatalf("%s: image = %q, want %q", tc.name, user.Image, tc.want)
		}
	}
}

func TestUserStoreBackfillPictures(t *testing.T) {
	store := newTestUserStore(t)
	ctx := context.Background()
	const issuer = "https://sso.facile.studio/"

	stale, err := store.CreateFromPassword(ctx, "stale@facile.studio", "Stale")
	if err != nil {
		t.Fatalf("create stale: %v", err)
	}
	other, err := store.CreateFromPassword(ctx, "other@facile.studio", "Other")
	if err != nil {
		t.Fatalf("create other: %v", err)
	}
	local, err := store.CreateFromPassword(ctx, "local@facile.studio", "Local")
	if err != nil {
		t.Fatalf("create local: %v", err)
	}
	if err := store.orm.Exec("UPDATE users SET image = 'https://porte.facile.studio/media/old.png' WHERE id = ?", stale).Error; err != nil {
		t.Fatalf("seed stale image: %v", err)
	}
	if err := store.orm.Exec(
		"INSERT INTO porte_identities (user_id, provider, subject) VALUES (?, ?, 'sub-stale'), (?, 'https://elsewhere.example', 'sub-other')",
		stale, issuer, other,
	).Error; err != nil {
		t.Fatalf("seed identities: %v", err)
	}

	updated, err := store.BackfillPictures(ctx, issuer)
	if err != nil {
		t.Fatalf("BackfillPictures: %v", err)
	}
	if updated != 1 {
		t.Fatalf("BackfillPictures updated %d rows, want 1", updated)
	}
	again, err := store.BackfillPictures(ctx, issuer)
	if err != nil {
		t.Fatalf("BackfillPictures again: %v", err)
	}
	if again != 0 {
		t.Fatalf("second BackfillPictures updated %d rows, want 0", again)
	}

	want := map[int64]string{
		stale: "https://sso.facile.studio/avatars/sub-stale",
		other: "",
		local: "",
	}
	for id, image := range want {
		var user schemas.User
		if err := store.orm.First(&user, id).Error; err != nil {
			t.Fatalf("load user %d: %v", id, err)
		}
		if user.Image != image {
			t.Fatalf("user %d image = %q, want %q", id, user.Image, image)
		}
	}
}
