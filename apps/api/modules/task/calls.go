package task

import "context"

// Calls runs the task routes for in-process callers such as the MCP server:
// the same checks, realtime broadcasts and Antenne events as the HTTP
// routes, without HTTP.
type Calls struct {
	controller *Controller
}

// Calls returns the in-process surface of the task routes.
func (service *Service) Calls() Calls {
	return Calls{controller: service.controller}
}

// Get is GET /task/{id}.
func (c Calls) Get(ctx context.Context, userID, taskID int64) (*TaskResponse, error) {
	return c.controller.get(ctx, userID, taskID)
}

// Create is POST /task/{projectId}.
func (c Calls) Create(ctx context.Context, userID, projectID int64, req *CreateTaskRequest) (*TaskResponse, error) {
	return c.controller.create(ctx, userID, projectID, req)
}

// Update is PUT /task/{id}: every field is replaced.
func (c Calls) Update(ctx context.Context, userID, taskID int64, req *UpdateTaskRequest) (*TaskResponse, error) {
	return c.controller.update(ctx, userID, taskID, req)
}

// Move is PUT /task/move/{id}.
func (c Calls) Move(ctx context.Context, userID, taskID int64, req *MoveTaskRequest) (*MoveTaskResponse, error) {
	return c.controller.move(ctx, userID, taskID, req)
}

// Delete is DELETE /task/{id}.
func (c Calls) Delete(ctx context.Context, userID, taskID int64) (*TaskResponse, error) {
	return c.controller.delete(ctx, userID, taskID)
}
