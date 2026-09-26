// Package workspace declares the workspace and invitation HTTP surface. Both
// live in one module because invitations are how a workspace gains members;
// the routes are registered Go-native, ahead of the legacy proxy.
package workspace

import (
	"net/http"
	"strconv"

	"github.com/FacileStudio/Opus/apps/api/internal/authcontext"
	"github.com/FacileStudio/Opus/apps/api/internal/middleware"
	"github.com/FacileStudio/Opus/apps/api/modules/auth"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/go-chi/chi/v5"
)

func parseWorkspaceID(r *http.Request) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, "workspaceId"), 10, 64)
}

// RegisterRoutes wires the authenticated workspace and invitation endpoints.
func RegisterRoutes(router chi.Router, service *Service, authService *auth.Service) {
	router.Route("/workspace", func(router chi.Router) {
		router.Use(middleware.RequireAuth(authService))
		registerWorkspaceCollection(router, service)
		registerWorkspaceItems(router, service)
	})
	registerInvitationRoutes(router, service, authService)
}

func registerWorkspaceCollection(router chi.Router, service *Service) {
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
		var req CreateWorkspaceRequest
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

func registerWorkspaceItems(router chi.Router, service *Service) {
	router.Route("/{workspaceId}", func(router chi.Router) {
		registerWorkspaceLifecycle(router, service)
		registerWorkspaceMembers(router, service)
		registerWorkspaceLogo(router, service)
		registerWorkspaceInvitations(router, service)
	})
}

func registerWorkspaceLifecycle(router chi.Router, service *Service) {
	router.Patch("/", func(w http.ResponseWriter, request *http.Request) {
		var req UpdateWorkspaceRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		id, err := parseWorkspaceID(request)
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid workspace id"))
			return
		}
		resp, err := service.controller.update(request.Context(), identity.UserID, id, &req)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})

	router.Delete("/", func(w http.ResponseWriter, request *http.Request) {
		identity, _ := authcontext.IdentityFromContext(request.Context())
		id, err := parseWorkspaceID(request)
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid workspace id"))
			return
		}
		resp, err := service.controller.delete(request.Context(), identity.UserID, id)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})

	router.Get("/full", func(w http.ResponseWriter, request *http.Request) {
		identity, _ := authcontext.IdentityFromContext(request.Context())
		id, err := parseWorkspaceID(request)
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid workspace id"))
			return
		}
		resp, err := service.controller.full(request.Context(), identity.UserID, id)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

func registerWorkspaceMembers(router chi.Router, service *Service) {
	registerMemberQueries(router, service)
	registerMemberInvite(router, service)
	registerMemberUpdateAndDelete(router, service)
}

func registerMemberQueries(router chi.Router, service *Service) {
	router.Get("/members", func(w http.ResponseWriter, request *http.Request) {
		identity, _ := authcontext.IdentityFromContext(request.Context())
		id, err := parseWorkspaceID(request)
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid workspace id"))
			return
		}
		resp, err := service.controller.members(request.Context(), identity.UserID, id)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})

	router.Get("/members/me", func(w http.ResponseWriter, request *http.Request) {
		identity, _ := authcontext.IdentityFromContext(request.Context())
		id, err := parseWorkspaceID(request)
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid workspace id"))
			return
		}
		resp, err := service.controller.memberMe(request.Context(), identity.UserID, id)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

func registerMemberInvite(router chi.Router, service *Service) {
	router.Post("/members", func(w http.ResponseWriter, request *http.Request) {
		var req InviteMemberRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		id, err := parseWorkspaceID(request)
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid workspace id"))
			return
		}
		resp, err := service.controller.inviteMember(request.Context(), identity.UserID, id, &req)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusCreated, resp)
	})
}

func registerMemberUpdateAndDelete(router chi.Router, service *Service) {
	router.Patch("/members/{userId}", func(w http.ResponseWriter, request *http.Request) {
		var req UpdateMemberRoleRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		id, err := parseWorkspaceID(request)
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid workspace id"))
			return
		}
		userID, err := strconv.ParseInt(chi.URLParam(request, "userId"), 10, 64)
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid user id"))
			return
		}
		resp, err := service.controller.updateMemberRole(request.Context(), identity.UserID, id, userID, &req)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})

	router.Delete("/members/{memberIdOrEmail}", func(w http.ResponseWriter, request *http.Request) {
		identity, _ := authcontext.IdentityFromContext(request.Context())
		id, err := parseWorkspaceID(request)
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid workspace id"))
			return
		}
		resp, err := service.controller.removeMember(request.Context(), identity.UserID, id, chi.URLParam(request, "memberIdOrEmail"))
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

func registerWorkspaceLogo(router chi.Router, service *Service) {
	router.Post("/logo", func(w http.ResponseWriter, request *http.Request) {
		identity, _ := authcontext.IdentityFromContext(request.Context())
		id, err := parseWorkspaceID(request)
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid workspace id"))
			return
		}
		file, header, err := request.FormFile("logo")
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("No logo file provided. Send as multipart with field name 'logo'"))
			return
		}
		defer file.Close()
		resp, err := service.controller.uploadLogo(request.Context(), identity.UserID, id, file, header)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})

	router.Delete("/logo", func(w http.ResponseWriter, request *http.Request) {
		identity, _ := authcontext.IdentityFromContext(request.Context())
		id, err := parseWorkspaceID(request)
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid workspace id"))
			return
		}
		resp, err := service.controller.deleteLogo(request.Context(), identity.UserID, id)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})

	router.Get("/logo", func(w http.ResponseWriter, request *http.Request) {
		id, err := parseWorkspaceID(request)
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid workspace id"))
			return
		}
		if err := service.controller.serveLogo(request.Context(), id, w); err != nil {
			httpjson.WriteError(w, err)
			return
		}
	})
}

func registerWorkspaceInvitations(router chi.Router, service *Service) {
	router.Get("/invitations", func(w http.ResponseWriter, request *http.Request) {
		identity, _ := authcontext.IdentityFromContext(request.Context())
		id, err := parseWorkspaceID(request)
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid workspace id"))
			return
		}
		resp, err := service.controller.workspaceInvitations(request.Context(), identity.UserID, id)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}