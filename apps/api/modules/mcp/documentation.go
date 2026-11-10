package mcp

import documentation "github.com/FacileStudio/Opus/apps/api/internal/documentation"

// Documentation describes the confirm page routes for the API reference. The
// MCP endpoint itself is outside /api: POST /mcp speaks the Model Context
// Protocol to any MCP client holding a Registre access token for this
// server, and /.well-known/oauth-protected-resource/mcp describes how to get
// one.
func Documentation() documentation.Module {
	return documentation.Module{
		Name: "mcp",
		Description: "Sensitive actions MCP clients asked for, waiting for the person to confirm them: deleting a " +
			"task or a project. A client that cannot ask the person itself hands them a link to the pending page, " +
			"which uses these routes. The MCP endpoint is POST /mcp, outside /api.",
		Routes: pendingRoutes(),
	}
}

// pendingRoutes documents the three confirm page routes.
func pendingRoutes() []documentation.Route {
	unauthorized := documentation.Error{Status: 401, Code: "unauthenticated", Description: "The request carries no valid session."}
	delegated := documentation.Error{Status: 403, Code: "permission_denied", Description: "A delegated token cannot confirm actions."}
	gone := documentation.Error{Status: 404, Code: "not_found", Description: "The action ran, was dismissed or expired."}
	return []documentation.Route{
		{
			Method: "GET", Path: "/mcp/pending", Summary: "List actions waiting for confirmation",
			Description: "Open actions, newest first, with their task or project as it is now. state is ready, " +
				"changed (edited after the request, so it cannot be confirmed) or missing.",
			Auth: "session cookie or bearer token", ResponseBody: PendingList{},
			Errors: []documentation.Error{unauthorized, delegated},
		},
		{
			Method: "POST", Path: "/mcp/pending/{id}/confirm", Summary: "Confirm an action",
			Description: "Deletes the task or the project, once. A target edited since the request is refused with " +
				"412 and the action closes.",
			Auth: "session cookie or bearer token", RequestBody: EmptyRequest{}, ResponseBody: PendingResult{},
			Errors: []documentation.Error{unauthorized, delegated, gone,
				{Status: 412, Code: "failed_precondition", Description: "The target changed after the request."}},
		},
		{
			Method: "POST", Path: "/mcp/pending/{id}/dismiss", Summary: "Dismiss an action",
			Description: "Closes the action without running it. The task or project stays.",
			Auth:        "session cookie or bearer token", RequestBody: EmptyRequest{}, ResponseBody: PendingResult{},
			Errors: []documentation.Error{unauthorized, delegated, gone},
		},
	}
}
