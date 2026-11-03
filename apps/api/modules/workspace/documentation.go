package workspace

import documentation "github.com/FacileStudio/Opus/apps/api/internal/documentation"

func workspaceCrudRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "GET",
			Path:         "/workspace",
			Summary:      "List the caller's workspaces",
			Description:  "Every workspace the authenticated user is a member of, ordered by name.",
			Auth:         "bearer",
			ResponseBody: `[{"id":int64,"name":string,"slug":string,"logo":string,"description":string,"createdAt":string,"registreSpaceId":string?}]`,
		},
		{
			Method:       "POST",
			Path:         "/workspace",
			Summary:      "Create a workspace",
			Description:  "The caller becomes the owner member in the same transaction. The slug defaults to a normalized form of the name and is made unique with a numeric suffix when taken.",
			Auth:         "bearer",
			RequestBody:  `{"name":string,"slug":string?,"description":string?}`,
			ResponseBody: `{"id":int64,"name":string,"slug":string,"logo":string,"description":string,"createdAt":string,"registreSpaceId":string?}`,
		},
		{
			Method:      "PATCH",
			Path:        "/workspace/{workspaceId}",
			Summary:     "Update a workspace",
			Description: "Owners and admins may edit. Only the keys present in the body are written; the slug must stay unique.",
			Auth:        "bearer",
			PathParams:  []documentation.Field{{Name: "workspaceId", Type: "int64", Description: "The workspace's id"}},
			RequestBody: `{"name":string?,"slug":string?,"description":string?}`,
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "The caller is not an owner or admin"},
				{Status: 409, Code: "already_exists", Description: "The slug belongs to another workspace"},
			},
		},
		{
			Method:      "DELETE",
			Path:        "/workspace/{workspaceId}",
			Summary:     "Delete a workspace",
			Description: "Owners only. Removes the memberships and invitations first, then the workspace row.",
			Auth:        "bearer",
			PathParams:  []documentation.Field{{Name: "workspaceId", Type: "int64", Description: "The workspace's id"}},
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "The caller is not the owner"},
			},
		},
		{
			Method:       "GET",
			Path:         "/workspace/{workspaceId}/full",
			Summary:      "Return a workspace and its members",
			Description:  "The workspace row and every member, the shape the organization screen renders.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "workspaceId", Type: "int64", Description: "The workspace's id"}},
			ResponseBody: `{"workspace":{...},"members":[{...}]}`,
		},
	}
}

func workspaceMemberRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "GET",
			Path:         "/workspace/{workspaceId}/members",
			Summary:      "List a workspace's members",
			Description:  "Member rows joined to their user profiles. The caller must be a member of the workspace.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "workspaceId", Type: "int64", Description: "The workspace's id"}},
			ResponseBody: `[{"id":int64,"name":string,"email":string,"image":string,"role":string}]`,
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "The caller is not a member of this workspace"},
			},
		},
		{
			Method:      "GET",
			Path:        "/workspace/{workspaceId}/members/me",
			Summary:     "Return the caller's membership row",
			Description: "The caller's own member row in the workspace, joined to their profile.",
			Auth:        "bearer",
			PathParams:  []documentation.Field{{Name: "workspaceId", Type: "int64", Description: "The workspace's id"}},
		},
		{
			Method:      "PATCH",
			Path:        "/workspace/{workspaceId}/members/{userId}",
			Summary:     "Change a member's role",
			Description: "Owners only. Nobody may change their own role and the last owner cannot be demoted.",
			Auth:        "bearer",
			PathParams:  []documentation.Field{{Name: "workspaceId", Type: "int64", Description: "The workspace's id"}, {Name: "userId", Type: "int64", Description: "The member's user id"}},
			RequestBody: `{"role":"owner"|"admin"|"member"}`,
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "The caller is not the owner"},
				{Status: 400, Code: "invalid_argument", Description: "Demoting the last owner or one's own role"},
			},
		},
		{
			Method:      "DELETE",
			Path:        "/workspace/{workspaceId}/members/{memberIdOrEmail}",
			Summary:     "Remove a member",
			Description: "Owners and admins may. The target is a user id or an email. The last owner cannot be removed.",
			Auth:        "bearer",
			PathParams:  []documentation.Field{{Name: "workspaceId", Type: "int64", Description: "The workspace's id"}, {Name: "memberIdOrEmail", Type: "string", Description: "The member's user id or email"}},
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "The caller is not an owner or admin"},
				{Status: 400, Code: "invalid_argument", Description: "Removing the last owner"},
			},
		},
	}
}

func workspaceLogoRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:      "POST",
			Path:        "/workspace/{workspaceId}/logo",
			Summary:     "Upload a workspace logo",
			Description: "Multipart form with a 'logo' file field. Allowed types: png, jpeg, webp, svg, gif. Maximum 2MB. Stored under UPLOAD_DIR/workspaces/{id} and served at GET /workspace/{workspaceId}/logo.",
			Auth:        "bearer",
			PathParams:  []documentation.Field{{Name: "workspaceId", Type: "int64", Description: "The workspace's id"}},
			RequestBody: "multipart/form-data: logo=<image file>",
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "The caller is not a member of this workspace"},
				{Status: 400, Code: "invalid_argument", Description: "Missing logo field or disallowed image type"},
				{Status: 413, Code: "resource_exhausted", Description: "File larger than 2MB"},
			},
		},
		{
			Method:      "DELETE",
			Path:        "/workspace/{workspaceId}/logo",
			Summary:     "Delete a workspace logo",
			Description: "Removes the stored files and clears the logo column.",
			Auth:        "bearer",
			PathParams:  []documentation.Field{{Name: "workspaceId", Type: "int64", Description: "The workspace's id"}},
		},
		{
			Method:      "GET",
			Path:        "/workspace/{workspaceId}/logo",
			Summary:     "Serve a workspace logo",
			Description: "The stored image bytes, cache-busted by the caller's ?v= stamp. Public, matching the legacy API.",
			Auth:        "public",
			PathParams:  []documentation.Field{{Name: "workspaceId", Type: "int64", Description: "The workspace's id"}},
			Errors: []documentation.Error{
				{Status: 404, Code: "not_found", Description: "The workspace has no logo stored"},
			},
		},
	}
}

// Documentation returns the workspace module API reference metadata.
func Documentation() documentation.Module {
	return documentation.Module{
		Name:        "workspace",
		Description: "Workspaces are the tenant boundary: a team's projects, columns and tasks live under one, and a user's membership in one is what the workspace-scoped access checks gate on. When OIDC is configured, Registre owns them: its shared spaces are imported once and then synced, a linked workspace (registreSpaceId set) refuses rename, delete, invitations and member changes with 409 and a link to its Registre page, and creating a workspace is refused once the import has succeeded.",
		Routes: append(workspaceCrudRoutes(), append(workspaceMemberRoutes(),
			append(workspaceLogoRoutes(), registreSyncRoute())...)...),
	}
}

// registreSyncRoute documents the nudge Registre posts after any change to a
// shared space.
func registreSyncRoute() documentation.Route {
	return documentation.Route{
		Method:      "POST",
		Path:        "/registre/sync",
		Summary:     "Nudge a Registre directory sync",
		Description: "Registre posts an empty body after any shared space change. Opus answers 202 and pulls the authenticated snapshot about two seconds later; the nudge itself is not trusted. Mounted only when OIDC is configured.",
		Auth:        "public",
	}
}
