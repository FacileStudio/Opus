package config

import (
	"strings"

	"github.com/FacileStudio/Opus/apps/api/internal/env"
)

// Service provides application configuration state.
type Service struct {
	appEnv     *env.Config
	controller *controller
	spacesURL  func() string
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

// Get returns the public application configuration values.
func (s *Service) Get() ConfigResponse {
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
