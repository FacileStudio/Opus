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
				Description:  "Returns public application settings.",
				Auth:         "public",
				ResponseBody: `{"ssoOnly":false,"hasGuestAccess":true}`,
			},
		},
	}
}
