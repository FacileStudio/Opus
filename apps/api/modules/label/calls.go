package label

import "context"

// Calls runs the label routes for in-process callers such as the MCP server:
// the same checks and realtime broadcasts as the HTTP routes, without HTTP.
type Calls struct {
	controller *Controller
}

// Calls returns the in-process surface of the label routes.
func (service *Service) Calls() Calls {
	return Calls{controller: service.controller}
}

// Create is POST /label: with TaskID set, it attaches a label to that task.
func (c Calls) Create(ctx context.Context, userID int64, req *CreateLabelRequest) (*LabelResponse, error) {
	return c.controller.create(ctx, userID, req)
}

// ListByWorkspace is GET /label/workspace/{workspaceId}.
func (c Calls) ListByWorkspace(ctx context.Context, userID, workspaceID int64) ([]LabelResponse, error) {
	return c.controller.listByWorkspace(ctx, userID, workspaceID)
}

// Delete is DELETE /label/{id}, which removes a label from its task.
func (c Calls) Delete(ctx context.Context, userID, labelID int64) (*LabelResponse, error) {
	return c.controller.delete(ctx, userID, labelID)
}
