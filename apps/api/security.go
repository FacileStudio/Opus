package main

import (
	"log/slog"
	"net/url"

	"github.com/FacileStudio/Opus/apps/api/internal/storage"
	troncmiddleware "github.com/FacileStudio/tronc/middleware"
	"github.com/FacileStudio/tronc/spa"
)

// embedFrames lists the origins the task editor's embed block may frame.
var embedFrames = []string{"https://www.youtube.com", "https://www.youtube-nocookie.com"}

// visionOrigin serves the Vision tracker that index.html loads on the production host.
const visionOrigin = "https://vision.facile.studio"

// securityConfig extends tronc's default security headers with what the built
// client needs: the inline scripts of its pages, which script-src 'self' would
// otherwise block, the Vision tracker's origin, the editor's embed frames, and
// the S3 origin presigned uploads go to. A client directory that cannot be hashed is logged and leaves
// the default script policy, which only reports.
func securityConfig(clientDir string, s3 *storage.S3Config, appLogger *slog.Logger) troncmiddleware.SecurityConfig {
	sources := map[string][]string{"frame-src": embedFrames, "script-src": {visionOrigin}}
	if origin := s3Origin(s3); origin != "" {
		sources["connect-src"] = []string{origin}
		sources["img-src"] = []string{origin}
	}
	if spa.Available(clientDir) {
		hashes, err := spa.InlineScriptHashes(clientDir)
		if err != nil {
			appLogger.Warn("failed to hash the client's inline scripts", slog.Any("error", err))
		} else {
			sources["script-src"] = append(sources["script-src"], hashes...)
		}
	}
	return troncmiddleware.SecurityConfig{CSPSources: sources}
}

func s3Origin(s3 *storage.S3Config) string {
	if s3 == nil {
		return ""
	}
	parsed, err := url.Parse(s3.Endpoint)
	if err != nil || parsed.Host == "" {
		return ""
	}
	return parsed.Scheme + "://" + parsed.Host
}
