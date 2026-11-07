package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/FacileStudio/tronc/errors"
)

var allowedImageMimes = map[string]bool{
	"image/apng":    true,
	"image/avif":    true,
	"image/gif":     true,
	"image/heic":    true,
	"image/heif":    true,
	"image/jpeg":    true,
	"image/jpg":     true,
	"image/png":     true,
	"image/svg+xml": true,
	"image/webp":    true,
}

var (
	nonAlphanumeric = regexp.MustCompile(`[^a-z0-9._-]+`)
	multipleHyphens = regexp.MustCompile(`-{2,}`)
)

// ObjectStream holds the readable body and metadata of a stored object.
type ObjectStream struct {
	Body          io.ReadCloser
	ContentType   string
	ContentLength int64
	ETag          string
	LastModified  time.Time
}

// IsImageContentType reports whether the content type is a supported image MIME type.
func IsImageContentType(contentType string) bool {
	return allowedImageMimes[strings.ToLower(strings.TrimSpace(contentType))]
}

// SanitizePathSegment normalizes a string into a safe filesystem and URL path segment.
func SanitizePathSegment(value string) string {
	cleaned := strings.ToLower(value)
	cleaned = nonAlphanumeric.ReplaceAllString(cleaned, "-")
	cleaned = multipleHyphens.ReplaceAllString(cleaned, "-")
	cleaned = strings.Trim(cleaned, "-")
	if cleaned == "" {
		return "file"
	}
	return cleaned
}

// GetFileExtension extracts and normalizes the extension from a filename.
func GetFileExtension(filename string) string {
	trimmed := strings.TrimSpace(filename)
	ext := filepath.Ext(trimmed)
	if ext == "" {
		return ""
	}
	cleaned := SanitizePathSegment(strings.TrimPrefix(ext, "."))
	if len(cleaned) > 12 {
		return cleaned[:12]
	}
	return cleaned
}

// BuildObjectKeyPrefix formats the directory prefix for a task attachment.
func BuildObjectKeyPrefix(workspaceID, projectID, taskID int64, surface string) string {
	surfaceFolder := "descriptions"
	if surface == "comment" {
		surfaceFolder = "comments"
	}
	return fmt.Sprintf("workspace/%d/project/%d/task/%d/%s", workspaceID, projectID, taskID, surfaceFolder)
}

func generateRandomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "00000000"
	}
	return hex.EncodeToString(b)
}

// BuildObjectKey generates a collision-resistant object key for task uploads.
func BuildObjectKey(workspaceID, projectID, taskID int64, surface, filename string) string {
	prefix := BuildObjectKeyPrefix(workspaceID, projectID, taskID, surface)
	ext := GetFileExtension(filename)
	base := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	cleanBase := SanitizePathSegment(base)
	if len(cleanBase) > 64 {
		cleanBase = cleanBase[:64]
	}
	timestamp := time.Now().UnixMilli()
	randomID := generateRandomHex(6)
	if ext != "" {
		return fmt.Sprintf("%s/%s-%d-%s.%s", prefix, cleanBase, timestamp, randomID, ext)
	}
	return fmt.Sprintf("%s/%s-%d-%s", prefix, cleanBase, timestamp, randomID)
}

// AssertTaskImageKeyMatchesContext verifies the key belongs to the task and surface.
func AssertTaskImageKeyMatchesContext(key string, workspaceID, projectID, taskID int64, surface string) bool {
	prefix := BuildObjectKeyPrefix(workspaceID, projectID, taskID, surface) + "/"
	return strings.HasPrefix(key, prefix)
}

// ValidateTaskAssetUploadInput checks content type and size constraints.
func ValidateTaskAssetUploadInput(contentType string, size, maxBytes int64) error {
	if strings.TrimSpace(contentType) == "" {
		return errors.Invalid("A valid content type is required.")
	}
	if size <= 0 {
		return errors.Invalid("Upload size must be greater than zero.")
	}
	if size > maxBytes {
		mb := maxBytes / (1024 * 1024)
		return errors.Invalid(fmt.Sprintf("Upload exceeds the maximum upload size of %dMB.", mb))
	}
	return nil
}

// BuildContentDisposition formats a safe Content-Disposition header.
func BuildContentDisposition(filename string) string {
	cleaned := strings.TrimSpace(filename)
	if cleaned == "" {
		cleaned = "file"
	}
	safeAscii := nonAlphanumeric.ReplaceAllString(cleaned, "_")
	encoded := url.PathEscape(cleaned)
	return fmt.Sprintf(`inline; filename="%s"; filename*=UTF-8''%s`, safeAscii, encoded)
}

func sanitizeStorageKey(key string) (string, error) {
	clean := filepath.Clean(filepath.FromSlash(key))
	if strings.HasPrefix(clean, "..") || filepath.IsAbs(clean) {
		return "", errors.Invalid("invalid object key path")
	}
	return clean, nil
}

// SaveLocalAsset writes asset bytes into the local uploads directory.
func SaveLocalAsset(uploadDir, key string, r io.Reader) error {
	cleanKey, err := sanitizeStorageKey(key)
	if err != nil {
		return err
	}
	targetPath := filepath.Join(uploadDir, "assets", cleanKey)
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return errors.Internal("failed to create asset directory", err)
	}
	dest, err := os.Create(targetPath)
	if err != nil {
		return errors.Internal("failed to create asset file", err)
	}
	defer dest.Close()
	if _, err := io.Copy(dest, r); err != nil {
		return errors.Internal("failed to write asset bytes", err)
	}
	return nil
}

// ReadLocalAsset opens a local asset file stream from disk.
func ReadLocalAsset(uploadDir, key string) (*ObjectStream, error) {
	cleanKey, err := sanitizeStorageKey(key)
	if err != nil {
		return nil, err
	}
	targetPath := filepath.Join(uploadDir, "assets", cleanKey)
	file, err := os.Open(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.NotFound("Asset object not found")
		}
		return nil, errors.Internal("failed to open asset file", err)
	}
	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, errors.Internal("failed to stat asset file", err)
	}
	return &ObjectStream{
		Body:          file,
		ContentLength: stat.Size(),
		LastModified:  stat.ModTime(),
	}, nil
}

// AssetStorage coordinates file storage operations across local disk and S3.
type AssetStorage struct {
	uploadDir string
	s3Config  *S3Config
}

// NewAssetStorage constructs a new asset storage manager.
func NewAssetStorage(uploadDir string, s3Config *S3Config) *AssetStorage {
	return &AssetStorage{uploadDir: uploadDir, s3Config: s3Config}
}

// IsS3 returns true when S3 object storage is configured.
func (s *AssetStorage) IsS3() bool {
	return s.s3Config != nil
}

// MaxUploadBytes returns the maximum permitted upload payload size.
func (s *AssetStorage) MaxUploadBytes() int64 {
	if s.s3Config != nil && s.s3Config.MaxImageUploadBytes > 0 {
		return s.s3Config.MaxImageUploadBytes
	}
	return 10 * 1024 * 1024
}

// GenerateUpload generates an upload URL and required headers.
func (s *AssetStorage) GenerateUpload(key, contentType string, size int64, localBaseURL string) (string, map[string]string, error) {
	headers := map[string]string{"Content-Type": contentType}
	if s.IsS3() {
		uploadURL, err := PresignURL(s.s3Config, http.MethodPut, key, contentType, s.s3Config.PresignTTLSeconds)
		if err != nil {
			return "", nil, errors.Internal("failed to presign upload url", err)
		}
		return uploadURL, headers, nil
	}
	cleanKey := strings.TrimPrefix(key, "/")
	uploadURL := fmt.Sprintf("%s/api/asset/upload/%s", strings.TrimSuffix(localBaseURL, "/"), cleanKey)
	return uploadURL, headers, nil
}

// StoreLocal saves the direct upload stream to disk when using local storage.
func (s *AssetStorage) StoreLocal(key string, r io.Reader) error {
	return SaveLocalAsset(s.uploadDir, key, r)
}

// Open retrieves an object stream from either S3 or local disk.
func (s *AssetStorage) Open(ctx context.Context, key string) (*ObjectStream, error) {
	if s.IsS3() {
		return FetchS3Object(ctx, s.s3Config, key)
	}
	return ReadLocalAsset(s.uploadDir, key)
}

const deleteBatchSize = 16

// Delete removes the objects behind keys from S3 or local disk, a batch at a
// time. A missing object is not an error, so the call can be repeated. Every
// key is attempted and the first failure is returned.
func (s *AssetStorage) Delete(ctx context.Context, keys []string) error {
	var first error
	for start := 0; start < len(keys); start += deleteBatchSize {
		end := min(start+deleteBatchSize, len(keys))
		errs := make([]error, end-start)
		var wg sync.WaitGroup
		for i, key := range keys[start:end] {
			wg.Add(1)
			go func() {
				defer wg.Done()
				errs[i] = s.deleteOne(ctx, key)
			}()
		}
		wg.Wait()
		for _, err := range errs {
			if err != nil && first == nil {
				first = err
			}
		}
	}
	return first
}

func (s *AssetStorage) deleteOne(ctx context.Context, key string) error {
	if s.IsS3() {
		return DeleteS3Object(ctx, s.s3Config, key)
	}
	cleanKey, err := sanitizeStorageKey(key)
	if err != nil {
		return err
	}
	if err := os.Remove(filepath.Join(s.uploadDir, "assets", cleanKey)); err != nil && !os.IsNotExist(err) {
		return errors.Internal("failed to remove asset file", err)
	}
	return nil
}
