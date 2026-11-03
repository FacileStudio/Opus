package config

// ConfigResponse represents the application configuration and feature flags.
// RegistreSpacesURL is set once Registre owns workspace management.
type ConfigResponse struct {
	SSOOnly                     bool    `json:"ssoOnly"`
	DisableRegistration         bool    `json:"disableRegistration"`
	DisablePasswordRegistration bool    `json:"disablePasswordRegistration"`
	IsDemoMode                  bool    `json:"isDemoMode"`
	HasSMTP                     bool    `json:"hasSmtp"`
	HasGithubSignIn             bool    `json:"hasGithubSignIn"`
	HasGoogleSignIn             bool    `json:"hasGoogleSignIn"`
	HasDiscordSignIn            bool    `json:"hasDiscordSignIn"`
	HasCustomOAuth              bool    `json:"hasCustomOAuth"`
	HasGuestAccess              bool    `json:"hasGuestAccess"`
	CustomOAuthLogoutURL        *string `json:"customOAuthLogoutUrl"`
	RegistreSpacesURL           string  `json:"registreSpacesUrl"`
}
