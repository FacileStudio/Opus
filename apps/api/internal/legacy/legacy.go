// Package legacy proxies routes the Go backend does not own yet to the
// transitional TypeScript API — the strangler-fig fallback while the port
// flips features Go-native one at a time.
package legacy

import (
	"net/http/httputil"
	"net/url"

	"github.com/go-chi/chi/v5"
)

// MountProxy attaches a reverse proxy that forwards every /api route not
// handled natively to the TypeScript API at remote, preserving the request
// path and passing WebSocket upgrades through to it.
func MountProxy(api chi.Router, remote *url.URL) {
	proxy := httputil.NewSingleHostReverseProxy(remote)
	api.Handle("/*", proxy)
}
