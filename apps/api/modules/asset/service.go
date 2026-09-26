package asset

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	stderrors "errors"
	"fmt"
	"io"
	"time"

	"github.com/FacileStudio/Opus/apps/api/internal/storage"
	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
	"gorm.io/gorm"
)

// Service provides asset persistence and object storage operations.
type Service struct {
	db         *gorm.DB
	storage    *storage.AssetStorage
	controller *controller
}

// NewService constructs a new asset domain service.
func NewService(db *gorm.DB, assetStorage *storage.AssetStorage) *Service {
	s := &Service{
		db:      db,
		storage: assetStorage,
	}
	s.controller = newController(s)
	return s
}

func generateAssetID() string {
	bytes := make([]byte, 12)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("ast_%d", time.Now().UnixNano())
	}
	return "ast_" + hex.EncodeToString(bytes)
}

func (s *Service) verifyTaskAndMember(ctx context.Context, userID, taskID int64) (*schemas.Task, *schemas.Project, error) {
	var task schemas.Task
	if err := s.db.WithContext(ctx).First(&task, taskID).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errors.NotFound("task not found")
		}
		return nil, nil, errors.Internal("failed to fetch task", err)
	}
	var project schemas.Project
	if err := s.db.WithContext(ctx).First(&project, task.ProjectID).Error; err != nil {
		return nil, nil, errors.NotFound("project not found")
	}
	var member schemas.WorkspaceMember
	err := s.db.WithContext(ctx).
		Where("workspace_id = ? AND user_id = ?", project.WorkspaceID, userID).
		First(&member).Error
	if err != nil {
		return nil, nil, errors.Forbidden("insufficient permissions in workspace")
	}
	return &task, &project, nil
}

// CreateImageUpload handles upload initiation and generates destination details.
func (s *Service) CreateImageUpload(ctx context.Context, userID, taskID int64, req CreateImageUploadRequest) (*CreateImageUploadResponse, error) {
	task, project, err := s.verifyTaskAndMember(ctx, userID, taskID)
	if err != nil {
		return nil, err
	}
	maxBytes := s.storage.MaxUploadBytes()
	if valErr := storage.ValidateTaskAssetUploadInput(req.ContentType, req.Size, maxBytes); valErr != nil {
		return nil, valErr
	}
	key := storage.BuildObjectKey(project.WorkspaceID, project.ID, task.ID, req.Surface, req.Filename)
	uploadURL, headers, err := s.storage.GenerateUpload(key, req.ContentType, req.Size, "")
	if err != nil {
		return nil, err
	}
	return &CreateImageUploadResponse{
		Key:       key,
		UploadURL: uploadURL,
		Headers:   headers,
	}, nil
}

// FinalizeImageUpload records the asset in the database after upload completes.
func (s *Service) FinalizeImageUpload(ctx context.Context, userID, taskID int64, req FinalizeImageUploadRequest) (*FinalizeImageUploadResponse, error) {
	task, project, err := s.verifyTaskAndMember(ctx, userID, taskID)
	if err != nil {
		return nil, err
	}
	if !storage.AssertTaskImageKeyMatchesContext(req.Key, project.WorkspaceID, project.ID, task.ID, req.Surface) {
		return nil, errors.Invalid("Image upload key does not match the task context.")
	}
	kind := "attachment"
	if storage.IsImageContentType(req.ContentType) {
		kind = "image"
	}
	assetRecord := schemas.Asset{
		ID:          generateAssetID(),
		WorkspaceID: project.WorkspaceID,
		ProjectID:   project.ID,
		TaskID:      &task.ID,
		ObjectKey:   req.Key,
		Filename:    req.Filename,
		MimeType:    req.ContentType,
		Size:        req.Size,
		Kind:        kind,
		Surface:     req.Surface,
		CreatedBy:   &userID,
	}
	if err := s.db.WithContext(ctx).Create(&assetRecord).Error; err != nil {
		return nil, errors.Internal("failed to save asset record", err)
	}
	return &FinalizeImageUploadResponse{
		ID:  assetRecord.ID,
		URL: fmt.Sprintf("/api/asset/%s", assetRecord.ID),
	}, nil
}

func (s *Service) checkAssetAccess(ctx context.Context, callerID int64, asset *schemas.Asset, shareToken string) (bool, error) {
	var project schemas.Project
	if err := s.db.WithContext(ctx).First(&project, asset.ProjectID).Error; err != nil {
		return false, errors.NotFound("project not found")
	}
	if callerID > 0 {
		var member schemas.WorkspaceMember
		err := s.db.WithContext(ctx).
			Where("workspace_id = ? AND user_id = ?", asset.WorkspaceID, callerID).
			First(&member).Error
		if err == nil {
			return project.IsPublic, nil
		}
	}
	if project.IsPublic {
		return true, nil
	}
	if shareToken != "" {
		var link schemas.ShareLink
		now := time.Now()
		err := s.db.WithContext(ctx).
			Where("token = ? AND project_id = ? AND is_active = ?", shareToken, asset.ProjectID, true).
			Where("expires_at IS NULL OR expires_at > ?", now).
			First(&link).Error
		if err == nil {
			return project.IsPublic, nil
		}
	}
	return false, errors.Unauthorized("Unauthorized")
}

// GetAsset opens the asset object stream after verifying read access.
func (s *Service) GetAsset(ctx context.Context, callerID int64, assetID, shareToken string) (*AssetDownload, error) {
	var asset schemas.Asset
	if err := s.db.WithContext(ctx).First(&asset, "id = ?", assetID).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("Asset not found")
		}
		return nil, errors.Internal("failed to query asset", err)
	}
	isPublic, err := s.checkAssetAccess(ctx, callerID, &asset, shareToken)
	if err != nil {
		return nil, err
	}
	stream, err := s.storage.Open(ctx, asset.ObjectKey)
	if err != nil {
		return nil, err
	}
	return &AssetDownload{
		Body:          stream.Body,
		Filename:      asset.Filename,
		MimeType:      asset.MimeType,
		ContentLength: stream.ContentLength,
		IsPublic:      isPublic,
		LastModified:  stream.LastModified,
	}, nil
}

// StoreLocalAsset writes incoming raw data to local disk.
func (s *Service) StoreLocalAsset(key string, r io.Reader) error {
	return s.storage.StoreLocal(key, r)
}
