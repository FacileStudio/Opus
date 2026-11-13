package main

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/Opus/apps/api/internal/database/dbtest"
	"github.com/FacileStudio/Opus/apps/api/internal/env"
	"github.com/FacileStudio/porte"
)

// signingIssuer is a provider that publishes one RSA key, so tokens it signs
// pass porte's verifier the way Registre's do.
type signingIssuer struct {
	url string
	key *rsa.PrivateKey
}

func newSigningIssuer(t *testing.T) *signingIssuer {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("key: %v", err)
	}
	issuer := &signingIssuer{key: key}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(issuer.document(r.URL.Path)); err != nil {
			t.Errorf("issuer: %v", err)
		}
	}))
	t.Cleanup(server.Close)
	issuer.url = server.URL
	return issuer
}

// document is the JWKS on /jwks and the discovery document everywhere else.
func (s *signingIssuer) document(path string) map[string]any {
	if path == "/jwks" {
		return map[string]any{"keys": []map[string]string{{
			"kty": "RSA", "alg": "RS256", "use": "sig", "kid": "k1",
			"n": base64.RawURLEncoding.EncodeToString(s.key.N.Bytes()),
			"e": base64.RawURLEncoding.EncodeToString(big.NewInt(int64(s.key.E)).Bytes()),
		}}}
	}
	return map[string]any{
		"issuer": s.url, "authorization_endpoint": s.url + "/authorize",
		"token_endpoint": s.url + "/token", "jwks_uri": s.url + "/jwks",
		"id_token_signing_alg_values_supported": []string{"RS256"},
	}
}

// logoutToken signs a Back-Channel Logout 1.0 token for subject, addressed to
// the opus client.
func (s *signingIssuer) logoutToken(t *testing.T, subject string) string {
	t.Helper()
	now := time.Now()
	header, _ := json.Marshal(map[string]string{"alg": "RS256", "typ": "logout+jwt", "kid": "k1"})
	payload, _ := json.Marshal(map[string]any{
		"iss": s.url, "aud": "opus", "sub": subject, "jti": subject + "-logout",
		"iat": now.Unix(), "exp": now.Add(time.Minute).Unix(),
		"events": map[string]any{"http://schemas.openid.net/event/backchannel-logout": map[string]any{}},
	})
	signing := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload)
	digest := sha256.Sum256([]byte(signing))
	signature, err := rsa.SignPKCS1v15(rand.Reader, s.key, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return signing + "." + base64.RawURLEncoding.EncodeToString(signature)
}

// backchannelRig is Opus's real router and auth stores over a test database,
// federated to a signing issuer.
type backchannelRig struct {
	idp    *signingIssuer
	kits   *authKits
	router http.Handler
}

func newBackchannelRig(t *testing.T) *backchannelRig {
	t.Helper()
	db := dbtest.Open(t)
	idp := newSigningIssuer(t)
	logger := slog.New(slog.DiscardHandler)
	appEnv := &env.Config{UploadDir: t.TempDir(), OIDC: &env.OIDCConfig{
		Issuer: idp.url, ClientID: "opus", ClientSecret: "s",
		RedirectURL: "http://localhost/api/auth/oidc/callback", SuccessURL: "http://localhost/",
	}}
	kits, err := buildAuth(context.Background(), db, appEnv, logger)
	if err != nil {
		t.Fatalf("auth: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("sql: %v", err)
	}
	router := buildRouter(routerDeps{db: db, sqlDB: sqlDB, appEnv: appEnv, appLogger: logger, kits: kits, ideas: &ideas{}})
	return &backchannelRig{idp: idp, kits: kits, router: router}
}

// signIn creates subject's account and identity, as the login callback does,
// and returns a session token for it.
func (r *backchannelRig) signIn(t *testing.T, subject string) string {
	t.Helper()
	ctx := context.Background()
	claims := porte.Claims{Provider: r.idp.url, Subject: subject, Email: subject + "@facile.studio", EmailVerified: true, Name: subject}
	userID, err := r.kits.users.UpsertFromOIDC(ctx, claims)
	if err != nil {
		t.Fatalf("user %s: %v", subject, err)
	}
	if err := r.kits.store.Identities().Save(ctx, porte.StoredIdentity{UserID: userID, Provider: r.idp.url, Subject: subject}); err != nil {
		t.Fatalf("identity %s: %v", subject, err)
	}
	token, _, err := r.kits.sessions.IssueFor(ctx, porte.Session{UserID: userID, Provider: r.idp.url, Subject: subject})
	if err != nil {
		t.Fatalf("session %s: %v", subject, err)
	}
	return token
}

// serve sends request through the router and returns the recorded response.
func (r *backchannelRig) serve(request *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	r.router.ServeHTTP(recorder, request)
	return recorder
}

// me answers the status GET /api/auth/me gives token.
func (r *backchannelRig) me(token string) int {
	request := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	return r.serve(request).Code
}

// Signing out of Registre must sign the person out of Opus: a logout token
// for their subject, posted to the route the router mounts under /api,
// revokes every session they hold here and nobody else's.
func TestBackchannelLogoutRevokesTheSubjectsSessions(t *testing.T) {
	rig := newBackchannelRig(t)
	alice, bob := rig.signIn(t, "alice"), rig.signIn(t, "bob")
	if rig.me(alice) != http.StatusOK || rig.me(bob) != http.StatusOK {
		t.Fatal("the sessions do not authenticate before the logout")
	}

	form := url.Values{"logout_token": {rig.idp.logoutToken(t, "alice")}}
	request := httptest.NewRequest(http.MethodPost, "/api"+porte.RouteBackchannelLogout, strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if recorder := rig.serve(request); recorder.Code != http.StatusOK {
		t.Fatalf("backchannel logout = %d %s, want 200", recorder.Code, recorder.Body)
	}

	if code := rig.me(alice); code != http.StatusUnauthorized {
		t.Errorf("the logged-out session answers %d, want 401", code)
	}
	if code := rig.me(bob); code != http.StatusOK {
		t.Errorf("another person's session answers %d, want 200", code)
	}
}
