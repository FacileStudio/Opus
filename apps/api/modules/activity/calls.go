package activity

import "context"

// Calls runs the activity routes for in-process callers such as the MCP
// server: the same checks as the HTTP routes, without HTTP.
type Calls struct {
	controller *Controller
}

// Calls returns the in-process surface of the activity routes.
func (service *Service) Calls() Calls {
	return Calls{controller: service.controller}
}

// List is GET /activity/{taskId}: the task's comments and events, newest
// first.
func (c Calls) List(ctx context.Context, userID, taskID int64) ([]ActivityResponse, error) {
	return c.controller.list(ctx, userID, taskID)
}

// Comment is POST /activity/comment.
func (c Calls) Comment(ctx context.Context, userID int64, req *CommentActivityRequest) (*ActivityResponse, error) {
	return c.controller.createComment(ctx, userID, req)
}
