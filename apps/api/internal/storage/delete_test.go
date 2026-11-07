package storage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestS3DeleteIsBatchedAndTreatsNotFoundAsDone(t *testing.T) {
	var mu sync.Mutex
	seen := map[string]bool{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		seen[r.URL.Path] = true
		mu.Unlock()
		if r.Method != http.MethodDelete || r.URL.Query().Get("X-Amz-Signature") == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if strings.HasSuffix(r.URL.Path, "gone.png") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	cfg := &S3Config{Endpoint: server.URL, Bucket: "b", AccessKeyID: "k", SecretAccessKey: "s", Region: "us-east-1", ForcePathStyle: true}
	keys := []string{"gone.png"}
	for i := range 40 {
		keys = append(keys, "w/"+string(rune('a'+i%26))+strings.Repeat("x", i)+".png")
	}
	if err := NewAssetStorage(t.TempDir(), cfg).Delete(context.Background(), keys); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if len(seen) != len(keys) {
		t.Fatalf("requests for %d distinct keys, want %d", len(seen), len(keys))
	}
}

func TestS3DeleteReportsServerErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	cfg := &S3Config{Endpoint: server.URL, Bucket: "b", AccessKeyID: "k", SecretAccessKey: "s", Region: "us-east-1", ForcePathStyle: true}
	if err := NewAssetStorage(t.TempDir(), cfg).Delete(context.Background(), []string{"a.png"}); err == nil {
		t.Fatal("a 500 must be reported")
	}
}

func TestLocalDeleteIsIdempotentAndRefusesTraversal(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "assets", "w", "a.png")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	store := NewAssetStorage(dir, nil)
	for range 2 {
		if err := store.Delete(context.Background(), []string{"w/a.png"}); err != nil {
			t.Fatalf("Delete: %v", err)
		}
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("file still there: %v", err)
	}
	if err := store.Delete(context.Background(), []string{"../../etc/passwd"}); err == nil {
		t.Fatal("traversal key must be refused")
	}
}
