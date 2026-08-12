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

// RegisterRoutes wires the authenticated workspace endpoints, plus the
// unauthenticated logo reader. The logo GET is deliberately public: the legacy
// API served it with no access check, and logo bytes are not confidential.
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

		router.Route("/{workspaceId}", func(router chi.Router) {
			workspaceID := func(r *http.Request) (int64, error) {
				return strconv.ParseInt(chi.URLParam(r, "workspaceId"), 10, 64)
			}

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
}