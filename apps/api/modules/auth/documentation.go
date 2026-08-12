package auth

import documentation "github.com/FacileStudio/Opus/apps/api/internal/documentation"

var Documentation = documentation.Module{
	Name:        "auth",
	Description: "Authentication, through the suite's porte kit. A session is a random 32 byte token stored SHA-256 hashed, sent as Authorization: Bearer <token> or, after a single sign-on login, as an HttpOnly __Host-session cookie. A cookie-authenticated request that is not GET, HEAD or OPTIONS must also carry an X-Facile-CSRF header with any non-empty value.",
	Routes: []documentation.Route{
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
			Description: "Returns sso_only and oidc_enabled, so the client knows whether to show the password form.",
			Auth:        "public",
		},
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
		{
			Method:      "POST",
			Path:        "/auth/change-password",
			Summary:     "Verify the current password and set a new one",
			Description: "porte verifies the current password, replaces the hash, revokes every other session and rotates the caller's own session. A cookie-authenticated mutation must carry X-Facile-CSRF.",
			Auth:        "bearer",
			RequestBody: `{"currentPassword":string,"newPassword":string}`,
			Errors: []documentation.Error{
				{Status: 401, Code: "unauthenticated", Description: "The current password is wrong"},
				{Status: 400, Code: "invalid_argument", Description: "The account has no password, or the new one is too short"},
			},
		},
		{
			Method:       "POST",
			Path:         "/auth/register",
			Summary:      "Create a local account and sign in",
			Description:  "Registered only when SSO_ONLY is false. porte hashes the password, sets the session cookie and returns a bearer token with the profile.",
			Auth:         "public",
			RequestBody:  `{"email":string,"name":string,"password":string}`,
			ResponseBody: `{"token":string,"user":{...}}`,
			Errors: []documentation.Error{
				{Status: 409, Code: "already_exists", Description: "An account with this email already exists"},
				{Status: 422, Code: "failed_precondition", Description: "Registration is disabled, the password is too short, or the email is invalid"},
			},
		},
		{
			Method:       "POST",
			Path:         "/auth/login",
			Summary:      "Sign in with a local password",
			Description:  "Verified by porte's local kit at suite parameters. Wrong password and unknown address both answer unauthorized, so the endpoint is not an enumeration oracle.",
			Auth:         "public",
			RequestBody:  `{"email":string,"password":string}`,
			ResponseBody: `{"token":string,"user":{...}}`,
			Errors: []documentation.Error{
				{Status: 401, Code: "unauthenticated", Description: "Invalid email or password"},
			},
		},
	},
}
