package auth

import documentation "github.com/FacileStudio/Opus/apps/api/internal/documentation"

func oidcRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "POST",
			Path:         "/auth/logout",
			Summary:      "End the current session",
			Description:  "Served by porte. Revokes the one session this request authenticated with, by id, and clears the session cookie.",
			Auth:         "bearer",
			ResponseBody: `{"logged_out":true}`,
		},
		{
			Method:      "GET",
			Path:        "/auth/oidc",
			Summary:     "Start the single sign-on flow",
			Description: "Registered only when OIDC_ISSUER is set. Redirects to the identity provider with PKCE, a nonce and a state.",
			Auth:        "public",
		},
		{
			Method:      "GET",
			Path:        "/auth/oidc/callback",
			Summary:     "Finish the single sign-on flow",
			Description: "The provider's redirect target. Verifies the state and the ID token, upserts the account, sets the session cookie and redirects to OIDC_SUCCESS_URL.",
			Auth:        "public",
		},
		{
			Method:       "POST",
			Path:         "/auth/backchannel-logout",
			Summary:      "Sign a person out on Registre's word",
			Description:  "Served by porte, registered only when OIDC_ISSUER is set. OpenID Connect Back-Channel Logout 1.0: Registre posts a signed logout token as the form field logout_token when the person signs out there. A valid token revokes every Opus session the subject holds. An unknown subject also answers 200; a bad token answers 401.",
			Auth:         "public",
			RequestBody:  `logout_token=<jwt> (application/x-www-form-urlencoded)`,
			ResponseBody: `{"logged_out":true}`,
		},
		{
			Method:       "POST",
			Path:         "/auth/oidc/exchange",
			Summary:      "Trade a CLI login code for a session token",
			Description:  "The other half of ?flow=cli. The code is single use, hashed at rest and valid for 60 seconds.",
			Auth:         "public",
			RequestBody:  `{"code":string}`,
			ResponseBody: `{"user_id":string,"token":string}`,
		},
		{
			Method:      "GET",
			Path:        "/auth/config",
			Summary:     "Describe the auth methods on offer",
			Description: "Returns sso_only and oidc_enabled. With an issuer configured sso_only is always true: there is no password sign-in.",
			Auth:        "public",
		},
	}
}

func profileRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "GET",
			Path:         "/auth/me",
			Summary:      "Return the authenticated caller's profile",
			Description:  "The profile row behind the session porte verified. One query per authenticated request.",
			Auth:         "bearer",
			ResponseBody: `{"id":int64,"name":string,"email":string,"emailVerified":bool,"image":string,"color":string,"locale":string,"createdAt":string,"updatedAt":string}`,
		},
		{
			Method:       "PATCH",
			Path:         "/auth/me",
			Summary:      "Update the caller's own profile",
			Description:  "Writes only the keys present in the body. A cookie-authenticated mutation must carry X-Facile-CSRF.",
			Auth:         "bearer",
			RequestBody:  `{"name":string?,"color":string?,"locale":string?}`,
			ResponseBody: `{"token":string,"user":{...}}`,
		},
	}
}

func tokenRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "GET",
			Path:         "/auth/api-tokens",
			Summary:      "List the caller's personal API tokens",
			Description:  "Returns the named tokens only. Browser logins share the session store but are not tokens and are not listed here.",
			Auth:         "bearer",
			ResponseBody: `{"tokens":[{"id":int64,"name":string,"createdAt":string,"lastUsedAt":string}]}`,
		},
		{
			Method:       "POST",
			Path:         "/auth/api-tokens",
			Summary:      "Mint a personal API token",
			Description:  "Creates a named token that never expires and authenticates as the caller on every route a session does. The plaintext value is returned once and cannot be read again. A cookie-authenticated mutation must carry X-Facile-CSRF.",
			Auth:         "bearer",
			RequestBody:  `{"name":string}`,
			ResponseBody: `{"id":int64,"name":string,"createdAt":string,"lastUsedAt":string,"token":string}`,
		},
		{
			Method:       "DELETE",
			Path:         "/auth/api-tokens/{id}",
			Summary:      "Revoke a personal API token",
			Description:  "Ends one named token by id. Refuses an id that is not a named token of the caller. A cookie-authenticated mutation must carry X-Facile-CSRF.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int64", Description: "Token id"}},
			ResponseBody: `{"deleted":true}`,
			Errors: []documentation.Error{
				{Status: 404, Code: "not_found", Description: "No such named token for this caller"},
			},
		},
	}
}

// Documentation returns the auth module API reference metadata.
func Documentation() documentation.Module {
	routes := make([]documentation.Route, 0)
	routes = append(routes, oidcRoutes()...)
	routes = append(routes, profileRoutes()...)
	routes = append(routes, tokenRoutes()...)
	return documentation.Module{
		Name:        "auth",
		Description: "Authentication, through the suite's porte kit. A session is a random 32 byte token stored SHA-256 hashed, sent as Authorization: Bearer <token> or, after a single sign-on login, as an HttpOnly __Host-session cookie. A cookie-authenticated request that is not GET, HEAD or OPTIONS must also carry an X-Facile-CSRF header with any non-empty value. A personal API token is that same session with a label, so it authenticates everywhere a login does and is revoked from /auth/api-tokens.",
		Routes:      routes,
	}
}
