package timeentry

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

// RegisterRoutes mounts all authenticated time tracking endpoints.
func RegisterRoutes(router chi.Router, service *Service, authService middleware.Authenticator) {
	router.Route("/time-entry", func(router chi.Router) {
		router.Use(middleware.RequireAuth(authService))
		registerCollectionRoutes(router, service)
		registerTaskRoutes(router, service)
		registerItemRoutes(router, service)
	})
}

func registerCollectionRoutes(router chi.Router, service *Service) {
	router.Post("/", func(w http.ResponseWriter, request *http.Request) {
		var req CreateTimeEntryRequest
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

func registerTaskRoutes(router chi.Router, service *Service) {
	router.Get("/task/{taskId}", func(w http.ResponseWriter, request *http.Request) {
		taskID, err := parseID(request, "taskId")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid task id"))
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.listForTask(request.Context(), identity.UserID, taskID)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

func registerItemRoutes(router chi.Router, service *Service) {
	router.Get("/{id}", func(w http.ResponseWriter, request *http.Request) {
		entryID, err := parseID(request, "id")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid time entry id"))
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.get(request.Context(), identity.UserID, entryID)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})

	router.Put("/{id}", func(w http.ResponseWriter, request *http.Request) {
		entryID, err := parseID(request, "id")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid time entry id"))
			return
		}
		var req UpdateTimeEntryRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.update(request.Context(), identity.UserID, entryID, &req)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})

	router.Delete("/{id}", func(w http.ResponseWriter, request *http.Request) {
		entryID, err := parseID(request, "id")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid time entry id"))
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.delete(request.Context(), identity.UserID, entryID)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}
