package config

import "github.com/FacileStudio/Opus/apps/api/internal/env"

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
	if s.appEnv != nil && s.appEnv.OIDC != nil {
		hasOIDC = true
	}
	ssoOnly := false
	if s.appEnv != nil {
		ssoOnly = s.appEnv.SSOOnly
	}
	registreSpacesURL := ""
	if s.spacesURL != nil {
		registreSpacesURL = s.spacesURL()
	}
	return ConfigResponse{
		SSOOnly:                     ssoOnly,
		DisableRegistration:         false,
		DisablePasswordRegistration: ssoOnly,
		IsDemoMode:                  false,
		HasSMTP:                     false,
		HasGithubSignIn:             false,
		HasGoogleSignIn:             false,
		HasDiscordSignIn:            false,
		HasCustomOAuth:              hasOIDC,
		HasGuestAccess:              true,
		CustomOAuthLogoutURL:        nil,
		RegistreSpacesURL:           registreSpacesURL,
	}
}
