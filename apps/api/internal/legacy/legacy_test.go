package legacy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/FacileStudio/porte"
	"github.com/go-chi/chi/v5"
)

type fakeAuthenticator struct{}

func (fakeAuthenticator) Optional(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity := porte.Identity{UserID: 7, SessionID: 1}
		next.ServeHTTP(w, r.WithContext(porte.WithIdentity(r.Context(), identity)))
	})
}

type fakeIdentityResolver struct{}

func (fakeIdentityResolver) IdentityForUser(_ context.Context, userID int64) (int64, string, error) {
	return userID, "camille@facile.studio", nil
}

func TestMountProxyInjectsBridgeHeaders(t *testing.T) {
	var gotUserID, gotEmail, gotAuth string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID = r.Header.Get(userIDHeader)
		gotEmail = r.Header.Get(userEmailHeader)
		gotAuth = r.Header.Get(authHeader)
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	remote, err := url.Parse(upstream.URL)
	if err != nil {
		t.Fatalf("parse upstream: %v", err)
	}
	router := chi.NewRouter()
	MountProxy(router, remote, fakeAuthenticator{}, fakeIdentityResolver{}, "hunter2")

	request := httptest.NewRequest("GET", "/api/task/1", nil)
	request.Header.Set("X-Opus-User-Id", "999")
	request.Header.Set("X-Opus-Auth", "forged")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if gotUserID != "7" {
		t.Fatalf("user id = %q, want the porte-authenticated 7", gotUserID)
	}
	if gotEmail != "camille@facile.studio" {
		t.Fatalf("email = %q, want the resolver's address", gotEmail)
	}
	if gotAuth != tag("hunter2", "7") {
		t.Fatalf("auth tag = %q, want the HMAC over the user id", gotAuth)
	}
}

func TestMountProxySkipsBridgeWithoutSecret(t *testing.T) {
	var gotHeaders []string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeaders = []string{r.Header.Get(userIDHeader), r.Header.Get(authHeader)}
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	remote, err := url.Parse(upstream.URL)
	if err != nil {
		t.Fatalf("parse upstream: %v", err)
	}
	router := chi.NewRouter()
	MountProxy(router, remote, fakeAuthenticator{}, fakeIdentityResolver{}, "")

	request := httptest.NewRequest("GET", "/api/task/1", nil)
	request.Header.Set("X-Opus-User-Id", "999")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if gotHeaders[0] != "" || gotHeaders[1] != "" {
		t.Fatalf("bridge headers forwarded without a secret: %v", gotHeaders)
	}
}

func TestMountProxyPassesThroughUnauthenticated(t *testing.T) {
	var gotUserID string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID = r.Header.Get(userIDHeader)
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	remote, err := url.Parse(upstream.URL)
	if err != nil {
		t.Fatalf("parse upstream: %v", err)
	}
	router := chi.NewRouter()
	MountProxy(router, remote, noopAuthenticator{}, fakeIdentityResolver{}, "hunter2")

	request := httptest.NewRequest("GET", "/api/config", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want the upstream's 200", response.Code)
	}
	if gotUserID != "" {
		t.Fatalf("anonymous request forwarded a user id %q", gotUserID)
	}
}

type noopAuthenticator struct{}

func (noopAuthenticator) Optional(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}