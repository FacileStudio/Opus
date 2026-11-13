package storage

import (
	"net/http"
	"testing"
	"time"

	"github.com/FacileStudio/tronc/errors"
)

func TestAnExpiredUploadURLIsRefused(t *testing.T) {
	store := NewAssetStorage(t.TempDir(), nil)
	signedAt := time.Now()
	query := store.signUpload("a/b.png", 1, signedAt)
	if _, err := store.VerifyUpload("a/b.png", query, signedAt.Add(UploadURLLifetime-time.Minute)); err != nil {
		t.Fatalf("a fresh URL was refused: %v", err)
	}
	_, err := store.VerifyUpload("a/b.png", query, signedAt.Add(UploadURLLifetime+time.Minute))
	if err == nil || errors.Status(err) != http.StatusForbidden {
		t.Fatalf("an expired URL: %v, want 403", err)
	}
}

func TestAnotherKeyCannotVerify(t *testing.T) {
	signer := NewAssetStorage(t.TempDir(), nil)
	verifier := NewAssetStorage(t.TempDir(), nil)
	query := signer.signUpload("a/b.png", 1, time.Now())
	if _, err := verifier.VerifyUpload("a/b.png", query, time.Now()); err == nil {
		t.Fatal("a URL signed with another key was accepted")
	}
	verifier.SetUploadKey(signer.uploadKey)
	if _, err := verifier.VerifyUpload("a/b.png", query, time.Now()); err != nil {
		t.Fatalf("the same key was refused: %v", err)
	}
}
