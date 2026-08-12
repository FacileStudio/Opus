package workspace

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/FacileStudio/Opus/apps/api/internal/storage"
	"github.com/FacileStudio/tronc/errors"
)

// maxLogoBytes is the legacy upload ceiling: 2MB.
const maxLogoBytes = 2 * 1024 * 1024

// Controller serves the workspace routes.
type Controller struct {
	service *Service
}

func newController(service *Service) *Controller {
	return &Controller{service: service}
}

func (controller *Controller) list(ctx context.Context, userID int64) ([]WorkspaceResponse, error) {
	return controller.service.ListForUser(ctx, userID)
}

func (controller *Controller) members(ctx context.Context, userID, workspaceID int64) ([]MemberResponse, error) {
	if err := controller.service.RequireMember(ctx, userID, workspaceID); err != nil {
		return nil, err
	}
	return controller.service.Members(ctx, workspaceID)
}

func (controller *Controller) uploadLogo(ctx context.Context, userID, workspaceID int64, file multipart.File, header *multipart.FileHeader) (*UploadLogoResponse, error) {
	if err := controller.service.RequireMember(ctx, userID, workspaceID); err != nil {
		return nil, err
	}
	if header == nil {
		return nil, errors.Invalid("No logo file provided. Send as multipart with field name 'logo'")
	}
	if header.Size > maxLogoBytes {
		return nil, errors.TooLarge("File too large. Maximum 2MB")
	}
	ext, allowed := storage.ExtensionFromMime(header.Header.Get("Content-Type"))
	if !allowed {
		return nil, errors.Invalid("Invalid image type. Allowed: png, jpeg, webp, svg, gif")
	}
	data, err := io.ReadAll(io.LimitReader(file, maxLogoBytes))
	if err != nil {
		return nil, errors.Internal("failed to read the logo upload", err)
	}
	if len(data) == 0 {
		return nil, errors.Invalid("No logo file provided. Send as multipart with field name 'logo'")
	}
	url, err := controller.service.SetLogo(ctx, workspaceID, data, ext)
	if err != nil {
		return nil, err
	}
	return &UploadLogoResponse{URL: url, Message: "Logo uploaded successfully"}, nil
}

func (controller *Controller) deleteLogo(ctx context.Context, userID, workspaceID int64) (*DeleteLogoResponse, error) {
	if err := controller.service.RequireMember(ctx, userID, workspaceID); err != nil {
		return nil, err
	}
	if err := controller.service.ClearLogo(ctx, workspaceID); err != nil {
		return nil, err
	}
	return &DeleteLogoResponse{Message: "Logo deleted"}, nil
}

// serveLogo streams the stored logo bytes, or 404 when the workspace has none.
func (controller *Controller) serveLogo(ctx context.Context, workspaceID int64, w http.ResponseWriter) error {
	data, ext, ok, err := controller.service.Logo(ctx, workspaceID)
	if err != nil {
		return err
	}
	if !ok {
		return errors.NotFound("No logo found")
	}
	contentType := map[string]string{
		"png": "image/png",
		"jpg": "image/jpeg",
		"webp": "image/webp",
		"svg": "image/svg+xml",
		"gif": "image/gif",
	}[ext]
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
	return nil
}
