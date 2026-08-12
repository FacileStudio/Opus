// Package authcontext carries the authenticated user's identity through a
// request context, after porte has verified the credential and the app has
// loaded the profile row.
package authcontext

import "context"

// Identity is what an authenticated handler reads. UserID is the app-side
// int64 primary key, matching the suite's porte/identity model.
type Identity struct {
	UserID int64
	Email  string
}

type contextKey struct{}

// WithIdentity returns a copy of parentContext carrying the given Identity.
func WithIdentity(parentContext context.Context, identity Identity) context.Context {
	return context.WithValue(parentContext, contextKey{}, identity)
}

// IdentityFromContext extracts the Identity stored on parentContext, reporting
// whether one was present.
func IdentityFromContext(parentContext context.Context) (Identity, bool) {
	identity, ok := parentContext.Value(contextKey{}).(Identity)
	return identity, ok
}
