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

func parseInvitationID(r *http.Request) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, "invitationId"), 10, 64)
}

func registerInvitationRoutes(router chi.Router, service *Service, authService *auth.Service) {
	router.Route("/invitation", func(router chi.Router) {
		router.Use(middleware.RequireAuth(authService))
		registerInvitationCollection(router, service)
		registerInvitationItem(router, service)
	})
}

func registerInvitationCollection(router chi.Router, service *Service) {
	router.Get("/", func(w http.ResponseWriter, request *http.Request) {
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.userInvitations(request.Context(), identity.Email)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

func registerInvitationItem(router chi.Router, service *Service) {
	router.Route("/{invitationId}", func(router chi.Router) {
		registerInvitationReadAndDelete(router, service)
		registerInvitationDecisions(router, service)
	})
}

func registerInvitationReadAndDelete(router chi.Router, service *Service) {
	router.Get("/", func(w http.ResponseWriter, request *http.Request) {
		identity, _ := authcontext.IdentityFromContext(request.Context())
		id, err := parseInvitationID(request)
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid invitation id"))
			return
		}
		resp, err := service.controller.invitationByID(request.Context(), identity.UserID, identity.Email, id)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})

	router.Delete("/", func(w http.ResponseWriter, request *http.Request) {
		identity, _ := authcontext.IdentityFromContext(request.Context())
		id, err := parseInvitationID(request)
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
}

func registerInvitationDecisions(router chi.Router, service *Service) {
	router.Post("/accept", func(w http.ResponseWriter, request *http.Request) {
		identity, _ := authcontext.IdentityFromContext(request.Context())
		id, err := parseInvitationID(request)
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
		id, err := parseInvitationID(request)
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
}
