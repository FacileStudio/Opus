package auth

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
)

// Controller serves the /auth routes porte does not own.
type Controller struct {
	service *Service
}

func newController(service *Service) *Controller {
	return &Controller{service: service}
}

// me returns the profile of the authenticated caller.
func (controller *Controller) me(ctx context.Context, userID int64) (*UserResponse, error) {
	user, err := controller.service.LoadUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return toUserResponse(user), nil
}

// updateProfile writes the caller's own profile fields and answers the same
// {token, user} shape the rest of the auth surface uses. The token echoes the
// bearer the request carried; a cookie-authenticated caller gets an empty
// string, which the frontend ignores.
func (controller *Controller) updateProfile(w http.ResponseWriter, r *http.Request, userID int64, req *UpdateProfileRequest) (*AuthResponse, error) {
	user, err := controller.service.UpdateProfile(r.Context(), userID, *req)
	if err != nil {
		return nil, err
	}
	return &AuthResponse{Token: bearerToken(r), User: toUserResponse(user)}, nil
}

// bearerToken extracts the Authorization bearer the current request carried,
// or an empty string when it authenticated some other way.
func bearerToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if len(header) > len(prefix) && strings.EqualFold(header[:len(prefix)], prefix) {
		return strings.TrimSpace(header[len(prefix):])
	}
	return ""
}

// createToken mints a personal API token and answers it once, in plaintext.
func (controller *Controller) createToken(ctx context.Context, userID int64, req *CreateTokenRequest) (*TokenResponse, error) {
	rawToken, token, err := controller.service.IssueToken(ctx, userID, req.Name)
	if err != nil {
		return nil, err
	}
	token.Token = rawToken
	return &token, nil
}

// listTokens returns the caller's own tokens, never anybody else's.
func (controller *Controller) listTokens(ctx context.Context, userID int64) (*TokenListResponse, error) {
	tokens, err := controller.service.ListTokens(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &TokenListResponse{Tokens: tokens}, nil
}

// revokeToken ends one token by id.
func (controller *Controller) revokeToken(ctx context.Context, userID int64, tokenID string) error {
	id, err := strconv.ParseInt(tokenID, 10, 64)
	if err != nil || id <= 0 {
		return errors.Invalid("token id must be a positive integer")
	}
	return controller.service.RevokeToken(ctx, userID, id)
}

func toUserResponse(user *schemas.User) *UserResponse {
	return &UserResponse{
		ID:            user.ID,
		Name:          user.Name,
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
		Image:         user.Image,
		Color:         user.Color,
		Locale:        user.Locale,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}
}
