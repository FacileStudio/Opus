// Package middleware carries the shared request-context plumbing between
// porte and the module handlers.
package middleware

import (
	"context"
	"net/http"

	"github.com/FacileStudio/Opus/apps/api/internal/authcontext"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
)

// Authenticator is the auth service: porte's session middleware, plus the
// lookup that turns the user id porte resolved into the identity the app
// reads. It stays one parameter so every module router keeps calling
// middleware.RequireAuth(authService).
type Authenticator interface {
	RequireAuth(http.Handler) http.Handler
	IdentityForUser(ctx context.Context, userID int64) (int64, string, error)
}

// RequireAuth runs porte's middleware and then fills in what porte deliberately
// does not carry: the email, looked up from the app's own row in one query,
// and lands in the same context every controller reads.
func RequireAuth(auth Authenticator) func(http.Handler) http.Handler {
	hydrate := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
			authenticated, ok := porte.From(request.Context())
			if !ok {
				httpjson.WriteError(w, errors.Unauthorized("missing auth token"))
				return
			}
			userID, email, err := auth.IdentityForUser(request.Context(), authenticated.UserID)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			actor, _ := porte.ActorFrom(request.Context())
			authContext := authcontext.WithIdentity(request.Context(), authcontext.Identity{
				UserID: userID,
				Email:  email,
				Actor:  actor,
			})
			next.ServeHTTP(w, request.WithContext(authContext))
		})
	}
	return func(next http.Handler) http.Handler {
		return auth.RequireAuth(hydrate(next))
	}
}

// OptionalAuth attempts to authenticate the request, populating identity if valid.
func OptionalAuth(auth Authenticator) func(http.Handler) http.Handler {
	hydrate := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
			authenticated, ok := porte.From(request.Context())
			if !ok {
				next.ServeHTTP(w, request)
				return
			}
			userID, email, err := auth.IdentityForUser(request.Context(), authenticated.UserID)
			if err != nil {
				next.ServeHTTP(w, request)
				return
			}
			authContext := authcontext.WithIdentity(request.Context(), authcontext.Identity{
				UserID: userID,
				Email:  email,
			})
			next.ServeHTTP(w, request.WithContext(authContext))
		})
	}
	return func(next http.Handler) http.Handler {
		if optionaler, ok := auth.(interface {
			Optional(http.Handler) http.Handler
		}); ok {
			return optionaler.Optional(hydrate(next))
		}
		return hydrate(next)
	}
}
