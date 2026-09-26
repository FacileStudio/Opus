package comment

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

// RegisterRoutes wires the authenticated comment endpoints.
func RegisterRoutes(router chi.Router, service *Service, authService middleware.Authenticator) {
	router.Route("/comment", func(router chi.Router) {
		router.Use(middleware.RequireAuth(authService))
		registerCommentCollection(router, service)
		registerCommentItem(router, service)
	})
}

func registerCommentCollection(router chi.Router, service *Service) {
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

	router.Post("/{taskId}", func(w http.ResponseWriter, request *http.Request) {
		taskID, err := parseID(request, "taskId")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid task id"))
			return
		}
		var req CreateCommentRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.create(request.Context(), identity.UserID, taskID, &req)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

func registerCommentItem(router chi.Router, service *Service) {
	router.Put("/{id}", func(w http.ResponseWriter, request *http.Request) {
		commentID, err := parseID(request, "id")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid comment id"))
			return
		}
		var req UpdateCommentRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.update(request.Context(), identity.UserID, commentID, &req)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})

	router.Delete("/{id}", func(w http.ResponseWriter, request *http.Request) {
		commentID, err := parseID(request, "id")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid comment id"))
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.delete(request.Context(), identity.UserID, commentID)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}
