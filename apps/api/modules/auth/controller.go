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

// register creates an account through porte/local and signs it in, keeping the
// frontend's {token, user} response shape. porte exports the service rather
// than only the routes for exactly this reason.
func (controller *Controller) register(w http.ResponseWriter, r *http.Request, req *RegisterRequest) (*AuthResponse, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" {
		return nil, errors.Invalid("invalid email")
	}
	user, token, err := controller.service.Register(r.Context(), w, r, email, strings.TrimSpace(req.Name), req.Password)
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
