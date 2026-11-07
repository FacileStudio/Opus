package auth

import (
	"net/http"

	"github.com/FacileStudio/Opus/apps/api/internal/authcontext"
	"github.com/FacileStudio/Opus/apps/api/internal/middleware"
	"github.com/FacileStudio/tronc/httpjson"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts what porte does not own: the profile and the personal API tokens.
func RegisterRoutes(router chi.Router, service *Service) {
	router.Route("/auth", func(router chi.Router) {
		registerMeRoutes(router, service)
	})
}

func registerMeRoutes(router chi.Router, service *Service) {
	router.Group(func(router chi.Router) {
		router.Use(middleware.RequireAuth(service))
		registerGetMe(router, service)
		registerPatchMe(router, service)
		registerTokenRoutes(router, service)
	})
}

// registerTokenRoutes mounts the personal API-token lifecycle. The tokens are
// porte sessions with a label, so they are listed, revoked and accepted as a
// bearer by the middleware that already guards every other route.
func registerTokenRoutes(router chi.Router, service *Service) {
	router.Get("/api-tokens", func(w http.ResponseWriter, request *http.Request) {
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.listTokens(request.Context(), identity.UserID)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})

	router.Post("/api-tokens", func(w http.ResponseWriter, request *http.Request) {
		var req CreateTokenRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.createToken(request.Context(), identity.UserID, &req)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusCreated, resp)
	})

	router.Delete("/api-tokens/{id}", func(w http.ResponseWriter, request *http.Request) {
		identity, _ := authcontext.IdentityFromContext(request.Context())
		err := service.controller.revokeToken(request.Context(), identity.UserID, chi.URLParam(request, "id"))
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, map[string]bool{"deleted": true})
	})
}

func registerGetMe(router chi.Router, service *Service) {
	router.Get("/me", func(w http.ResponseWriter, request *http.Request) {
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.me(request.Context(), identity.UserID)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

func registerPatchMe(router chi.Router, service *Service) {
	router.Patch("/me", func(w http.ResponseWriter, request *http.Request) {
		var req UpdateProfileRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.updateProfile(w, request, identity.UserID, &req)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}
