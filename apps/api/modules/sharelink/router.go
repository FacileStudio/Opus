package sharelink

import (
	"net/http"
	"strconv"

	"github.com/FacileStudio/Opus/apps/api/internal/authcontext"
	"github.com/FacileStudio/Opus/apps/api/internal/middleware"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/go-chi/chi/v5"
)

func parseID(r *http.Request, key string) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, key), 10, 64)
}

// RegisterRoutes mounts authenticated and public share link endpoints.
func RegisterRoutes(router chi.Router, service *Service, authService middleware.Authenticator) {
	router.Route("/share-link", func(router chi.Router) {
		router.Use(middleware.RequireAuth(authService))
		registerCollectionRoutes(router, service)
		registerItemRoutes(router, service)
	})
	router.Get("/shared-project/{token}", handleSharedProject(service))
	router.Get("/share/{token}", handleSharedProject(service))
}

func registerCollectionRoutes(router chi.Router, service *Service) {
	router.Get("/", func(w http.ResponseWriter, request *http.Request) {
		projectID, err := strconv.ParseInt(request.URL.Query().Get("projectId"), 10, 64)
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid project id"))
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.list(request.Context(), identity.UserID, projectID)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})

	router.Post("/", func(w http.ResponseWriter, request *http.Request) {
		var req CreateShareLinkRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.create(request.Context(), identity.UserID, &req)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

func registerItemRoutes(router chi.Router, service *Service) {
	router.Put("/{id}", func(w http.ResponseWriter, request *http.Request) {
		linkID, err := parseID(request, "id")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid share link id"))
			return
		}
		var req UpdateShareLinkRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.update(request.Context(), identity.UserID, linkID, &req)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})

	router.Delete("/{id}", func(w http.ResponseWriter, request *http.Request) {
		linkID, err := parseID(request, "id")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid share link id"))
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.delete(request.Context(), identity.UserID, linkID)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

func handleSharedProject(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		token := chi.URLParam(request, "token")
		resp, err := service.controller.getShared(request.Context(), token)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	}
}
