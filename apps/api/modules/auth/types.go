package auth

import "time"

// RegisterRequest is the body for a local registration.
type RegisterRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name,omitempty"`
	Password string `json:"password"`
}

// LoginRequest is the body for a local login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse is the {token, user} shape register, login and PATCH /auth/me
// answer with. GET /auth/me answers the bare UserResponse. The frontend was
// better-auth shaped, whose session is exactly the pair; the re-plumb keeps it.
type AuthResponse struct {
	Token string        `json:"token"`
	User  *UserResponse `json:"user"`
}

// UpdateProfileRequest is the body of PATCH /auth/me. Every field is a pointer
// so an omitted key is not silently cleared.
type UpdateProfileRequest struct {
	Name   *string `json:"name,omitempty"`
	Color  *string `json:"color,omitempty"`
	Locale *string `json:"locale,omitempty"`
}

// ChangePasswordRequest is the body of POST /auth/change-password.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

// ChangePasswordResponse is what POST /auth/change-password answers.
type ChangePasswordResponse struct {
	Message string `json:"message"`
}

// UserResponse is the profile the frontend reads. IDs are int64 on the wire:
// the port adopts the suite's identity model, and the frontend treats them as
// opaque strings anyway.
type UserResponse struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"emailVerified"`
	Image         string    `json:"image"`
	Color         string    `json:"color"`
	Locale        string    `json:"locale"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
