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

// AuthResponse is the {token, user} shape register, login and /auth/me all
// answer with. The frontend was better-auth shaped, whose session is exactly
// this pair; the re-plumb keeps it.
type AuthResponse struct {
	Token string        `json:"token"`
	User  *UserResponse `json:"user"`
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
