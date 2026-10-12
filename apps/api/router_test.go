package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FacileStudio/tronc/health"
	"github.com/go-chi/chi/v5"
)

func TestHealthEndpoints(t *testing.T) {
	r := chi.NewRouter()
	health.Mount(r)
	r.Head("/health", health.Live())
	r.Head("/api/health", health.Live())
	r.Route("/api", func(api chi.Router) {
		health.Mount(api)
		api.Head("/health", health.Live())
	})

	endpoints := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/health"},
		{http.MethodHead, "/health"},
		{http.MethodGet, "/api/health"},
		{http.MethodHead, "/api/health"},
	}

	for _, ep := range endpoints {
		req := httptest.NewRequest(ep.method, ep.path, nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("%s %s expected 200, got %d", ep.method, ep.path, rec.Code)
		}
	}
}

func TestReferenceConfig(t *testing.T) {
	cfg := referenceConfig()
	if cfg.Title != "Opus API" {
		t.Fatalf("expected title 'Opus API', got '%s'", cfg.Title)
	}

	expectedModules := map[string]bool{
		"auth":         false,
		"workspace":    false,
		"invitation":   false,
		"project":      false,
		"column":       false,
		"task":         false,
		"taskrelation": false,
		"label":        false,
		"comment":      false,
		"activity":     false,
		"sharelink":    false,
		"timeentry":    false,
		"workflowrule": false,
		"notification": false,
		"search":       false,
	}

	for _, mod := range cfg.Registry.Modules {
		expectedModules[mod.Name] = true
		if len(mod.Routes) == 0 {
			t.Errorf("module %s has 0 routes registered", mod.Name)
		}
	}

	for name, found := range expectedModules {
		if !found {
			t.Errorf("expected module %s was not found in documentation registry", name)
		}
	}
}
