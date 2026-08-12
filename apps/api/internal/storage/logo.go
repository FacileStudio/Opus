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

// logoExtensions is the set written by SaveWorkspaceLogo; the png family
// resolves below so uploads and lookups agree on the filenames.
var logoExtensions = []string{"png", "jpg", "webp", "svg", "gif"}

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
	for _, other := range logoExtensions {
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
	for _, ext := range logoExtensions {
		if err := os.Remove(filepath.Join(dir, "logo."+ext)); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

// ReadWorkspaceLogo returns the stored logo bytes and their extension. ok is
// false when the workspace has no logo on disk.
func ReadWorkspaceLogo(uploadDir string, workspaceID int64) (data []byte, ext string, ok bool, err error) {
	dir := workspaceLogoDir(uploadDir, workspaceID)
	for _, candidate := range logoExtensions {
		path := filepath.Join(dir, "logo."+candidate)
		raw, readErr := os.ReadFile(path)
		if readErr != nil {
			if os.IsNotExist(readErr) {
				continue
			}
			return nil, "", false, readErr
		}
		return raw, candidate, true, nil
	}
	return nil, "", false, nil
}