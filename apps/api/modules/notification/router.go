package notification

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

// RegisterRoutes mounts notification and preference endpoints.
func RegisterRoutes(router chi.Router, service *Service, authService middleware.Authenticator) {
	router.Route("/notification", func(router chi.Router) {
		router.Use(middleware.RequireAuth(authService))
		registerNotificationRoutes(router, service)
	})
	router.Route("/notifications", func(router chi.Router) {
		router.Use(middleware.RequireAuth(authService))
		registerNotificationRoutes(router, service)
	})
	router.Route("/notification-preferences", func(router chi.Router) {
		router.Use(middleware.RequireAuth(authService))
		registerPreferenceRoutes(router, service)
	})
}

func registerNotificationRoutes(router chi.Router, service *Service) {
	registerNotificationCollection(router, service)
	registerNotificationMutations(router, service)
}

func registerNotificationCollection(router chi.Router, service *Service) {
	router.Get("/", func(w http.ResponseWriter, request *http.Request) {
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.list(request.Context(), identity.UserID)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})

	router.Post("/", func(w http.ResponseWriter, request *http.Request) {
		var req CreateNotificationRequest
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

func registerNotificationMutations(router chi.Router, service *Service) {
	router.Patch("/{id}/read", func(w http.ResponseWriter, request *http.Request) {
		notifID, err := parseID(request, "id")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid notification id"))
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.markAsRead(request.Context(), identity.UserID, notifID)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})

	router.Patch("/{id}/unread", func(w http.ResponseWriter, request *http.Request) {
		notifID, err := parseID(request, "id")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid notification id"))
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.markAsUnread(request.Context(), identity.UserID, notifID)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})

	router.Delete("/{id}", func(w http.ResponseWriter, request *http.Request) {
		notifID, err := parseID(request, "id")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid notification id"))
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.remove(request.Context(), identity.UserID, notifID)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})

	router.Patch("/read-all", func(w http.ResponseWriter, request *http.Request) {
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.markAllAsRead(request.Context(), identity.UserID)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})

	router.Delete("/clear-all", func(w http.ResponseWriter, request *http.Request) {
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.clearAll(request.Context(), identity.UserID)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

func registerPreferenceRoutes(router chi.Router, service *Service) {
	registerGlobalPreferences(router, service)
	registerWorkspacePreferences(router, service)
}

func registerGlobalPreferences(router chi.Router, service *Service) {
	router.Get("/", func(w http.ResponseWriter, request *http.Request) {
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.getPreferences(request.Context(), identity.UserID, "")
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})

	router.Put("/", func(w http.ResponseWriter, request *http.Request) {
		var req UpdatePreferencesRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.updatePreferences(request.Context(), identity.UserID, "", &req)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

func registerWorkspacePreferences(router chi.Router, service *Service) {
	router.Put("/workspaces/{workspaceId}", func(w http.ResponseWriter, request *http.Request) {
		wsID, err := parseID(request, "workspaceId")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid workspace id"))
			return
		}
		var req UpsertWorkspaceRuleRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.upsertWorkspaceRule(request.Context(), identity.UserID, wsID, "", &req)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})

	router.Delete("/workspaces/{workspaceId}", func(w http.ResponseWriter, request *http.Request) {
		wsID, err := parseID(request, "workspaceId")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid workspace id"))
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.deleteWorkspaceRule(request.Context(), identity.UserID, wsID, "")
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}
