// Package githubapptest is an in-memory GitHub API for tests: it verifies the
// App JWT, mints installation tokens, answers collaborator permissions and
// serves issues with an ETag.
package githubapptest

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/FacileStudio/Opus/apps/api/internal/githubapp"
	"github.com/go-jose/go-jose/v4"
)

// ClientID is the App client id the fake expects as the JWT issuer.
const ClientID = "Iv1.test"

// Fake is the GitHub stand-in. Set its maps before the calls under test.
type Fake struct {
	Server *httptest.Server
	Client *githubapp.Client

	Installed   map[string]bool
	Logins      map[string]string
	Permissions map[string]string
	Issues      map[string][]map[string]any

	key         *rsa.PrivateKey
	mu          sync.Mutex
	scopes      []map[string]any
	issueCalls  int
	notModified int
}

// New starts the fake and a client pointed at it.
func New(t *testing.T) *Fake {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	fake := &Fake{
		Installed:   map[string]bool{},
		Logins:      map[string]string{},
		Permissions: map[string]string{},
		Issues:      map[string][]map[string]any{},
		key:         key,
	}
	fake.Server = httptest.NewServer(fake.routes())
	t.Cleanup(fake.Server.Close)
	pemKey := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	client, err := githubapp.New(ClientID, base64.StdEncoding.EncodeToString(pemKey))
	if err != nil {
		t.Fatal(err)
	}
	fake.Client = client.WithBaseURL(fake.Server.URL)
	return fake
}

// Scopes returns the body of every token mint, in order.
func (f *Fake) Scopes() []map[string]any {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]map[string]any(nil), f.scopes...)
}

// IssueCalls returns how many issue requests reached the fake, and how many
// of them it answered 304.
func (f *Fake) IssueCalls() (int, int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.issueCalls, f.notModified
}

func (f *Fake) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /repos/{owner}/{repo}/installation", f.installation)
	mux.HandleFunc("POST /app/installations/{id}/access_tokens", f.mint)
	mux.HandleFunc("GET /repos/{owner}/{repo}", f.withToken(f.repository))
	mux.HandleFunc("GET /user/{id}", f.withToken(f.user))
	mux.HandleFunc("GET /repos/{owner}/{repo}/collaborators/{login}/permission", f.withToken(f.permission))
	mux.HandleFunc("GET /repos/{owner}/{repo}/issues", f.withToken(f.issues))
	return mux
}

func repoName(r *http.Request) string {
	return strings.ToLower(r.PathValue("owner") + "/" + r.PathValue("repo"))
}

func (f *Fake) validJWT(r *http.Request) bool {
	raw := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	signed, err := jose.ParseSignedCompact(raw, []jose.SignatureAlgorithm{jose.RS256})
	if err != nil {
		return false
	}
	payload, err := signed.Verify(&f.key.PublicKey)
	if err != nil {
		return false
	}
	var claims struct {
		Iss string `json:"iss"`
		Exp int64  `json:"exp"`
	}
	return json.Unmarshal(payload, &claims) == nil && claims.Iss == ClientID && claims.Exp > time.Now().Unix()
}

func (f *Fake) installation(w http.ResponseWriter, r *http.Request) {
	if !f.validJWT(r) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if !f.Installed[repoName(r)] {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	writeJSON(w, map[string]any{"id": 7})
}

func (f *Fake) mint(w http.ResponseWriter, r *http.Request) {
	if !f.validJWT(r) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	f.mu.Lock()
	f.scopes = append(f.scopes, body)
	f.mu.Unlock()
	writeJSON(w, map[string]any{"token": "ghs_fake", "expires_at": time.Now().Add(time.Hour).UTC().Format(time.RFC3339)})
}

func (f *Fake) withToken(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer ghs_fake" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func (f *Fake) repository(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{"full_name": r.PathValue("owner") + "/" + r.PathValue("repo"), "default_branch": "trunk"})
}

func (f *Fake) user(w http.ResponseWriter, r *http.Request) {
	login, ok := f.Logins[r.PathValue("id")]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	writeJSON(w, map[string]any{"login": login})
}

func (f *Fake) permission(w http.ResponseWriter, r *http.Request) {
	level, ok := f.Permissions[r.PathValue("login")]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	writeJSON(w, map[string]any{"permission": level})
}

func (f *Fake) issues(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	f.issueCalls++
	if r.Header.Get("If-None-Match") == `"v1"` {
		f.notModified++
		f.mu.Unlock()
		w.WriteHeader(http.StatusNotModified)
		return
	}
	f.mu.Unlock()
	w.Header().Set("ETag", `"v1"`)
	writeJSON(w, f.Issues[repoName(r)])
}

func writeJSON(w http.ResponseWriter, value any) {
	encoded, err := json.Marshal(value)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(encoded); err != nil {
		return
	}
}
