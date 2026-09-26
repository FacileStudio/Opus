package asset

import (
	"io"
	"time"
)

// CreateImageUploadRequest holds parameters for initiating an upload.
type CreateImageUploadRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	Size        int64  `json:"size"`
	Surface     string `json:"surface"`
}

// CreateImageUploadResponse contains the upload destination and headers.
type CreateImageUploadResponse struct {
	Key       string            `json:"key"`
	UploadURL string            `json:"uploadUrl"`
	Headers   map[string]string `json:"headers"`
}

// FinalizeImageUploadRequest contains metadata to record an uploaded asset.
type FinalizeImageUploadRequest struct {
	Key         string `json:"key"`
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	Size        int64  `json:"size"`
	Surface     string `json:"surface"`
}

// FinalizeImageUploadResponse holds the created asset identifier and access URL.
type FinalizeImageUploadResponse struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

// AssetDownload represents an opened asset stream and its HTTP headers.
type AssetDownload struct {
	Body          io.ReadCloser
	Filename      string
	MimeType      string
	ContentLength int64
	IsPublic      bool
	LastModified  time.Time
}
