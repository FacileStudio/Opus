package auth

import (
	"net/http"

	"github.com/FacileStudio/Opus/apps/api/internal/authcontext"
	"github.com/FacileStudio/Opus/apps/api/internal/env"
	"github.com/FacileStudio/Opus/apps/api/internal/middleware"
	"github.com/FacileStudio/tronc/httpjson"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts what porte does not own.
//
// /auth/config, /auth/logout, /auth/oidc and the rest of the OIDC flow come
// from porte's session manager and its OIDC kit, mounted in main.go. What is
// left here is the local password path (keeping Opus's {token, user} shape)
// and /auth/me, the profile answer the frontend reads after a session exists.
//
// The local kit is deliberately NOT Mount-ed here: those handlers answer porte's
// lean {user_id, token}, and Opus's frontend expects {token, user}.
//
// Under SSO_ONLY the credential routes are not registered rather than
// rejected, so there is no endpoint left to probe for an account.
func RegisterRoutes(router chi.Router, service *Service, appEnv env.Config) {
	router.Route("/auth", func(router chi.Router) {
		router.Group(func(router chi.Router) {
			router.Use(middleware.RequireAuth(service))
			router.Get("/me", func(w http.ResponseWriter, request *http.Request) {
				identity, _ := authcontext.IdentityFromContext(request.Context())
				resp, err := service.controller.me(request.Context(), identity.UserID)
				if err != nil {
					httpjson.WriteError(w, err)
					return
				}
				httpjson.WriteJSON(w, http.StatusOK, resp)
			})

			router.Patch("/me", func(w http.ResponseWriter, request *http.Request) {
				var req UpdateProfileRequest
				if err := httpjson.DecodeJSON(w, request, &req); err != nil {
					httpjson.WriteError(w, err)
					return
				}
				identity, _ := authcontext.IdentityFromContext(request.Context())
				resp, err := service.controller.updateProfile(w, request, identity.UserID, &req)
				if err != nil {
					httpjson.WriteError(w, err)
					return
				}
				httpjson.WriteJSON(w, http.StatusOK, resp)
			})

			router.Post("/change-password", func(w http.ResponseWriter, request *http.Request) {
				var req ChangePasswordRequest
				if err := httpjson.DecodeJSON(w, request, &req); err != nil {
					httpjson.WriteError(w, err)
					return
				}
				identity, _ := authcontext.IdentityFromContext(request.Context())
				resp, err := service.controller.changePassword(w, request, identity.UserID, &req)
				if err != nil {
					httpjson.WriteError(w, err)
					return
				}
				httpjson.WriteJSON(w, http.StatusOK, resp)
			})
		})

		if appEnv.SSOOnly {
			return
		}
		router.Post("/register", func(w http.ResponseWriter, request *http.Request) {
			var req RegisterRequest
			if err := httpjson.DecodeJSON(w, request, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			resp, err := service.controller.register(w, request, &req)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusCreated, resp)
		})

		router.Post("/login", func(w http.ResponseWriter, request *http.Request) {
			var req LoginRequest
			if err := httpjson.DecodeJSON(w, request, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			resp, err := service.controller.login(w, request, &req)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})
	})
}