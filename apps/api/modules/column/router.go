package column

import (
	"net/http"
	"strconv"

	"github.com/FacileStudio/Opus/apps/api/internal/authcontext"
	"github.com/FacileStudio/Opus/apps/api/internal/middleware"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/go-chi/chi/v5"
)

func parseIDParam(r *http.Request, key string) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, key), 10, 64)
}

// RegisterRoutes wires the authenticated column endpoints.
func RegisterRoutes(router chi.Router, service *Service, authService middleware.Authenticator) {
	router.Route("/column", func(sub chi.Router) {
		sub.Use(middleware.RequireAuth(authService))
		registerColumnOrder(sub, service)
		registerColumnCollection(sub, service)
		registerColumnItem(sub, service)
	})
}

func registerColumnCollection(router chi.Router, service *Service) {
	router.Get("/{projectId}", func(w http.ResponseWriter, request *http.Request) {
		identity, _ := authcontext.IdentityFromContext(request.Context())
		projectID, err := parseIDParam(request, "projectId")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid project id"))
			return
		}
		resp, err := service.controller.list(request.Context(), identity.UserID, projectID)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})

	router.Post("/{projectId}", func(w http.ResponseWriter, request *http.Request) {
		var req CreateColumnRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		projectID, err := parseIDParam(request, "projectId")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid project id"))
			return
		}
		resp, err := service.controller.create(request.Context(), identity.UserID, projectID, &req)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

func registerColumnItem(router chi.Router, service *Service) {
	router.Put("/{id}", func(w http.ResponseWriter, request *http.Request) {
		var req UpdateColumnRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		columnID, err := parseIDParam(request, "id")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid column id"))
			return
		}
		resp, err := service.controller.update(request.Context(), identity.UserID, columnID, &req)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})

	router.Delete("/{id}", func(w http.ResponseWriter, request *http.Request) {
		identity, _ := authcontext.IdentityFromContext(request.Context())
		columnID, err := parseIDParam(request, "id")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid column id"))
			return
		}
		resp, err := service.controller.delete(request.Context(), identity.UserID, columnID)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

func registerColumnOrder(router chi.Router, service *Service) {
	router.Put("/reorder/{projectId}", func(w http.ResponseWriter, request *http.Request) {
		var req ReorderColumnsRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		projectID, err := parseIDParam(request, "projectId")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid project id"))
			return
		}
		resp, err := service.controller.reorder(request.Context(), identity.UserID, projectID, &req)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}
