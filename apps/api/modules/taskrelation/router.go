package taskrelation

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

// RegisterRoutes wires the authenticated task relation endpoints into the router.
func RegisterRoutes(router chi.Router, service *Service, authService middleware.Authenticator) {
	router.Route("/task-relation", func(router chi.Router) {
		router.Use(middleware.RequireAuth(authService))
		registerTaskRelationCollection(router, service)
		registerTaskRelationItem(router, service)
	})
}

func registerTaskRelationCollection(router chi.Router, service *Service) {
	router.Get("/{taskId}", func(w http.ResponseWriter, request *http.Request) {
		taskID, err := parseID(request, "taskId")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid task id"))
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.list(request.Context(), identity.UserID, taskID)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})

	router.Post("/", func(w http.ResponseWriter, request *http.Request) {
		var req CreateTaskRelationRequest
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

func registerTaskRelationItem(router chi.Router, service *Service) {
	router.Delete("/{id}", func(w http.ResponseWriter, request *http.Request) {
		id, err := parseID(request, "id")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid relation id"))
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.delete(request.Context(), identity.UserID, id)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}
