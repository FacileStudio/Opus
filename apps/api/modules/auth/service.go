package auth

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/local"
	"github.com/FacileStudio/porte/session"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

// tokenNameMaxLength caps what a caller may name a token, so a name cannot be
// used as an unbounded log line or list cell.
const tokenNameMaxLength = 64

// Service is what is left of Opus's authentication after porte took the
// credential: the profile lookup the rest of the app reads, the /me answer,
// and a thin wrapper over porte/local so the register and login routes keep a
// shape the frontend can read whole.
type Service struct {
	orm        *gorm.DB
	sessions   *session.Manager
	passwords  *local.Kit
	logger     *slog.Logger
	controller *Controller
}

// NewService wires the auth service with a fresh route controller.
func NewService(orm *gorm.DB, sessions *session.Manager, passwords *local.Kit, logger *slog.Logger) *Service {
	service := &Service{orm: orm, sessions: sessions, passwords: passwords, logger: logger}
	service.controller = newController(service)
	return service
}

// RequireAuth is porte's session middleware, re-exported so every module
// router keeps passing this one service to middleware.RequireAuth.
func (service *Service) RequireAuth(next http.Handler) http.Handler {
	return service.sessions.RequireAuth(next)
}

// Optional is porte's optional session middleware.
func (service *Service) Optional(next http.Handler) http.Handler {
	return service.sessions.Optional(next)
}

// IdentityForUser resolves the user id porte authenticated into the identity
// the rest of the app reads. The email costs one query per authenticated
// request; porte deliberately carries neither email nor role.
func (service *Service) IdentityForUser(ctx context.Context, userID int64) (int64, string, error) {
	var out struct {
		ID    int64
		Email string
	}
	err := service.orm.WithContext(ctx).
		Model(&schemas.User{}).
		Select("id", "email").
		Where("id = ?", userID).
		Scan(&out).Error
	if err != nil {
		return 0, "", errors.Internal("failed to load the account", err)
	}
	if out.ID == 0 {
		return 0, "", errors.Unauthorized("invalid auth token")
	}
	return out.ID, out.Email, nil
}

// LoadUser returns the profile row for an authenticated user, or an
// unauthorized error when the session outlived the account.
func (service *Service) LoadUser(ctx context.Context, userID int64) (*schemas.User, error) {
	var user schemas.User
	err := service.orm.WithContext(ctx).Where("id = ?", userID).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.Unauthorized("invalid auth token")
		}
		return nil, errors.Internal("failed to load the account", err)
	}
	return &user, nil
}

// Register creates an account through porte/local and signs it in. The cookie
// is set on the way out and the token returns for the bearer transport.
func (service *Service) Register(w http.ResponseWriter, r *http.Request, email, name, password string) (*schemas.User, string, error) {
	userID, token, err := service.passwords.Register(r.Context(), w, r, email, name, password)
	if err != nil {
		return nil, "", err
	}
	user, err := service.LoadUser(r.Context(), userID)
	if err != nil {
		return nil, "", err
	}
	return user, token, nil
}

// Login signs the caller in through porte/local and returns their profile.
func (service *Service) Login(ctx context.Context, w http.ResponseWriter, r *http.Request, email, password string) (*schemas.User, string, error) {
	userID, token, err := service.passwords.Login(ctx, w, r, email, password)
	if err != nil {
		return nil, "", err
	}
	user, err := service.LoadUser(ctx, userID)
	if err != nil {
		return nil, "", err
	}
	return user, token, nil
}

// UpdateProfile writes the callable profile fields and returns the refreshed
// row. Only the keys actually present are updated, so a partial patch cannot
// wipe a field the client did not send.
func (service *Service) UpdateProfile(ctx context.Context, userID int64, req UpdateProfileRequest) (*schemas.User, error) {
	updates := map[string]any{}
	if req.Name != nil {
		updates["name"] = strings.TrimSpace(*req.Name)
	}
	if req.Color != nil {
		updates["color"] = *req.Color
	}
	if req.Locale != nil {
		updates["locale"] = *req.Locale
	}
	if len(updates) > 0 {
		if err := service.orm.WithContext(ctx).Model(&schemas.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
			return nil, errors.Internal("failed to update the profile", err)
		}
	}
	return service.LoadUser(ctx, userID)
}

// ChangePassword verifies the caller's current password through porte/local,
// sets a new one and rotates the session so a stolen token dies with the old
// password. The error is passed on so the controller can answer the right
// status instead of a flat unauthorized.
func (service *Service) ChangePassword(w http.ResponseWriter, r *http.Request, userID int64, current, next string) error {
	_, _, err := service.passwords.ChangePassword(r.Context(), w, r, userID, current, next)
	return err
}

// IssueToken mints a personal API token for a user. A named token is a porte
// session carrying a label, so it authenticates on every route a session does,
// it never expires, and porte's sweeper leaves it alone.
func (service *Service) IssueToken(ctx context.Context, userID int64, name string) (string, TokenResponse, error) {
	label, err := normalizeTokenName(name)
	if err != nil {
		return "", TokenResponse{}, err
	}
	rawToken, session, err := service.sessions.Issue(ctx, userID, label)
	if err != nil {
		return "", TokenResponse{}, err
	}
	return rawToken, toTokenResponse(session), nil
}

// ListTokens returns the caller's named tokens. Unlabelled sessions are
// browser logins, and showing somebody their own laptop as an API token, with
// a revoke button next to it, is not the same feature.
func (service *Service) ListTokens(ctx context.Context, userID int64) ([]TokenResponse, error) {
	held, err := service.sessions.List(ctx, userID)
	if err != nil {
		return nil, errors.Internal("failed to list api tokens", err)
	}
	tokens := make([]TokenResponse, 0, len(held))
	for _, session := range held {
		if session.IsAPIToken() {
			tokens = append(tokens, toTokenResponse(session))
		}
	}
	return tokens, nil
}

// RevokeToken ends one named token. It refuses an unlabelled session, so the
// endpoint cannot be used to log a browser out, and porte's Revoke already
// scopes the write to the caller's own id.
func (service *Service) RevokeToken(ctx context.Context, userID, tokenID int64) error {
	held, err := service.sessions.List(ctx, userID)
	if err != nil {
		return errors.Internal("failed to read api tokens", err)
	}
	for _, session := range held {
		if session.ID == tokenID && session.IsAPIToken() {
			return service.sessions.Revoke(ctx, userID, tokenID)
		}
	}
	return errors.NotFound("token not found")
}

func normalizeTokenName(name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", errors.Invalid("name is required")
	}
	if len(trimmed) > tokenNameMaxLength {
		return "", errors.Invalid("name must be at most 64 characters")
	}
	return trimmed, nil
}

func toTokenResponse(session porte.Session) TokenResponse {
	return TokenResponse{
		ID:         session.ID,
		Name:       session.Label,
		CreatedAt:  session.CreatedAt,
		LastUsedAt: session.LastUsedAt,
	}
}