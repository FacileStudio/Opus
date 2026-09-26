package asset

import (
	"context"
	"io"
	"strings"

	"github.com/FacileStudio/tronc/errors"
)

type controller struct {
	service *Service
}

func newController(service *Service) *controller {
	return &controller{service: service}
}

func (c *controller) createImageUpload(ctx context.Context, userID, taskID int64, req CreateImageUploadRequest) (*CreateImageUploadResponse, error) {
	if strings.TrimSpace(req.Filename) == "" {
		return nil, errors.Invalid("filename is required")
	}
	if req.Surface != "description" && req.Surface != "comment" {
		req.Surface = "description"
	}
	return c.service.CreateImageUpload(ctx, userID, taskID, req)
}

func (c *controller) finalizeImageUpload(ctx context.Context, userID, taskID int64, req FinalizeImageUploadRequest) (*FinalizeImageUploadResponse, error) {
	if strings.TrimSpace(req.Key) == "" {
		return nil, errors.Invalid("object key is required")
	}
	if strings.TrimSpace(req.Filename) == "" {
		return nil, errors.Invalid("filename is required")
	}
	if req.Surface != "description" && req.Surface != "comment" {
		req.Surface = "description"
	}
	return c.service.FinalizeImageUpload(ctx, userID, taskID, req)
}

func (c *controller) getAsset(ctx context.Context, callerID int64, assetID, shareToken string) (*AssetDownload, error) {
	if strings.TrimSpace(assetID) == "" {
		return nil, errors.Invalid("asset id is required")
	}
	return c.service.GetAsset(ctx, callerID, assetID, shareToken)
}

func (c *controller) storeDirectUpload(key string, body io.Reader) error {
	cleanKey := strings.TrimSpace(key)
	if cleanKey == "" {
		return errors.Invalid("upload key is required")
	}
	return c.service.StoreLocalAsset(cleanKey, body)
}
