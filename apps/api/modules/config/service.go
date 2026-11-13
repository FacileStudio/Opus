package config

import (
	"context"
	"strings"

	"github.com/FacileStudio/Opus/apps/api/internal/env"
)

// Service provides application configuration state.
type Service struct {
	appEnv     *env.Config
	controller *controller
	spacesURL  func() string
	installURL func(context.Context) string
}

// NewService constructs a new configuration service.
func NewService(appEnv *env.Config) *Service {
	s := &Service{appEnv: appEnv}
	s.controller = newController(s)
	return s
}

// SetSpacesURL wires the source of the Registre space page the frontend links
// to once Registre owns workspaces.
func (s *Service) SetSpacesURL(spacesURL func() string) {
	s.spacesURL = spacesURL
}

// SetGitHubAppInstallURL wires the source of the page where a GitHub account
// or organization installs the Facile Opus App.
func (s *Service) SetGitHubAppInstallURL(installURL func(context.Context) string) {
	s.installURL = installURL
}

// Get returns the public application configuration values.
func (s *Service) Get(ctx context.Context) ConfigResponse {
	hasOIDC := false
	var logoutURL *string
	if s.appEnv != nil && s.appEnv.OIDC != nil {
		hasOIDC = true
		logoutURL = registreLogoutURL(s.appEnv.OIDC.Issuer)
	}
	registreSpacesURL := ""
	if s.spacesURL != nil {
		registreSpacesURL = s.spacesURL()
	}
	githubAppInstallURL := ""
	if s.installURL != nil {
		githubAppInstallURL = s.installURL(ctx)
	}
	return ConfigResponse{
		SSOOnly:                     hasOIDC,
		DisableRegistration:         true,
		DisablePasswordRegistration: true,
		IsDemoMode:                  false,
		HasSMTP:                     false,
		HasGithubSignIn:             false,
		HasGoogleSignIn:             false,
		HasDiscordSignIn:            false,
		HasCustomOAuth:              hasOIDC,
		HasGuestAccess:              true,
		CustomOAuthLogoutURL:        logoutURL,
		RegistreSpacesURL:           registreSpacesURL,
		GitHubAppInstallURL:         githubAppInstallURL,
	}
}

// registreLogoutURL is Registre's own sign-out page, which ends the Registre
// session so the next sign-in cannot silently pick it up again. porte keeps no
// ID token, so the id_token_hint that end_session needs is not available.
func registreLogoutURL(issuer string) *string {
	if issuer == "" {
		return nil
	}
	url := strings.TrimRight(issuer, "/") + "/logout"
	return &url
}
