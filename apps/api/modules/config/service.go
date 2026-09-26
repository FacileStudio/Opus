package config

import "github.com/FacileStudio/Opus/apps/api/internal/env"

// Service provides application configuration state.
type Service struct {
	appEnv     *env.Config
	controller *controller
}

// NewService constructs a new configuration service.
func NewService(appEnv *env.Config) *Service {
	s := &Service{appEnv: appEnv}
	s.controller = newController(s)
	return s
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
	return ConfigResponse{
		SSOOnly:                     ssoOnly,
		DisableRegistration:        false,
		DisablePasswordRegistration: ssoOnly,
		IsDemoMode:                  false,
		HasSMTP:                     false,
		HasGithubSignIn:             false,
		HasGoogleSignIn:             false,
		HasDiscordSignIn:            false,
		HasCustomOAuth:              hasOIDC,
		HasGuestAccess:              true,
		CustomOAuthLogoutURL:        nil,
	}
}
