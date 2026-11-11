package auth

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/FacileStudio/Opus/apps/api/internal/database/dbtest"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/session"
	"github.com/FacileStudio/tronc/errors"
)

type memorySessions struct {
	rows []porte.Session
	next int64
}

func (m *memorySessions) Create(_ context.Context, record porte.Session) (porte.Session, error) {
	m.next++
	record.ID = m.next
	m.rows = append(m.rows, record)
	return record, nil
}

func (m *memorySessions) Find(_ context.Context, tokenHash string) (porte.Session, error) {
	for _, row := range m.rows {
		if row.TokenHash == tokenHash {
			return row, nil
		}
	}
	return porte.Session{}, porte.ErrNotFound
}

func (m *memorySessions) Touch(_ context.Context, tokenHash string, at time.Time) error {
	for i := range m.rows {
		if m.rows[i].TokenHash == tokenHash {
			m.rows[i].LastUsedAt = at
		}
	}
	return nil
}

func (m *memorySessions) Delete(_ context.Context, tokenHash string) error {
	for i, row := range m.rows {
		if row.TokenHash == tokenHash {
			m.rows = append(m.rows[:i], m.rows[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *memorySessions) DeleteByUser(_ context.Context, userID int64) (int64, error) {
	return m.drop(func(row porte.Session) bool { return row.UserID == userID })
}

func (m *memorySessions) DeleteLogins(_ context.Context, userID, except int64) (int64, error) {
	return m.drop(func(row porte.Session) bool {
		return row.UserID == userID && !row.IsAPIToken() && row.ID != except
	})
}

func (m *memorySessions) ListByUser(_ context.Context, userID int64) ([]porte.Session, error) {
	held := make([]porte.Session, 0)
	for _, row := range m.rows {
		if row.UserID == userID {
			held = append(held, row)
		}
	}
	return held, nil
}

func (m *memorySessions) DeleteByID(_ context.Context, userID, sessionID int64) error {
	for i, row := range m.rows {
		if row.UserID == userID && row.ID == sessionID {
			m.rows = append(m.rows[:i], m.rows[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *memorySessions) DeleteExpired(_ context.Context, now time.Time) (int64, error) {
	return m.drop(func(row porte.Session) bool { return row.Expired(now) })
}

func (m *memorySessions) drop(match func(porte.Session) bool) (int64, error) {
	kept := make([]porte.Session, 0, len(m.rows))
	var removed int64
	for _, row := range m.rows {
		if match(row) {
			removed++
			continue
		}
		kept = append(kept, row)
	}
	m.rows = kept
	return removed, nil
}

func newTokenService(t *testing.T) (*Service, *session.Manager) {
	t.Helper()
	db := dbtest.Open(t)
	manager, err := session.New(porte.Config{}, session.Deps{
		Sessions: &memorySessions{},
		Logger:   slog.New(slog.DiscardHandler),
	})
	if err != nil {
		t.Fatalf("session manager: %v", err)
	}
	return NewService(db, manager, slog.New(slog.DiscardHandler)), manager
}

func TestIssuedTokenAuthenticatesAsBearer(t *testing.T) {
	service, manager := newTokenService(t)

	rawToken, token, err := service.IssueToken(context.Background(), 7, "laptop")
	if err != nil {
		t.Fatalf("IssueToken: %v", err)
	}
	if rawToken == "" || token.Name != "laptop" || token.ID == 0 {
		t.Fatalf("issued %+v with token %q", token, rawToken)
	}

	var authenticated int64
	handler := manager.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, ok := porte.From(r.Context())
		if !ok {
			t.Errorf("no identity in context")
		}
		authenticated = identity.UserID
		w.WriteHeader(http.StatusOK)
	}))

	request := httptest.NewRequest(http.MethodGet, "/api/workspace", nil)
	request.Header.Set("Authorization", "Bearer "+rawToken)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("token bearer rejected: status %d", recorder.Code)
	}
	if authenticated != 7 {
		t.Fatalf("authenticated user = %d, want 7", authenticated)
	}
}

func TestListTokensHidesBrowserLogins(t *testing.T) {
	service, manager := newTokenService(t)
	ctx := context.Background()

	if _, _, err := service.IssueToken(ctx, 7, "nightly"); err != nil {
		t.Fatalf("IssueToken: %v", err)
	}
	if _, _, err := manager.Issue(ctx, 7, ""); err != nil {
		t.Fatalf("Issue login: %v", err)
	}

	tokens, err := service.ListTokens(ctx, 7)
	if err != nil {
		t.Fatalf("ListTokens: %v", err)
	}
	if len(tokens) != 1 || tokens[0].Name != "nightly" {
		t.Fatalf("tokens = %+v, want only the named one", tokens)
	}
}

func TestRevokeTokenRefusesUnnamedAndForeign(t *testing.T) {
	service, manager := newTokenService(t)
	ctx := context.Background()

	_, token, err := service.IssueToken(ctx, 7, "nightly")
	if err != nil {
		t.Fatalf("IssueToken: %v", err)
	}
	_, login, err := manager.Issue(ctx, 7, "")
	if err != nil {
		t.Fatalf("Issue login: %v", err)
	}

	if err := service.RevokeToken(ctx, 7, login.ID); errors.Status(err) != 404 {
		t.Fatalf("revoking a browser login status = %d, want 404", errors.Status(err))
	}
	if err := service.RevokeToken(ctx, 8, token.ID); errors.Status(err) != 404 {
		t.Fatalf("revoking another user's token status = %d, want 404", errors.Status(err))
	}
	if err := service.RevokeToken(ctx, 7, token.ID); err != nil {
		t.Fatalf("RevokeToken: %v", err)
	}

	tokens, err := service.ListTokens(ctx, 7)
	if err != nil {
		t.Fatalf("ListTokens: %v", err)
	}
	if len(tokens) != 0 {
		t.Fatalf("tokens after revoke = %+v, want none", tokens)
	}
}

func TestIssueTokenRejectsEmptyName(t *testing.T) {
	service, _ := newTokenService(t)

	if _, _, err := service.IssueToken(context.Background(), 7, "   "); errors.Status(err) != 400 {
		t.Fatalf("empty name status = %d, want 400", errors.Status(err))
	}
}
