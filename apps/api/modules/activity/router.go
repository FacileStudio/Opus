package activity

import (
	"net/http"
	"strconv"

	"github.com/FacileStudio/Opus/apps/api/internal/authcontext"
	"github.com/FacileStudio/Opus/apps/api/internal/middleware"
	"github.com/FacileStudio/Opus/apps/api/internal/wireid"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/go-chi/chi/v5"
)

func parseID(r *http.Request, key string) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, key), 10, 64)
}

// RegisterRoutes registers the authenticated activity HTTP routes.
func RegisterRoutes(router chi.Router, service *Service, authService middleware.Authenticator) {
	router.Route("/activity", func(router chi.Router) {
		router.Use(middleware.RequireAuth(authService))
		registerActivityCollection(router, service)
		registerActivityComment(router, service)
	})
}

func registerActivityCollection(router chi.Router, service *Service) {
	router.Get("/{taskId}", handleGetActivities(service))
	router.Post("/create", handleCreateActivity(service))
}

func handleGetActivities(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
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
	}
}

func handleCreateActivity(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		var req CreateActivityRequest
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
	}
}

func registerActivityComment(router chi.Router, service *Service) {
	router.Post("/comment", handlePostComment(service))
	router.Put("/comment", handlePutComment(service))
	router.Delete("/comment", handleDeleteComment(service))
}

func handlePostComment(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		var req CommentActivityRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.createComment(request.Context(), identity.UserID, &req)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	}
}

func handlePutComment(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		var req UpdateCommentActivityRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.updateComment(request.Context(), identity.UserID, &req)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	}
}

func handleDeleteComment(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		var req DeleteCommentActivityRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		if req.ActivityID <= 0 {
			if qID := request.URL.Query().Get("activityId"); qID != "" {
				if parsed, err := strconv.ParseInt(qID, 10, 64); err == nil {
					req.ActivityID = wireid.ID(parsed)
				}
			}
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.deleteComment(request.Context(), identity.UserID, &req)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	}
}
