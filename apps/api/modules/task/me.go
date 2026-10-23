// Package task implements the personal task listing for the current user.
package task

import (
	"context"
	"net/http"
	"strconv"

	"github.com/FacileStudio/Opus/apps/api/internal/authcontext"
	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/go-chi/chi/v5"
)

// MyTaskResponse is a task assigned to the caller with its project display fields.
type MyTaskResponse struct {
	TaskResponse
	ProjectSlug string `json:"projectSlug"`
	ProjectName string `json:"projectName"`
	ProjectIcon string `json:"projectIcon"`
}

func registerTaskMeRoutes(router chi.Router, service *Service) {
	router.Get("/me", func(w http.ResponseWriter, request *http.Request) {
		workspaceID, err := strconv.ParseInt(request.URL.Query().Get("workspaceId"), 10, 64)
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid workspace id"))
			return
		}
		includeDone := request.URL.Query().Get("includeDone") == "true"
		identity, _ := authcontext.IdentityFromContext(request.Context())
		resp, err := service.controller.listMine(request.Context(), identity.UserID, workspaceID, includeDone)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

func (controller *Controller) listMine(ctx context.Context, userID, workspaceID int64, includeDone bool) ([]MyTaskResponse, error) {
	records, projects, err := controller.service.ListMine(ctx, userID, workspaceID, includeDone)
	if err != nil {
		return nil, err
	}
	out := make([]MyTaskResponse, len(records))
	for i, rec := range records {
		project := projects[rec.Task.ProjectID]
		out[i] = MyTaskResponse{
			TaskResponse: recordResponse(rec),
			ProjectSlug:  project.Slug,
			ProjectName:  project.Name,
			ProjectIcon:  project.Icon,
		}
	}
	return out, nil
}

// ListMine returns the tasks assigned to the user in a workspace, with their projects by ID.
func (service *Service) ListMine(ctx context.Context, userID, workspaceID int64, includeDone bool) ([]TaskRecord, map[int64]schemas.Project, error) {
	if err := service.verifyWorkspaceMembership(ctx, userID, workspaceID); err != nil {
		return nil, nil, err
	}
	query := service.db.WithContext(ctx).
		Where("workspace_id = ? AND actor_id = ? AND status <> ?", workspaceID, userID, "archived").
		Where("NOT EXISTS (SELECT 1 FROM projects p WHERE p.id = tasks.project_id AND p.archived_at IS NOT NULL)")
	if !includeDone {
		query = query.Where(
			"NOT EXISTS (SELECT 1 FROM columns c WHERE c.project_id = tasks.project_id AND c.slug = tasks.status AND c.is_final)",
		)
	}
	var tasks []schemas.Task
	if err := query.Order("project_id asc, number asc").Find(&tasks).Error; err != nil {
		return nil, nil, errors.Internal("failed to fetch tasks", err)
	}
	records, err := service.loadTaskRecords(ctx, tasks)
	if err != nil {
		return nil, nil, err
	}
	var projects []schemas.Project
	if err := service.db.WithContext(ctx).Where("workspace_id = ?", workspaceID).Find(&projects).Error; err != nil {
		return nil, nil, errors.Internal("failed to fetch projects", err)
	}
	byID := make(map[int64]schemas.Project, len(projects))
	for _, project := range projects {
		byID[project.ID] = project
	}
	return records, byID, nil
}
