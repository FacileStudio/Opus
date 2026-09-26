package label

import (
	"net/http"
	"strconv"

	"github.com/FacileStudio/Opus/apps/api/internal/authcontext"
	"github.com/FacileStudio/Opus/apps/api/internal/middleware"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/go-chi/chi/v5"
)

func parseParamID(r *http.Request, key string) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, key), 10, 64)
}

// RegisterRoutes wires the authenticated label endpoints into the router.
func RegisterRoutes(router chi.Router, service *Service, authService middleware.Authenticator) {
	router.Route("/label", func(router chi.Router) {
		router.Use(middleware.RequireAuth(authService))
		registerLabelCollection(router, service)
		registerLabelItem(router, service)
	})
}

func registerLabelCollection(router chi.Router, service *Service) {
	router.Post("/", func(w http.ResponseWriter, request *http.Request) {
		var req CreateLabelRequest
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

	router.Get("/workspace/{workspaceId}", func(w http.ResponseWriter, request *http.Request) {
		workspaceID, err := parseParamID(request, "workspaceId")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid workspace id"))
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.listByWorkspace(request.Context(), identity.UserID, workspaceID)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})

	router.Get("/task/{taskId}", func(w http.ResponseWriter, request *http.Request) {
		taskID, err := parseParamID(request, "taskId")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid task id"))
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.listByTask(request.Context(), identity.UserID, taskID)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

func registerLabelItem(router chi.Router, service *Service) {
	router.Route("/{id}", func(router chi.Router) {
		registerLabelLifecycle(router, service)
		registerLabelTaskBinding(router, service)
	})
}

func registerLabelLifecycle(router chi.Router, service *Service) {
	router.Get("/", func(w http.ResponseWriter, request *http.Request) {
		id, err := parseParamID(request, "id")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid label id"))
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.get(request.Context(), identity.UserID, id)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})

	router.Put("/", func(w http.ResponseWriter, request *http.Request) {
		id, err := parseParamID(request, "id")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid label id"))
			return
		}
		var req UpdateLabelRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.update(request.Context(), identity.UserID, id, &req)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})

	router.Delete("/", func(w http.ResponseWriter, request *http.Request) {
		id, err := parseParamID(request, "id")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid label id"))
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

func registerLabelTaskBinding(router chi.Router, service *Service) {
	router.Put("/task", func(w http.ResponseWriter, request *http.Request) {
		id, err := parseParamID(request, "id")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid label id"))
			return
		}
		var req AssignTaskRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.assignToTask(request.Context(), identity.UserID, id, &req)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})

	router.Delete("/task", func(w http.ResponseWriter, request *http.Request) {
		id, err := parseParamID(request, "id")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid label id"))
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.unassignFromTask(request.Context(), identity.UserID, id)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}
