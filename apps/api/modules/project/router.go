// Package project manages project workspaces, columns, and workflows.
package project

import (
	"net/http"
	"strconv"

	"github.com/FacileStudio/Opus/apps/api/internal/authcontext"
	"github.com/FacileStudio/Opus/apps/api/internal/middleware"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes wires the authenticated project endpoints into the router.
func RegisterRoutes(router chi.Router, service *Service, authService middleware.Authenticator) {
	router.Route("/project", func(router chi.Router) {
		router.Use(middleware.RequireAuth(authService))
		registerProjectCollection(router, service)
		registerProjectItem(router, service)
	})
}

func registerProjectCollection(router chi.Router, service *Service) {
	router.Get("/", func(w http.ResponseWriter, request *http.Request) {
		identity, _ := authcontext.IdentityFromContext(request.Context())
		workspaceIDStr := request.URL.Query().Get("workspaceId")
		includeArchived := request.URL.Query().Get("includeArchived") == "true"
		resp, err := service.controller.list(request.Context(), identity.UserID, workspaceIDStr, includeArchived)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})

	router.Post("/", func(w http.ResponseWriter, request *http.Request) {
		var req CreateProjectRequest
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
		httpjson.WriteJSON(w, http.StatusCreated, resp)
	})
}

func registerProjectItem(router chi.Router, service *Service) {
	router.Route("/{id}", func(router chi.Router) {
		registerProjectDetail(router, service)
		registerProjectActions(router, service)
	})
}

func registerProjectDetail(router chi.Router, service *Service) {
	router.Get("/", func(w http.ResponseWriter, request *http.Request) {
		id, err := parseProjectID(request)
		if err != nil {
			httpjson.WriteError(w, err)
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
		id, err := parseProjectID(request)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		var req UpdateProjectRequest
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
		id, err := parseProjectID(request)
		if err != nil {
			httpjson.WriteError(w, err)
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

func registerProjectActions(router chi.Router, service *Service) {
	router.Put("/archive", func(w http.ResponseWriter, request *http.Request) {
		id, err := parseProjectID(request)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.archive(request.Context(), identity.UserID, id)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})

	router.Put("/unarchive", func(w http.ResponseWriter, request *http.Request) {
		id, err := parseProjectID(request)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.unarchive(request.Context(), identity.UserID, id)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

func parseProjectID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.Invalid("invalid project id")
	}
	return id, nil
}
