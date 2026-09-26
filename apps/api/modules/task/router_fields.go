// Package task implements HTTP route handlers for single-field task mutations.
package task

import (
	"net/http"

	"github.com/FacileStudio/Opus/apps/api/internal/authcontext"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/go-chi/chi/v5"
)

func registerPropertyMutationRoutes(router chi.Router, service *Service) {
	registerStatusRoute(router, service)
	registerPriorityRoute(router, service)
	registerActorRoute(router, service)
}

func registerContentMutationRoutes(router chi.Router, service *Service) {
	registerDueDateRoute(router, service)
	registerTitleRoute(router, service)
	registerDescriptionRoute(router, service)
}

func registerStatusRoute(router chi.Router, service *Service) {
	router.Put("/status/{id}", func(w http.ResponseWriter, request *http.Request) {
		taskID, err := parseIDParam(request, "id")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid task id"))
			return
		}
		var req UpdateStatusRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.updateStatus(request.Context(), identity.UserID, taskID, req.Status)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

func registerPriorityRoute(router chi.Router, service *Service) {
	router.Put("/priority/{id}", func(w http.ResponseWriter, request *http.Request) {
		taskID, err := parseIDParam(request, "id")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid task id"))
			return
		}
		var req UpdatePriorityRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.updatePriority(request.Context(), identity.UserID, taskID, req.Priority)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

func registerActorRoute(router chi.Router, service *Service) {
	router.Put("/actor/{id}", func(w http.ResponseWriter, request *http.Request) {
		taskID, err := parseIDParam(request, "id")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid task id"))
			return
		}
		var req UpdateActorRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.updateActor(request.Context(), identity.UserID, taskID, &req)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

func registerDueDateRoute(router chi.Router, service *Service) {
	router.Put("/due-date/{id}", func(w http.ResponseWriter, request *http.Request) {
		taskID, err := parseIDParam(request, "id")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid task id"))
			return
		}
		var req UpdateDueDateRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.updateDueDate(request.Context(), identity.UserID, taskID, &req)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

func registerTitleRoute(router chi.Router, service *Service) {
	router.Put("/title/{id}", func(w http.ResponseWriter, request *http.Request) {
		taskID, err := parseIDParam(request, "id")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid task id"))
			return
		}
		var req UpdateTitleRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.updateTitle(request.Context(), identity.UserID, taskID, req.Title)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

func registerDescriptionRoute(router chi.Router, service *Service) {
	router.Put("/description/{id}", func(w http.ResponseWriter, request *http.Request) {
		taskID, err := parseIDParam(request, "id")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid task id"))
			return
		}
		var req UpdateDescriptionRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.updateDescription(request.Context(), identity.UserID, taskID, req.Description)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}
