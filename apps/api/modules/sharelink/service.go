package sharelink

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	stderrors "errors"
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
	"gorm.io/gorm"
)

// Service manages share link persistence and workspace security.
type Service struct {
	db         *gorm.DB
	controller *Controller
}

// NewService constructs a new share link service instance.
func NewService(db *gorm.DB) *Service {
	service := &Service{db: db}
	service.controller = newController(service)
	return service
}

// RequireMember verifies caller belongs to the given workspace.
func (service *Service) RequireMember(ctx context.Context, userID, workspaceID int64) error {
	var count int64
	err := service.db.WithContext(ctx).
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

func (service *Service) verifyProjectAccess(ctx context.Context, userID, projectID int64) error {
	var project schemas.Project
	err := service.db.WithContext(ctx).
		Select("id, workspace_id").
		First(&project, "id = ?", projectID).Error
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.NotFound("Project not found")
		}
		return errors.Internal("failed to query project", err)
	}
	return service.RequireMember(ctx, userID, project.WorkspaceID)
}

func (service *Service) verifyShareLinkAccess(ctx context.Context, userID, shareLinkID int64) (*schemas.ShareLink, error) {
	var link schemas.ShareLink
	err := service.db.WithContext(ctx).First(&link, "id = ?", shareLinkID).Error
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("Share link not found")
		}
		return nil, errors.Internal("failed to query share link", err)
	}
	if err := service.verifyProjectAccess(ctx, userID, link.ProjectID); err != nil {
		return nil, err
	}
	return &link, nil
}

// ListByProject returns all share links associated with a project.
func (service *Service) ListByProject(ctx context.Context, userID, projectID int64) ([]schemas.ShareLink, error) {
	if err := service.verifyProjectAccess(ctx, userID, projectID); err != nil {
		return nil, err
	}
	var links []schemas.ShareLink
	err := service.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("created_at desc").
		Find(&links).Error
	if err != nil {
		return nil, errors.Internal("failed to list share links", err)
	}
	return links, nil
}

func generateToken() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Create inserts a new share link with a unique cryptographic token.
func (service *Service) Create(ctx context.Context, userID int64, req *CreateShareLinkRequest) (*schemas.ShareLink, error) {
	if err := service.verifyProjectAccess(ctx, userID, int64(req.ProjectID)); err != nil {
		return nil, err
	}
	token, err := generateToken()
	if err != nil {
		return nil, errors.Internal("failed to generate share token", err)
	}
	now := time.Now()
	label := ""
	if req.Label != nil {
		label = *req.Label
	}
	link := schemas.ShareLink{
		ProjectID: int64(req.ProjectID),
		Token:     token,
		Label:     label,
		ExpiresAt: req.ExpiresAt,
		IsActive:  true,
		CreatedBy: userID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := service.db.WithContext(ctx).Create(&link).Error; err != nil {
		return nil, errors.Internal("failed to create share link", err)
	}
	return &link, nil
}

// Update modifies label, expiry, or active state of a share link.
func (service *Service) Update(ctx context.Context, userID, shareLinkID int64, req *UpdateShareLinkRequest) (*schemas.ShareLink, error) {
	link, err := service.verifyShareLinkAccess(ctx, userID, shareLinkID)
	if err != nil {
		return nil, err
	}
	if req.Label != nil {
		link.Label = *req.Label
	}
	if req.ExpiresAt != nil {
		link.ExpiresAt = req.ExpiresAt
	}
	if req.IsActive != nil {
		link.IsActive = *req.IsActive
	}
	link.UpdatedAt = time.Now()
	if err := service.db.WithContext(ctx).Save(link).Error; err != nil {
		return nil, errors.Internal("failed to update share link", err)
	}
	return link, nil
}

// Delete removes a share link after verifying access.
func (service *Service) Delete(ctx context.Context, userID, shareLinkID int64) (*schemas.ShareLink, error) {
	link, err := service.verifyShareLinkAccess(ctx, userID, shareLinkID)
	if err != nil {
		return nil, err
	}
	if err := service.db.WithContext(ctx).Delete(link).Error; err != nil {
		return nil, errors.Internal("failed to delete share link", err)
	}
	return link, nil
}

// GetSharedProject looks up a project by token and records access telemetry.
func (service *Service) GetSharedProject(ctx context.Context, token string) (*SharedProjectResult, error) {
	var link schemas.ShareLink
	err := service.db.WithContext(ctx).Where("token = ?", token).First(&link).Error
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("Share link not found")
		}
		return nil, errors.Internal("failed to query share link", err)
	}
	if !link.IsActive {
		return nil, errors.NotFound("Share link is inactive")
	}
	if link.ExpiresAt != nil && link.ExpiresAt.Before(time.Now()) {
		return nil, errors.NotFound("Share link has expired")
	}
	now := time.Now()
	service.db.WithContext(ctx).Model(&link).Updates(map[string]any{
		"last_accessed_at": now,
		"access_count":     gorm.Expr("access_count + 1"),
	})
	return service.loadSharedProject(ctx, link.ProjectID)
}

func (service *Service) loadSharedProject(ctx context.Context, projectID int64) (*SharedProjectResult, error) {
	var project schemas.Project
	if err := service.db.WithContext(ctx).First(&project, "id = ?", projectID).Error; err != nil {
		return nil, errors.NotFound("Project not found")
	}
	var columns []schemas.Column
	service.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("position asc").
		Find(&columns)
	var tasks []schemas.Task
	service.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("position asc").
		Find(&tasks)
	return &SharedProjectResult{
		Project: project,
		Columns: columns,
		Tasks:   tasks,
	}, nil
}
