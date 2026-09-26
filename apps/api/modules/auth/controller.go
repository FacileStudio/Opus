package auth

import (
	"context"
	"net/http"
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

// changePassword verifies the current password, sets a new one and rotates the
// session in one porte call.
func (controller *Controller) changePassword(w http.ResponseWriter, r *http.Request, userID int64, req *ChangePasswordRequest) (*ChangePasswordResponse, error) {
	if req.CurrentPassword == "" || req.NewPassword == "" {
		return nil, errors.Invalid("currentPassword and newPassword are required")
	}
	if err := controller.service.ChangePassword(w, r, userID, req.CurrentPassword, req.NewPassword); err != nil {
		return nil, err
	}
	return &ChangePasswordResponse{Message: "Password changed successfully"}, nil
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

// register creates an account through porte/local and signs it in, keeping the
// frontend's {token, user} response shape. porte exports the service rather
// than only the routes for exactly this reason.
func (controller *Controller) register(w http.ResponseWriter, r *http.Request, req *RegisterRequest) (*AuthResponse, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" {
		return nil, errors.Invalid("invalid email")
	}
	user, token, err := controller.service.Register(w, r, email, strings.TrimSpace(req.Name), req.Password)
	if err != nil {
		return nil, err
	}
	return &AuthResponse{Token: token, User: toUserResponse(user)}, nil
}

func (controller *Controller) login(w http.ResponseWriter, r *http.Request, req *LoginRequest) (*AuthResponse, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" || req.Password == "" {
		return nil, errors.Invalid("email and password required")
	}
	user, token, err := controller.service.Login(r.Context(), w, r, email, req.Password)
	if err != nil {
		return nil, err
	}
	return &AuthResponse{Token: token, User: toUserResponse(user)}, nil
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
