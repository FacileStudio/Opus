package config

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FacileStudio/Opus/apps/api/internal/env"
	"github.com/go-chi/chi/v5"
)

func TestConfigGet(t *testing.T) {
	appEnv := &env.Config{OIDC: &env.OIDCConfig{Issuer: "https://sso.test"}}
	service := NewService(appEnv)
	service.SetGitHubAppInstallURL(func(context.Context) string {
		return "https://github.com/apps/facile-opus/installations/new"
	})
	r := chi.NewRouter()
	RegisterRoutes(r, service)

	req := httptest.NewRequest(http.MethodGet, "/config", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp ConfigResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.SSOOnly || !resp.DisableRegistration {
		t.Fatal("expected an issuer to close local sign-in and registration")
	}
	if resp.CustomOAuthLogoutURL == nil || *resp.CustomOAuthLogoutURL != "https://sso.test/logout" {
		t.Fatalf("expected Registre's logout page, got %v", resp.CustomOAuthLogoutURL)
	}
	if resp.GitHubAppInstallURL != "https://github.com/apps/facile-opus/installations/new" {
		t.Fatalf("expected the GitHub App install page, got %q", resp.GitHubAppInstallURL)
	}
	if !resp.HasGuestAccess {
		t.Fatal("expected HasGuestAccess to be true")
	}
}
