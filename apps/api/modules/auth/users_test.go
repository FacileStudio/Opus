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
	// The store resolves identities by joining porte/pg's table; the test
	// creates the minimal shape so the upsert path is exercised.
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

	// With the identity row present, the store resolves through the JOIN
	// without the email fallback.
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