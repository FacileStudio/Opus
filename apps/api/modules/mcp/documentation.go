package mcp

import (
	documentation "github.com/FacileStudio/Opus/apps/api/internal/documentation"
	mcpkit "github.com/FacileStudio/tronc/mcp"
)

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
	return mcpkit.PendingRoutes[PendingItem](mcpkit.PendingDoc{
		Auth: "session cookie or bearer token",
		List: "Open actions, newest first, with their task or project as it is now. state is ready, " +
			"changed (edited after the request, so it cannot be confirmed) or missing.",
		Confirm: "Deletes the task or the project, once. A target edited since the request is refused with " +
			"412 and the action closes.",
		Dismiss: "Closes the action without running it. The task or project stays.",
		Changed: "The target changed after the request.",
	})
}
