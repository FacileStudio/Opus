package workspace

import documentation "github.com/FacileStudio/Opus/apps/api/internal/documentation"

func workspaceInviteRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "POST",
			Path:         "/workspace/{workspaceId}/members",
			Summary:      "Invite someone by email",
			Description:  "Owners and admins invite. Creates a pending invitation, never a member row, with an expiry seven days out. The invitee accepts to join.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "workspaceId", Type: "int64", Description: "The workspace's id"}},
			RequestBody:  `{"email":string,"role":"owner"|"admin"|"member"?}`,
			ResponseBody: `{"id":int64,"email":string,"role":string,"status":string,"createdAt":string,"expiresAt":string?"workspaceId":int64,"workspaceName":string,"inviterName":string}`,
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "The caller is not an owner or admin"},
				{Status: 409, Code: "already_exists", Description: "Already a member, or an invitation is already pending"},
			},
		},
		{
			Method:      "GET",
			Path:        "/workspace/{workspaceId}/invitations",
			Summary:     "List a workspace's invitations",
			Description: "Every invitation for the workspace, any status, newest first.",
			Auth:        "bearer",
			PathParams:  []documentation.Field{{Name: "workspaceId", Type: "int64", Description: "The workspace's id"}},
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "The caller is not a member"},
			},
		},
	}
}

func userInviteRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "GET",
			Path:         "/invitation",
			Summary:      "List the caller's pending invitations",
			Description:  "Invitations addressed to the caller's email that are still pending.",
			Auth:         "bearer",
			ResponseBody: `[{"id":int64,"email":string,"role":string,"status":string,"workspaceId":int64,"workspaceName":string,"inviterName":string}]`,
		},
		{
			Method:      "GET",
			Path:        "/invitation/{invitationId}",
			Summary:     "Load one invitation",
			Description: "One invitation, joined to its workspace and inviter, for the accept page and getInvitation.",
			Auth:        "bearer",
			PathParams:  []documentation.Field{{Name: "invitationId", Type: "int64", Description: "The invitation's id"}},
			Errors: []documentation.Error{
				{Status: 404, Code: "not_found", Description: "No such invitation"},
			},
		},
		{
			Method:      "POST",
			Path:        "/invitation/{invitationId}/accept",
			Summary:     "Accept an invitation",
			Description: "The invited address accepts: the invitation moves to accepted and a member row is created with the invitation's role.",
			Auth:        "bearer",
			PathParams:  []documentation.Field{{Name: "invitationId", Type: "int64", Description: "The invitation's id"}},
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "The invitation was not sent to the caller"},
				{Status: 409, Code: "already_exists", Description: "Already accepted, canceled, or the caller is already a member"},
			},
		},
		{
			Method:      "POST",
			Path:        "/invitation/{invitationId}/reject",
			Summary:     "Reject an invitation",
			Description: "The invited address declines; the invitation is marked rejected.",
			Auth:        "bearer",
			PathParams:  []documentation.Field{{Name: "invitationId", Type: "int64", Description: "The invitation's id"}},
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "The invitation was not sent to the caller"},
			},
		},
		{
			Method:      "DELETE",
			Path:        "/invitation/{invitationId}",
			Summary:     "Cancel an invitation",
			Description: "The inviter, or an owner or admin of the workspace, recalls a pending invitation.",
			Auth:        "bearer",
			PathParams:  []documentation.Field{{Name: "invitationId", Type: "int64", Description: "The invitation's id"}},
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "The caller is not the inviter nor an owner or admin of the workspace"},
			},
		},
	}
}

// InvitationsDocumentation returns the invitation routes for the reference page.
func InvitationsDocumentation() documentation.Module {
	return documentation.Module{
		Name:        "invitation",
		Description: "Invitations are how a workspace gains a member: addressed to an email, created by an admin, and acted on by the invited address alone.",
		Routes:      append(workspaceInviteRoutes(), userInviteRoutes()...),
	}
}