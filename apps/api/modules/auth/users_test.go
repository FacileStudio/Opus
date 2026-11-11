package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"slices"
	"testing"

	"github.com/FacileStudio/Opus/apps/api/internal/database/dbtest"
	"github.com/FacileStudio/porte"
	tronc "github.com/FacileStudio/tronc/errors"
)

func newTestUserStore(t *testing.T) (*UserStore, *sql.DB) {
	t.Helper()
	sqlDB, err := dbtest.Open(t).DB()
	if err != nil {
		t.Fatalf("sql handle: %v", err)
	}
	store, err := NewUserStore(context.Background(), sqlDB, nil)
	if err != nil {
		t.Fatalf("NewUserStore: %v", err)
	}
	return store, sqlDB
}

func signIn(t *testing.T, store *UserStore, claims porte.Claims) int64 {
	t.Helper()
	id, err := store.UpsertFromOIDC(context.Background(), claims)
	if err != nil {
		t.Fatalf("UpsertFromOIDC: %v", err)
	}
	return id
}

func TestUserStoreLeavesOnlyColorAndGitHubUnwritten(t *testing.T) {
	store, _ := newTestUserStore(t)
	if got, want := store.Unwritten(), []string{"color", "github_id"}; !slices.Equal(got, want) {
		t.Fatalf("Unwritten() = %v, want %v", got, want)
	}
}

func TestNewAccountFillsEveryColumnOldAccountsHave(t *testing.T) {
	store, db := newTestUserStore(t)
	id := signIn(t, store, porte.Claims{
		Provider: "https://sso.facile.studio", Subject: "sub-new", Email: "New@Facile.Studio", EmailVerified: true,
		Name: "Nouveau", Picture: "https://sso.facile.studio/avatars/sub-new", Locale: "fr",
	})
	var nulls int
	err := db.QueryRow(`SELECT (email IS NULL)::int + (email_verified IS NULL)::int + (name IS NULL)::int +
		(image IS NULL)::int + (locale IS NULL)::int + (created_at IS NULL)::int + (updated_at IS NULL)::int
		FROM users WHERE id = $1`, id).Scan(&nulls)
	if err != nil || nulls != 0 {
		t.Fatalf("NULL columns on a new account = %d, %v, want 0", nulls, err)
	}
}

func TestUserStoreLinksOnlyByVerifiedEmail(t *testing.T) {
	store, db := newTestUserStore(t)
	var existing int64
	if err := db.QueryRow(`INSERT INTO users (email, name) VALUES ('ines@facile.studio', 'Ines') RETURNING id`).Scan(&existing); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	claims := porte.Claims{Provider: "https://sso.facile.studio", Subject: "sub-ines", Email: "Ines@Facile.Studio"}
	if _, err := store.UpsertFromOIDC(context.Background(), claims); tronc.Status(err) != 409 {
		t.Fatalf("unverified email status = %d (%v), want 409", tronc.Status(err), err)
	}
	claims.EmailVerified = true
	if linked := signIn(t, store, claims); linked != existing {
		t.Fatalf("verified email linked to %d, want %d", linked, existing)
	}
}

func TestUserStoreStoresGitHubIDClaim(t *testing.T) {
	store, db := newTestUserStore(t)
	id := signIn(t, store, porte.Claims{
		Provider:      "https://sso.facile.studio",
		Subject:       "sub-gh",
		Email:         "gh@facile.studio",
		EmailVerified: true,
		Extra:         map[string]json.RawMessage{"connections": json.RawMessage(`{"github":"987654"}`)},
	})
	var githubID sql.NullString
	if err := db.QueryRow(`SELECT github_id FROM users WHERE id = $1`, id).Scan(&githubID); err != nil {
		t.Fatalf("load user: %v", err)
	}
	if githubID.String != "987654" {
		t.Fatalf("github_id = %q, want 987654", githubID.String)
	}
}
