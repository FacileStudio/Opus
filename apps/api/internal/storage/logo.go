// Package storage files uploaded assets on local disk. The s3 parenthesised
// storage option remains a Phase 6 concern; today logos live under UPLOAD_DIR.
package storage

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ExtensionFromMime maps an upload's content type to the file extension stored
// on disk. The second result reports whether the type is allowed at all.
func ExtensionFromMime(mime string) (string, bool) {
	mime = strings.ToLower(strings.TrimSpace(mime))
	switch mime {
	case "image/png":
		return "png", true
	case "image/jpg", "image/jpeg":
		return "jpg", true
	case "image/webp":
		return "webp", true
	case "image/svg+xml":
		return "svg", true
	case "image/gif":
		return "gif", true
	default:
		return "", false
	}
}

func allowedLogoExtensions() []string {
	return []string{"png", "jpg", "webp", "svg", "gif"}
}

func workspaceLogoDir(uploadDir string, workspaceID int64) string {
	return filepath.Join(uploadDir, "workspaces", strconv.FormatInt(workspaceID, 10))
}

// SaveWorkspaceLogo writes the logo under the workspace's directory, removing
// any previously stored extension so only one file carries the image.
func SaveWorkspaceLogo(uploadDir string, workspaceID int64, data []byte, ext string) error {
	dir := workspaceLogoDir(uploadDir, workspaceID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for _, other := range allowedLogoExtensions() {
		if other == ext {
			continue
		}
		if err := os.Remove(filepath.Join(dir, "logo."+other)); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return os.WriteFile(filepath.Join(dir, "logo."+ext), data, 0o644)
}

// DeleteWorkspaceLogo removes every stored logo extension for a workspace.
func DeleteWorkspaceLogo(uploadDir string, workspaceID int64) error {
	dir := workspaceLogoDir(uploadDir, workspaceID)
	for _, ext := range allowedLogoExtensions() {
		if err := os.Remove(filepath.Join(dir, "logo."+ext)); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

// LogoFile holds the stored logo bytes and format extension.
type LogoFile struct {
	Data      []byte
	Extension string
}

// ReadWorkspaceLogo returns the stored logo bytes and their extension.
func ReadWorkspaceLogo(uploadDir string, workspaceID int64) (*LogoFile, error) {
	dir := workspaceLogoDir(uploadDir, workspaceID)
	for _, candidate := range allowedLogoExtensions() {
		path := filepath.Join(dir, "logo."+candidate)
		raw, readErr := os.ReadFile(path)
		if readErr != nil {
			if os.IsNotExist(readErr) {
				continue
			}
			return nil, readErr
		}
		return &LogoFile{Data: raw, Extension: candidate}, nil
	}
	return nil, nil
}
