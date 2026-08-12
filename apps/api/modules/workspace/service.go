package workspace

import (
	"context"
	"strconv"
	"time"

	"github.com/FacileStudio/Opus/apps/api/internal/storage"
	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

// Service owns the workspace reads and the logo file operations.
type Service struct {
	orm        *gorm.DB
	uploadDir  string
	controller *Controller
}

// NewService wires the workspace service with its database handle and the
// upload directory logos are filed under.
func NewService(orm *gorm.DB, uploadDir string) *Service {
	service := &Service{orm: orm, uploadDir: uploadDir}
	service.controller = newController(service)
	return service
}

// ListForUser returns every workspace the user is a member of, ordered by
// name for a stable list. Response fields match the legacy select exactly.
func (service *Service) ListForUser(ctx context.Context, userID int64) ([]WorkspaceResponse, error) {
	rows := []struct {
		ID          int64
		Name        string
		Slug        string
		Logo        string
		Description string
		CreatedAt   time.Time
	}{}
	err := service.orm.WithContext(ctx).
		Table("workspaces").
		Select("workspaces.id, workspaces.name, workspaces.slug, workspaces.logo, workspaces.description, workspaces.created_at").
		Joins("JOIN workspace_members ON workspace_members.workspace_id = workspaces.id").
		Where("workspace_members.user_id = ?", userID).
		Order("workspaces.name asc").
		Scan(&rows).Error
	if err != nil {
		return nil, errors.Internal("failed to list workspaces", err)
	}

	out := make([]WorkspaceResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, WorkspaceResponse{
			ID:          row.ID,
			Name:        row.Name,
			Slug:        row.Slug,
			Logo:        row.Logo,
			Description: row.Description,
			CreatedAt:   row.CreatedAt,
		})
	}
	return out, nil
}

// Members returns every member of a workspace, joined to the user rows the
// same way the legacy controller did.
func (service *Service) Members(ctx context.Context, workspaceID int64) ([]MemberResponse, error) {
	rows := []MemberResponse{}
	err := service.orm.WithContext(ctx).
		Table("users").
		Select("users.id, users.name, users.email, users.image, workspace_members.role").
		Joins("JOIN workspace_members ON workspace_members.user_id = users.id").
		Where("workspace_members.workspace_id = ?", workspaceID).
		Order("users.name asc").
		Scan(&rows).Error
	if err != nil {
		return nil, errors.Internal("failed to list workspace members", err)
	}
	return rows, nil
}

// RequireMember verifies the caller belongs to the workspace, answering the
// legacy middleware's 403. It is the one access check the workspace module
// needs until Phase 2 brings project-scoped checks.
func (service *Service) RequireMember(ctx context.Context, userID, workspaceID int64) error {
	var count int64
	err := service.orm.WithContext(ctx).
		Model(&schemas.WorkspaceMember{}).
		Where("user_id = ? AND workspace_id = ?", userID, workspaceID).
		Count(&count).Error
	if err != nil {
		return errors.Internal("failed to check workspace access", err)
	}
	if count == 0 {
		return errors.Forbidden("You don't have access to this workspace")
	}
	return nil
}

// SetLogo stores the uploaded bytes under the workspace's directory and stamps
// the logo column with a cache-busting value the frontend puts in the URL.
func (service *Service) SetLogo(ctx context.Context, workspaceID int64, data []byte, ext string) (string, error) {
	if err := storage.SaveWorkspaceLogo(service.uploadDir, workspaceID, data, ext); err != nil {
		return "", errors.Internal("failed to store the logo", err)
	}
	stamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	if err := service.orm.WithContext(ctx).Model(&schemas.Workspace{}).
		Where("id = ?", workspaceID).Update("logo", stamp).Error; err != nil {
		return "", errors.Internal("failed to record the logo", err)
	}
	return "/api/workspace/" + strconv.FormatInt(workspaceID, 10) + "/logo?v=" + stamp, nil
}

// ClearLogo deletes the stored files and empties the logo column.
func (service *Service) ClearLogo(ctx context.Context, workspaceID int64) error {
	if err := storage.DeleteWorkspaceLogo(service.uploadDir, workspaceID); err != nil {
		return errors.Internal("failed to delete the logo", err)
	}
	if err := service.orm.WithContext(ctx).Model(&schemas.Workspace{}).
		Where("id = ?", workspaceID).Update("logo", "").Error; err != nil {
		return errors.Internal("failed to clear the logo", err)
	}
	return nil
}

// Logo loads the stored logo bytes for serving. ok is false when the
// workspace has none.
func (service *Service) Logo(ctx context.Context, workspaceID int64) (data []byte, ext string, ok bool, err error) {
	data, ext, ok, readErr := storage.ReadWorkspaceLogo(service.uploadDir, workspaceID)
	if readErr != nil {
		return nil, "", false, errors.Internal("failed to read the logo", readErr)
	}
	return data, ext, ok, nil
}
