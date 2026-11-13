package storage

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	stderrors "errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/FacileStudio/tronc/errors"
)

// UploadURLLifetime is how long a signed local upload URL stays valid.
const UploadURLLifetime = 15 * time.Minute

func randomUploadKey() []byte {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		panic("storage: no randomness for the upload signing key")
	}
	return key
}

// SetUploadKey replaces the key that signs local upload URLs, so a URL stays
// valid across restarts and replicas. An empty key is ignored.
func (s *AssetStorage) SetUploadKey(key []byte) {
	if len(key) > 0 {
		s.uploadKey = key
	}
}

func (s *AssetStorage) uploadSignature(key string, expires, userID, maxBytes int64) string {
	mac := hmac.New(sha256.New, s.uploadKey)
	mac.Write([]byte(strings.Join([]string{"opus-upload", key,
		strconv.FormatInt(expires, 10), strconv.FormatInt(userID, 10), strconv.FormatInt(maxBytes, 10)}, "\x00")))
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *AssetStorage) signUpload(key string, userID int64, now time.Time) url.Values {
	expires := now.Add(UploadURLLifetime).Unix()
	maxBytes := s.MaxUploadBytes()
	return url.Values{
		"exp": {strconv.FormatInt(expires, 10)},
		"uid": {strconv.FormatInt(userID, 10)},
		"max": {strconv.FormatInt(maxBytes, 10)},
		"sig": {s.uploadSignature(key, expires, userID, maxBytes)},
	}
}

// VerifyUpload checks a local upload URL's signature for key and returns the
// size cap it grants. No signature is 401; a wrong or expired one is 403.
func (s *AssetStorage) VerifyUpload(key string, query url.Values, now time.Time) (int64, error) {
	sig := query.Get("sig")
	if sig == "" {
		return 0, errors.Unauthorized("upload URL is not signed")
	}
	expires, errExp := strconv.ParseInt(query.Get("exp"), 10, 64)
	userID, errUID := strconv.ParseInt(query.Get("uid"), 10, 64)
	maxBytes, errMax := strconv.ParseInt(query.Get("max"), 10, 64)
	if errExp != nil || errUID != nil || errMax != nil || maxBytes <= 0 {
		return 0, errors.Forbidden("upload URL is malformed")
	}
	want := s.uploadSignature(key, expires, userID, maxBytes)
	if !hmac.Equal([]byte(sig), []byte(want)) {
		return 0, errors.Forbidden("upload URL signature does not match")
	}
	if now.Unix() > expires {
		return 0, errors.Forbidden("upload URL has expired")
	}
	return maxBytes, nil
}

// StoreLocal saves the direct upload stream to disk when using local storage.
// It never replaces an existing object.
func (s *AssetStorage) StoreLocal(key string, r io.Reader) error {
	return SaveLocalAsset(s.uploadDir, key, r)
}

// SaveLocalAsset writes asset bytes into the local uploads directory. It
// refuses to overwrite an existing file, and a body cut by
// http.MaxBytesReader answers 413 and leaves no partial file.
func SaveLocalAsset(uploadDir, key string, r io.Reader) error {
	cleanKey, err := sanitizeStorageKey(key)
	if err != nil {
		return err
	}
	targetPath := filepath.Join(uploadDir, "assets", cleanKey)
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return errors.Internal("failed to create asset directory", err)
	}
	dest, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if os.IsExist(err) {
			return errors.Conflict("asset object already exists")
		}
		return errors.Internal("failed to create asset file", err)
	}
	_, copyErr := io.Copy(dest, r)
	closeErr := dest.Close()
	if copyErr == nil && closeErr == nil {
		return nil
	}
	os.Remove(targetPath)
	var tooLarge *http.MaxBytesError
	if stderrors.As(copyErr, &tooLarge) {
		return errors.TooLarge("upload exceeds the signed size limit")
	}
	return errors.Internal("failed to write asset bytes", stderrors.Join(copyErr, closeErr))
}
