package config

import (
	"net/http"

	"github.com/FacileStudio/tronc/httpjson"
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts the public configuration route onto the router.
func RegisterRoutes(router chi.Router, service *Service) {
	router.Get("/config", func(w http.ResponseWriter, r *http.Request) {
		httpjson.WriteJSON(w, http.StatusOK, service.controller.get(r.Context()))
	})
}
