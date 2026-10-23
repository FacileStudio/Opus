package realtime

import (
	"context"
	"strconv"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"gorm.io/gorm"
)

// Authorizer reports whether a user may listen on a project or workspace channel.
type Authorizer func(ctx context.Context, userID int64, projectID, workspaceID string) bool

// NewMembershipAuthorizer allows a user onto a channel only when they belong to its workspace.
func NewMembershipAuthorizer(db *gorm.DB) Authorizer {
	return func(ctx context.Context, userID int64, projectID, workspaceID string) bool {
		workspace, err := strconv.ParseInt(workspaceID, 10, 64)
		if projectID != "" {
			project, parseErr := strconv.ParseInt(projectID, 10, 64)
			if parseErr != nil {
				return false
			}
			var found schemas.Project
			if db.WithContext(ctx).Select("workspace_id").First(&found, project).Error != nil {
				return false
			}
			workspace, err = found.WorkspaceID, nil
		}
		if err != nil {
			return false
		}
		var count int64
		db.WithContext(ctx).Model(&schemas.WorkspaceMember{}).
			Where("workspace_id = ? AND user_id = ?", workspace, userID).
			Count(&count)
		return count > 0
	}
}

// NewProjectWorkspaceResolver maps a project id to the id of its workspace, or "" when unknown.
func NewProjectWorkspaceResolver(db *gorm.DB) func(projectID string) string {
	return func(projectID string) string {
		project, err := strconv.ParseInt(projectID, 10, 64)
		if err != nil {
			return ""
		}
		var found schemas.Project
		if db.Select("workspace_id").First(&found, project).Error != nil {
			return ""
		}
		return strconv.FormatInt(found.WorkspaceID, 10)
	}
}
