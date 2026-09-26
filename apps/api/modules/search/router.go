package search

import (
	"net/http"

	"github.com/FacileStudio/Opus/apps/api/internal/authcontext"
	"github.com/FacileStudio/Opus/apps/api/internal/middleware"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts authenticated search endpoints.
func RegisterRoutes(router chi.Router, service *Service, authService middleware.Authenticator) {
	router.Route("/search", func(router chi.Router) {
		router.Use(middleware.RequireAuth(authService))
		registerSearchRoute(router, service)
	})
}

func registerSearchRoute(router chi.Router, service *Service) {
	router.Get("/", func(w http.ResponseWriter, request *http.Request) {
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.search(request.Context(), identity.UserID, request)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}
