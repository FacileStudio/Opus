package asset

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/FacileStudio/Opus/apps/api/internal/storage"
	"github.com/go-chi/chi/v5"
)

const uploadTestKey = "workspace/1/project/2/task/3/descriptions/shot-1-abcdef.png"

func uploadRouter(t *testing.T) (http.Handler, *storage.AssetStorage, string) {
	t.Helper()
	dir := t.TempDir()
	store := storage.NewAssetStorage(dir, nil)
	r := chi.NewRouter()
	RegisterRoutes(r, NewService(nil, store), &testAuth{})
	return r, store, dir
}

func signedPath(t *testing.T, store *storage.AssetStorage, key string) string {
	t.Helper()
	uploadURL, _, err := store.GenerateUpload(key, "image/png", 7, "")
	if err != nil {
		t.Fatalf("generate upload: %v", err)
	}
	return strings.TrimPrefix(uploadURL, "/api")
}

func put(r http.Handler, path string, body []byte) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, path, bytes.NewReader(body)))
	return w
}

func TestASignedUploadIsStored(t *testing.T) {
	r, store, dir := uploadRouter(t)
	w := put(r, signedPath(t, store, uploadTestKey), []byte("png bytes"))
	if w.Code != http.StatusOK {
		t.Fatalf("signed upload: status %d %s", w.Code, w.Body.String())
	}
	got, err := os.ReadFile(filepath.Join(dir, "assets", filepath.FromSlash(uploadTestKey)))
	if err != nil || string(got) != "png bytes" {
		t.Fatalf("stored %q, %v", got, err)
	}
}

func TestAnUnsignedUploadIsRefused(t *testing.T) {
	r, _, dir := uploadRouter(t)
	w := put(r, "/asset/upload/"+uploadTestKey, []byte("x"))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("unsigned upload: status %d, want 401", w.Code)
	}
	if _, err := os.Stat(filepath.Join(dir, "assets", filepath.FromSlash(uploadTestKey))); !os.IsNotExist(err) {
		t.Fatalf("unsigned upload wrote a file: %v", err)
	}
}

func TestATamperedUploadIsRefused(t *testing.T) {
	r, store, _ := uploadRouter(t)
	signed := signedPath(t, store, uploadTestKey)
	cases := map[string]string{
		"key":  strings.Replace(signed, "shot-1", "shot-2", 1),
		"max":  strings.Replace(signed, "max=10485760", "max=99999999999", 1),
		"user": strings.Replace(signed, "uid=7", "uid=8", 1),
	}
	for name, path := range cases {
		if path == signed {
			t.Fatalf("%s: the tamper did not change the URL %s", name, signed)
		}
		if w := put(r, path, []byte("x")); w.Code != http.StatusForbidden {
			t.Errorf("tampered %s: status %d, want 403", name, w.Code)
		}
	}
}

func TestAnOversizedUploadIsRefused(t *testing.T) {
	r, store, dir := uploadRouter(t)
	body := make([]byte, store.MaxUploadBytes()+1)
	w := put(r, signedPath(t, store, uploadTestKey), body)
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversized upload: status %d, want 413", w.Code)
	}
	if _, err := os.Stat(filepath.Join(dir, "assets", filepath.FromSlash(uploadTestKey))); !os.IsNotExist(err) {
		t.Fatalf("oversized upload left a file: %v", err)
	}
}

func TestAnUploadNeverOverwrites(t *testing.T) {
	r, store, _ := uploadRouter(t)
	path := signedPath(t, store, uploadTestKey)
	if w := put(r, path, []byte("first")); w.Code != http.StatusOK {
		t.Fatalf("first upload: status %d", w.Code)
	}
	if w := put(r, path, []byte("second")); w.Code != http.StatusConflict {
		t.Fatalf("second upload to the same key: status %d, want 409", w.Code)
	}
}
