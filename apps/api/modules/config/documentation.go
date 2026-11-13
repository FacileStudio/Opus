package config

import documentation "github.com/FacileStudio/Opus/apps/api/internal/documentation"

// Documentation describes the config module endpoints.
func Documentation() documentation.Module {
	return documentation.Module{
		Name:        "config",
		Description: "Public system settings and feature toggles.",
		Routes: []documentation.Route{
			{
				Method:       "GET",
				Path:         "/config",
				Summary:      "Get configuration",
				Description:  "Returns public application settings. customOAuthLogoutUrl is Registre's sign-out page when OIDC_ISSUER is set: the client sends the browser there after POST /auth/logout so leaving Opus also ends the Registre session. githubAppInstallUrl is the page where any GitHub account or organization installs the Facile Opus App; it is empty when the App is not configured or GitHub did not answer.",
				Auth:         "public",
				ResponseBody: `{"ssoOnly":true,"hasGuestAccess":true,"customOAuthLogoutUrl":"https://sso.facile.studio/logout","registreSpacesUrl":string,"githubAppInstallUrl":"https://github.com/apps/facile-opus/installations/new"}`,
			},
		},
	}
}
