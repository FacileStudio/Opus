// Package task implements HTTP request routing for the task API.
package task

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

// RegisterRoutes wires the authenticated task endpoints.
func RegisterRoutes(router chi.Router, service *Service, authService middleware.Authenticator) {
	router.Route("/task", func(sub chi.Router) {
		sub.Use(middleware.RequireAuth(authService))
		registerTaskBulkRoutes(sub, service)
		registerTaskQueryRoutes(sub, service)
		registerTaskMutationRoutes(sub, service)
	})
}

func registerTaskQueryRoutes(router chi.Router, service *Service) {
	router.Get("/tasks/{projectId}", func(w http.ResponseWriter, request *http.Request) {
		projectID, err := parseIDParam(request, "projectId")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid project id"))
			return
		}
		filter, _ := parseListFilter(request)
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.list(request.Context(), identity.UserID, projectID, filter)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
	router.Get("/{id}", func(w http.ResponseWriter, request *http.Request) {
		taskID, err := parseIDParam(request, "id")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid task id"))
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.get(request.Context(), identity.UserID, taskID)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

func registerTaskBulkRoutes(router chi.Router, service *Service) {
	router.Patch("/bulk", func(w http.ResponseWriter, request *http.Request) {
		var req BulkUpdateRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.bulk(request.Context(), identity.UserID, &req)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

func registerTaskMutationRoutes(router chi.Router, service *Service) {
	registerPropertyMutationRoutes(router, service)
	registerContentMutationRoutes(router, service)
	registerLifecycleMutationRoutes(router, service)
}

func registerLifecycleMutationRoutes(router chi.Router, service *Service) {
	registerCreateTaskRoute(router, service)
	registerMoveTaskRoute(router, service)
	registerUpdateTaskRoute(router, service)
	registerDeleteTaskRoute(router, service)
}

func registerCreateTaskRoute(router chi.Router, service *Service) {
	router.Post("/{projectId}", func(w http.ResponseWriter, request *http.Request) {
		projectID, err := parseIDParam(request, "projectId")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid project id"))
			return
		}
		var req CreateTaskRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.create(request.Context(), identity.UserID, projectID, &req)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

func registerMoveTaskRoute(router chi.Router, service *Service) {
	router.Put("/move/{id}", func(w http.ResponseWriter, request *http.Request) {
		taskID, err := parseIDParam(request, "id")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid task id"))
			return
		}
		var req MoveTaskRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.move(request.Context(), identity.UserID, taskID, &req)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

func registerUpdateTaskRoute(router chi.Router, service *Service) {
	router.Put("/{id}", func(w http.ResponseWriter, request *http.Request) {
		taskID, err := parseIDParam(request, "id")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid task id"))
			return
		}
		var req UpdateTaskRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.update(request.Context(), identity.UserID, taskID, &req)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

func registerDeleteTaskRoute(router chi.Router, service *Service) {
	router.Delete("/{id}", func(w http.ResponseWriter, request *http.Request) {
		taskID, err := parseIDParam(request, "id")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid task id"))
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.delete(request.Context(), identity.UserID, taskID)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}
