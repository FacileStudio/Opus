// Package env loads the Opus API configuration from the environment.
package env

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	troncenv "github.com/FacileStudio/tronc/env"
	"github.com/joho/godotenv"
)

// Config is the fully-loaded application configuration produced by Load.
type Config struct {
	troncenv.Core

	// LegacyAPI is the transitional TypeScript API every /api route the Go
	// backend does not own yet is proxied to. nil when not configured.
	LegacyAPI *url.URL
}

// Load reads the .env file (if present) and the environment into a Config,
// validating PORT and LOG_LEVEL before returning it.
func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return Config{}, fmt.Errorf("failed to parse .env file: %w", err)
	}
	core, err := troncenv.LoadCore()
	if err != nil {
		return Config{}, err
	}
	if core.Port < 1 || core.Port > 65535 {
		return Config{}, fmt.Errorf("PORT must be a valid TCP port")
	}
	if !validLogLevel(core.LogLevel) {
		return Config{}, fmt.Errorf("LOG_LEVEL must be one of debug, info, warn, error")
	}
	var legacyAPI *url.URL
	if raw := troncenv.String("LEGACY_API_URL", ""); raw != "" {
		parsed, err := url.Parse(raw)
		if err != nil {
			return Config{}, fmt.Errorf("LEGACY_API_URL must be a valid URL")
		}
		legacyAPI = parsed
	}
	return Config{Core: core, LegacyAPI: legacyAPI}, nil
}

// validLogLevel reports whether level names a level the suite logger knows.
func validLogLevel(level string) bool {
	switch strings.ToLower(level) {
	case "debug", "info", "warn", "error":
		return true
	default:
		return false
	}
}
