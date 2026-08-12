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

// RegisterRoutes wires the authenticated workspace and invitation endpoints,
// plus the unauthenticated logo reader.
func RegisterRoutes(router chi.Router, service *Service, authService *auth.Service) {
	router.Route("/workspace", func(router chi.Router) {
		router.Use(middleware.RequireAuth(authService))
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

		router.Route("/{workspaceId}", func(router chi.Router) {
			workspaceID := func(r *http.Request) (int64, error) {
				return strconv.ParseInt(chi.URLParam(r, "workspaceId"), 10, 64)
			}

			router.Patch("/", func(w http.ResponseWriter, request *http.Request) {
				var req UpdateWorkspaceRequest
				if err := httpjson.DecodeJSON(w, request, &req); err != nil {
					httpjson.WriteError(w, err)
					return
				}
				identity, _ := authcontext.IdentityFromContext(request.Context())
				id, err := workspaceID(request)
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
				id, err := workspaceID(request)
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
				id, err := workspaceID(request)
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

			router.Get("/invitations", func(w http.ResponseWriter, request *http.Request) {
				identity, _ := authcontext.IdentityFromContext(request.Context())
				id, err := workspaceID(request)
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

			router.Get("/members", func(w http.ResponseWriter, request *http.Request) {
				identity, _ := authcontext.IdentityFromContext(request.Context())
				id, err := workspaceID(request)
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
				id, err := workspaceID(request)
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

			router.Post("/members", func(w http.ResponseWriter, request *http.Request) {
				var req InviteMemberRequest
				if err := httpjson.DecodeJSON(w, request, &req); err != nil {
					httpjson.WriteError(w, err)
					return
				}
				identity, _ := authcontext.IdentityFromContext(request.Context())
				id, err := workspaceID(request)
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

			router.Patch("/members/{userId}", func(w http.ResponseWriter, request *http.Request) {
				var req UpdateMemberRoleRequest
				if err := httpjson.DecodeJSON(w, request, &req); err != nil {
					httpjson.WriteError(w, err)
					return
				}
				identity, _ := authcontext.IdentityFromContext(request.Context())
				id, err := workspaceID(request)
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
				id, err := workspaceID(request)
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

			router.Post("/logo", func(w http.ResponseWriter, request *http.Request) {
				identity, _ := authcontext.IdentityFromContext(request.Context())
				id, err := workspaceID(request)
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
				id, err := workspaceID(request)
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
				id, err := workspaceID(request)
				if err != nil {
					httpjson.WriteError(w, errors.Invalid("invalid workspace id"))
					return
				}
				if err := service.controller.serveLogo(request.Context(), id, w); err != nil {
					httpjson.WriteError(w, err)
					return
				}
			})
		})
	})

	router.Route("/invitation", func(router chi.Router) {
		router.Use(middleware.RequireAuth(authService))
		router.Get("/", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			resp, err := service.controller.userInvitations(request.Context(), identity.Email)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Route("/{invitationId}", func(router chi.Router) {
			invitationID := func(r *http.Request) (int64, error) {
				return strconv.ParseInt(chi.URLParam(r, "invitationId"), 10, 64)
			}

			router.Get("/", func(w http.ResponseWriter, request *http.Request) {
				id, err := invitationID(request)
				if err != nil {
					httpjson.WriteError(w, errors.Invalid("invalid invitation id"))
					return
				}
				resp, err := service.controller.invitationByID(request.Context(), id)
				if err != nil {
					httpjson.WriteError(w, err)
					return
				}
				httpjson.WriteJSON(w, http.StatusOK, resp)
			})

			router.Delete("/", func(w http.ResponseWriter, request *http.Request) {
				identity, _ := authcontext.IdentityFromContext(request.Context())
				id, err := invitationID(request)
				if err != nil {
					httpjson.WriteError(w, errors.Invalid("invalid invitation id"))
					return
				}
				resp, err := service.controller.cancelInvitation(request.Context(), identity.UserID, id)
				if err != nil {
					httpjson.WriteError(w, err)
					return
				}
				httpjson.WriteJSON(w, http.StatusOK, resp)
			})

			router.Post("/accept", func(w http.ResponseWriter, request *http.Request) {
				identity, _ := authcontext.IdentityFromContext(request.Context())
				id, err := invitationID(request)
				if err != nil {
					httpjson.WriteError(w, errors.Invalid("invalid invitation id"))
					return
				}
				resp, err := service.controller.acceptInvitation(request.Context(), identity.UserID, identity.Email, id)
				if err != nil {
					httpjson.WriteError(w, err)
					return
				}
				httpjson.WriteJSON(w, http.StatusOK, resp)
			})

			router.Post("/reject", func(w http.ResponseWriter, request *http.Request) {
				identity, _ := authcontext.IdentityFromContext(request.Context())
				id, err := invitationID(request)
				if err != nil {
					httpjson.WriteError(w, errors.Invalid("invalid invitation id"))
					return
				}
				resp, err := service.controller.rejectInvitation(request.Context(), identity.UserID, identity.Email, id)
				if err != nil {
					httpjson.WriteError(w, err)
					return
				}
				httpjson.WriteJSON(w, http.StatusOK, resp)
			})
		})
	})
}