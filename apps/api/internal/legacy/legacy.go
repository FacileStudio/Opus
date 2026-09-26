// Package legacy proxies routes the Go backend does not own yet to the
// transitional TypeScript API — the strangler-fig fallback while the port
// flips features Go-native one at a time.
//
// The proxy is also the auth bridge: the browser now holds a porte session
// that the legacy API cannot read, so the proxy authenticates every proxied
// request with porte and, when a session is valid, stamps the forwarded
// request with the user's identity plus an HMAC tag derived from a shared
// secret. The legacy API short-circuits its own better-auth lookup on those
// headers. Unauthenticated requests pass through untouched, so public proxied
// routes keep working and protected ones keep failing closed with the legacy
// 401, exactly as before the bridge.
package legacy

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"

	"github.com/FacileStudio/porte"
	"github.com/go-chi/chi/v5"
)

const (
	userIDHeader    = "X-Opus-User-Id"
	userEmailHeader = "X-Opus-User-Email"
	authHeader      = "X-Opus-Auth"
)

// corsHeaders are set by the legacy API's own cors middleware; forwarding
// them would duplicate the ones the Go router already applies.
var corsHeaders = []string{
	"Access-Control-Allow-Origin",
	"Access-Control-Allow-Credentials",
	"Access-Control-Allow-Headers",
	"Access-Control-Allow-Methods",
	"Access-Control-Expose-Headers",
	"Access-Control-Max-Age",
}

// Authenticator is the slice of the session manager the bridge needs: an
// optional auth check that attaches an identity when one is present and lets
// the request through either way.
type Authenticator interface {
	Optional(http.Handler) http.Handler
}

// IdentityResolver turns the user id porte authenticated into the profile the
// proxied request carries. The auth service satisfies it.
type IdentityResolver interface {
	IdentityForUser(ctx context.Context, userID int64) (int64, string, error)
}

// tag computes the value the legacy API verifies before trusting the bridge
// headers. It binds the tag to the user id, so one stamped request cannot be
// replayed against a different identity.
func tag(secret, userID string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(userID))
	return hex.EncodeToString(mac.Sum(nil))
}

// MountProxy attaches a reverse proxy that forwards every /api route not
// handled natively to the TypeScript API at remote, preserving the request
// path and passing WebSocket upgrades through to it.
//
// The proxy authenticates each request with the optional session check first.
// On success it strips any caller-supplied bridge headers and injects the
// verified identity; on failure it forwards the request untouched, so the
// legacy API decides. The tag header is only ever written when a porte
// session was verified, and the legacy API only honours the headers when the
// tag matches the shared secret.
//
// Upstream CORS headers are stripped: the legacy API runs its own hono cors
// middleware, and forwarding its Access-Control-* headers alongside the ones
// tronc's CORS middleware sets duplicates Access-Control-Allow-Origin on the
// response. Browsers reject duplicated ACAO on credentialed requests, which
// silently fails every proxied fetch in the web app while the server still
// logs a 200.
func MountProxy(api chi.Router, remote *url.URL, sessions Authenticator, identities IdentityResolver, secret string) {
	proxy := httputil.NewSingleHostReverseProxy(remote)
	director := proxy.Director
	proxy.Director = func(request *http.Request) {
		director(request)
		request.Header.Del(userIDHeader)
		request.Header.Del(userEmailHeader)
		request.Header.Del(authHeader)
		if secret == "" {
			return
		}
		identity, ok := porte.From(request.Context())
		if !ok || identity.UserID == 0 {
			return
		}
		userID, email, err := identities.IdentityForUser(request.Context(), identity.UserID)
		if err != nil {
			return
		}
		request.Header.Set(userIDHeader, strconv.FormatInt(userID, 10))
		request.Header.Set(userEmailHeader, email)
		request.Header.Set(authHeader, tag(secret, strconv.FormatInt(userID, 10)))
	}
	modifyResponse := proxy.ModifyResponse
	proxy.ModifyResponse = func(response *http.Response) error {
		for _, header := range corsHeaders {
			response.Header.Del(header)
		}
		if modifyResponse != nil {
			return modifyResponse(response)
		}
		return nil
	}
	api.Handle("/*", sessions.Optional(proxy))
}
