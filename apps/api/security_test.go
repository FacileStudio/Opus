package main

import (
	"log/slog"
	"slices"
	"strings"
	"testing"

	"github.com/FacileStudio/Opus/apps/api/internal/storage"
)

func TestSecurityConfigAllowsCatalogAndS3(t *testing.T) {
	cfg := securityConfig(t.TempDir(), &storage.S3Config{Endpoint: "https://s3.example.com/bucket"}, slog.New(slog.DiscardHandler))
	connect := cfg.CSPSources["connect-src"]
	if !slices.Contains(connect, catalogOrigin) || !slices.Contains(connect, "https://s3.example.com") {
		t.Fatalf("connect-src = %v", connect)
	}
}

func TestSecurityConfigAllowsNoVisionScript(t *testing.T) {
	cfg := securityConfig(t.TempDir(), nil, slog.New(slog.DiscardHandler))
	if slices.ContainsFunc(cfg.CSPSources["script-src"], func(source string) bool {
		return strings.Contains(source, "vision.facile.studio")
	}) {
		t.Fatalf("script-src = %v", cfg.CSPSources["script-src"])
	}
}
