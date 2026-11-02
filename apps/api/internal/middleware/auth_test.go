package middleware_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/FacileStudio/Opus/apps/api/internal/authcontext"
	"github.com/FacileStudio/Opus/apps/api/internal/middleware"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/session"
)

const (
	readToken   = "eyJhbGciOiJSUzI1NiJ9.eyJzY29wZSI6InJlYWQifQ.c2ln"
	writeToken  = "eyJhbGciOiJSUzI1NiJ9.eyJzY29wZSI6IndyaXRlIn0.c2ln"
	opaqueToken = "opaque-session-token-without-dots"
)

type stubVerifier struct{}

func (stubVerifier) VerifyJWT(_ context.Context, raw string) (porte.Identity, error) {
	identity := porte.Identity{UserID: 7, Actor: "bulle"}
	switch raw {
	case readToken:
		identity.Scopes = []string{"opus.read"}
	case writeToken:
		identity.Scopes = []string{"opus.write"}
	}
	return identity, nil
}

type stubSessions struct{ hash string }

func (s stubSessions) Create(context.Context, porte.Session) (porte.Session, error) {
	return porte.Session{}, nil
}

func (s stubSessions) Find(_ context.Context, tokenHash string) (porte.Session, error) {
	if tokenHash == s.hash {
		return porte.Session{ID: 1, UserID: 7, TokenHash: tokenHash, ExpiresAt: time.Now().Add(time.Hour)}, nil
	}
	return porte.Session{}, porte.ErrNotFound
}

func (s stubSessions) Touch(context.Context, string, time.Time) error { return nil }
func (s stubSessions) Delete(context.Context, string) error           { return nil }
func (s stubSessions) DeleteByUser(context.Context, int64) (int64, error) {
	return 0, nil
}
func (s stubSessions) DeleteLogins(context.Context, int64, int64) (int64, error) {
	return 0, nil
}
func (s stubSessions) ListByUser(context.Context, int64) ([]porte.Session, error) {
	return nil, nil
}
func (s stubSessions) DeleteByID(context.Context, int64, int64) error { return nil }
func (s stubSessions) DeleteExpired(context.Context, time.Time) (int64, error) {
	return 0, nil
}

type authenticator struct{ manager *session.Manager }

func (a authenticator) RequireAuth(next http.Handler) http.Handler {
	return a.manager.RequireAuth(next)
}

func (a authenticator) IdentityForUser(_ context.Context, userID int64) (int64, string, error) {
	return userID, "user@example.com", nil
}

func newHandler(t *testing.T) http.Handler {
	t.Helper()
	cfg := porte.Config{MachineAudience: "opus-client", ScopePrefix: "opus"}
	manager, err := session.New(cfg, session.Deps{
		Sessions: stubSessions{hash: porte.HashToken(opaqueToken)},
		Logger:   slog.New(slog.DiscardHandler),
	})
	if err != nil {
		t.Fatalf("session manager: %v", err)
	}
	manager = manager.WithJWT(stubVerifier{})
	protected := middleware.RequireAuth(authenticator{manager: manager})(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			identity, _ := authcontext.IdentityFromContext(r.Context())
			if identity.UserID != 7 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("X-Actor", identity.Actor)
			w.WriteHeader(http.StatusOK)
		}),
	)
	return protected
}

func status(handler http.Handler, method, token string) (int, string) {
	request := httptest.NewRequest(method, "/api/workspace", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder.Code, recorder.Header().Get("X-Actor")
}

func TestDelegatedJWTIsScopeGated(t *testing.T) {
	handler := newHandler(t)

	if code, actor := status(handler, http.MethodGet, readToken); code != http.StatusOK || actor != "bulle" {
		t.Fatalf("read scope on GET: status %d actor %q, want 200 bulle", code, actor)
	}
	if code, _ := status(handler, http.MethodPost, readToken); code != http.StatusForbidden {
		t.Fatalf("read scope on POST: status %d, want 403", code)
	}
	if code, _ := status(handler, http.MethodPost, writeToken); code != http.StatusOK {
		t.Fatalf("write scope on POST: status %d, want 200", code)
	}
}

func TestOpaqueAPITokenIsNotScopeGated(t *testing.T) {
	handler := newHandler(t)

	for _, method := range []string{http.MethodGet, http.MethodPost} {
		if code, actor := status(handler, method, opaqueToken); code != http.StatusOK || actor != "" {
			t.Fatalf("opaque token on %s: status %d actor %q, want 200 and no actor", method, code, actor)
		}
	}
}
