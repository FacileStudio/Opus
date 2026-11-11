package auth

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/FacileStudio/Opus/apps/api/internal/database/dbtest"
	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/porte"
	tronc "github.com/FacileStudio/tronc/errors"
)

func newTestUserStore(t *testing.T) *UserStore {
	t.Helper()
	db := dbtest.Open(t)
	return NewUserStore(db)
}

func createUser(t *testing.T, store *UserStore, email, name string) int64 {
	t.Helper()
	user, err := store.create(context.Background(), email, name, "")
	if err != nil {
		t.Fatalf("create %s: %v", email, err)
	}
	return user.ID
}

func TestUserStoreLinksOnlyByVerifiedEmail(t *testing.T) {
	store := newTestUserStore(t)
	ctx := context.Background()
	existing := createUser(t, store, "ines@facile.studio", "Ines")
	claims := porte.Claims{Provider: "https://sso.facile.studio", Subject: "sub-ines", Email: "Ines@Facile.Studio"}

	if _, err := store.UpsertFromOIDC(ctx, claims); tronc.Status(err) != 409 {
		t.Fatalf("unverified email status = %d (%v), want 409", tronc.Status(err), err)
	}

	claims.EmailVerified = true
	linked, err := store.UpsertFromOIDC(ctx, claims)
	if err != nil {
		t.Fatalf("verified email: %v", err)
	}
	if linked != existing {
		t.Fatalf("verified email linked to %d, want %d", linked, existing)
	}
}

func TestUserStoreRefusesASecondSubjectOfTheSameIssuer(t *testing.T) {
	store := newTestUserStore(t)
	ctx := context.Background()
	const issuer = "https://sso.facile.studio"
	owner := createUser(t, store, "noe@facile.studio", "Noe")
	if err := store.orm.Exec("INSERT INTO porte_identities (user_id, provider, subject) VALUES (?, ?, 'sub-owner')", owner, issuer).Error; err != nil {
		t.Fatalf("seed identity: %v", err)
	}

	_, err := store.UpsertFromOIDC(ctx, porte.Claims{Provider: issuer, Subject: "sub-newcomer", Email: "noe@facile.studio", EmailVerified: true})
	if tronc.Status(err) != 409 {
		t.Fatalf("second subject status = %d (%v), want 409", tronc.Status(err), err)
	}

	other, err := store.UpsertFromOIDC(ctx, porte.Claims{Provider: "https://elsewhere.example", Subject: "sub-x", Email: "noe@facile.studio", EmailVerified: true})
	if err != nil || other != owner {
		t.Fatalf("another issuer = %d, %v, want link to %d", other, err, owner)
	}
}

func TestUserStoreStoresTheLocaleClaim(t *testing.T) {
	store := newTestUserStore(t)
	ctx := context.Background()
	claims := porte.Claims{Provider: "https://sso.facile.studio", Subject: "sub-loc", Email: "loc@facile.studio", EmailVerified: true, Locale: "de"}

	id, err := store.UpsertFromOIDC(ctx, claims)
	if err != nil {
		t.Fatalf("first sign-in: %v", err)
	}
	claims.Locale = "es"
	if _, err := store.UpsertFromOIDC(ctx, claims); err != nil {
		t.Fatalf("second sign-in: %v", err)
	}
	var user schemas.User
	if err := store.orm.First(&user, id).Error; err != nil {
		t.Fatalf("load user: %v", err)
	}
	if user.Locale != "es" {
		t.Fatalf("locale = %q, want es", user.Locale)
	}
}

func TestUserStoreStoresGitHubIDClaim(t *testing.T) {
	store := newTestUserStore(t)
	ctx := context.Background()
	claims := porte.Claims{
		Provider:      "https://sso.facile.studio",
		Subject:       "sub-gh",
		Email:         "gh@facile.studio",
		EmailVerified: true,
		Extra: map[string]json.RawMessage{
			"connections": json.RawMessage(`{"github":"987654"}`),
		},
	}

	id, err := store.UpsertFromOIDC(ctx, claims)
	if err != nil {
		t.Fatalf("upsert from oidc: %v", err)
	}
	var user schemas.User
	if err := store.orm.First(&user, id).Error; err != nil {
		t.Fatalf("load user: %v", err)
	}
	if user.GitHubID == nil || *user.GitHubID != "987654" {
		t.Fatalf("github_id = %v, want 987654", user.GitHubID)
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

	stale := createUser(t, store, "stale@facile.studio", "Stale")
	other := createUser(t, store, "other@facile.studio", "Other")
	local := createUser(t, store, "local@facile.studio", "Local")
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
