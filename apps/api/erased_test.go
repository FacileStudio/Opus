package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/FacileStudio/Opus/apps/api/internal/database/dbtest"
	"github.com/FacileStudio/Opus/apps/api/internal/env"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/oidc/jwt"
)

// discoveryIssuer serves the OIDC discovery document porte's kit needs to start.
func discoveryIssuer(t *testing.T) string {
	t.Helper()
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		err := json.NewEncoder(w).Encode(map[string]any{
			"issuer": server.URL, "authorization_endpoint": server.URL + "/authorize",
			"token_endpoint": server.URL + "/token", "jwks_uri": server.URL + "/jwks",
			"id_token_signing_alg_values_supported": []string{"RS256"},
		})
		if err != nil {
			t.Errorf("discovery: %v", err)
		}
	}))
	t.Cleanup(server.Close)
	return server.URL
}

// A subject the deletion receiver handled gets no account from porte's own
// write path: a bearer token minted before the erasure, as the first MCP
// call presents it, is refused and nothing is written.
func TestAHandledSubjectIsRefusedAnAccount(t *testing.T) {
	db := dbtest.Open(t)
	ctx := context.Background()
	issuer := discoveryIssuer(t)
	state := `{"done":{"sub-gone":"` + time.Now().UTC().Format(time.RFC3339) + `"}}`
	if err := db.Exec(`INSERT INTO directory_deletion_state (id, state, updated_at) VALUES (1, ?, now())`, state).Error; err != nil {
		t.Fatalf("seed state: %v", err)
	}
	appEnv := &env.Config{UploadDir: t.TempDir(), OIDC: &env.OIDCConfig{
		Issuer: issuer, ClientID: "opus", ClientSecret: "s",
		RedirectURL: "http://localhost/api/auth/oidc/callback", SuccessURL: "http://localhost/",
	}}
	kits, err := buildAuth(ctx, db, appEnv, slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("auth: %v", err)
	}
	verified := true
	_, err = kits.oidc.ProvisionBearer(ctx, jwt.Claims{
		Subject: "sub-gone", Email: "gone@facile.studio", Profile: jwt.Profile{EmailVerified: &verified},
	})
	if !errors.Is(err, porte.ErrSubjectErased) {
		t.Fatalf("ProvisionBearer = %v, want porte.ErrSubjectErased", err)
	}
	var n int64
	if err := db.Raw(`SELECT (SELECT count(*) FROM users) + (SELECT count(*) FROM porte_identities)`).Scan(&n).Error; err != nil || n != 0 {
		t.Errorf("a refused provisioning wrote %d rows (%v)", n, err)
	}
}
