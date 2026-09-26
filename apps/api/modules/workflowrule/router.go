package workflowrule

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

// RegisterRoutes mounts authenticated workflow rule endpoints.
func RegisterRoutes(router chi.Router, service *Service, authService middleware.Authenticator) {
	router.Route("/workflow-rule", func(router chi.Router) {
		router.Use(middleware.RequireAuth(authService))
		registerProjectRuleRoutes(router, service)
		registerItemRoutes(router, service)
	})
}

func registerProjectRuleRoutes(router chi.Router, service *Service) {
	router.Get("/{projectId}", func(w http.ResponseWriter, request *http.Request) {
		projectID, err := parseID(request, "projectId")
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

	router.Put("/{projectId}", func(w http.ResponseWriter, request *http.Request) {
		projectID, err := parseID(request, "projectId")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid project id"))
			return
		}
		var req UpsertWorkflowRuleRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.upsert(request.Context(), identity.UserID, projectID, &req)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

func registerItemRoutes(router chi.Router, service *Service) {
	router.Delete("/{id}", func(w http.ResponseWriter, request *http.Request) {
		ruleID, err := parseID(request, "id")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid workflow rule id"))
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.delete(request.Context(), identity.UserID, ruleID)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}
