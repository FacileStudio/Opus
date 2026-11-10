package project

import (
	"context"
	"strconv"
)

// Calls runs the project routes for in-process callers such as the MCP
// server: the same checks and Antenne events as the HTTP routes, without
// HTTP.
type Calls struct {
	controller *controller
}

// Calls returns the in-process surface of the project routes.
func (s *Service) Calls() Calls {
	return Calls{controller: s.controller}
}

// List is GET /project?workspaceId=.
func (c Calls) List(ctx context.Context, userID, workspaceID int64, includeArchived bool) ([]ProjectResponse, error) {
	return c.controller.list(ctx, userID, strconv.FormatInt(workspaceID, 10), includeArchived)
}

// Get is GET /project/{id}.
func (c Calls) Get(ctx context.Context, userID, projectID int64) (*ProjectResponse, error) {
	return c.controller.get(ctx, userID, projectID)
}

// Create is POST /project.
func (c Calls) Create(ctx context.Context, userID int64, req *CreateProjectRequest) (*ProjectResponse, error) {
	return c.controller.create(ctx, userID, req)
}

// Update is PUT /project/{id}: only the fields set change.
func (c Calls) Update(ctx context.Context, userID, projectID int64, req *UpdateProjectRequest) (*ProjectResponse, error) {
	return c.controller.update(ctx, userID, projectID, req)
}

// Archive is PUT /project/{id}/archive.
func (c Calls) Archive(ctx context.Context, userID, projectID int64) (*ProjectResponse, error) {
	return c.controller.archive(ctx, userID, projectID)
}

// Unarchive is PUT /project/{id}/unarchive.
func (c Calls) Unarchive(ctx context.Context, userID, projectID int64) (*ProjectResponse, error) {
	return c.controller.unarchive(ctx, userID, projectID)
}

// Delete is DELETE /project/{id}.
func (c Calls) Delete(ctx context.Context, userID, projectID int64) (*ProjectResponse, error) {
	return c.controller.delete(ctx, userID, projectID)
}
