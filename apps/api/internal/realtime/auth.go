package realtime

import (
	"net/http"

	"github.com/FacileStudio/porte"
)

// Authenticator validates an incoming HTTP connection before WebSocket upgrade.
type Authenticator interface {
	Authenticate(w http.ResponseWriter, r *http.Request) (int64, error)
}

// AuthFunc is a function adapter implementing Authenticator.
type AuthFunc func(w http.ResponseWriter, r *http.Request) (int64, error)

// Authenticate calls the wrapped function.
func (f AuthFunc) Authenticate(w http.ResponseWriter, r *http.Request) (int64, error) {
	return f(w, r)
}

// SessionVerifier is the porte interface required to authenticate session credentials.
type SessionVerifier interface {
	Authenticate(w http.ResponseWriter, r *http.Request) (porte.Identity, error)
}

// SessionAuthenticator adapts a porte SessionVerifier to Authenticator.
type SessionAuthenticator struct {
	verifier SessionVerifier
}

// NewSessionAuthenticator creates an Authenticator backed by a porte SessionVerifier.
func NewSessionAuthenticator(verifier SessionVerifier) *SessionAuthenticator {
	return &SessionAuthenticator{verifier: verifier}
}

// Authenticate verifies the session cookie or bearer token on the request.
func (a *SessionAuthenticator) Authenticate(w http.ResponseWriter, r *http.Request) (int64, error) {
	req := prepareAuthRequest(r)
	identity, err := a.verifier.Authenticate(w, req)
	if err != nil {
		return 0, err
	}
	return identity.UserID, nil
}

func prepareAuthRequest(r *http.Request) *http.Request {
	if r.Header.Get("Authorization") != "" {
		return r
	}
	token := r.URL.Query().Get("token")
	if token == "" {
		token = r.URL.Query().Get("access_token")
	}
	if token != "" {
		req := r.Clone(r.Context())
		req.Header.Set("Authorization", "Bearer "+token)
		return req
	}
	return r
}
