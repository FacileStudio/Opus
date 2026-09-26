// Package env loads the Opus API configuration from the environment.
package env

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/FacileStudio/porte"
	troncenv "github.com/FacileStudio/tronc/env"
	"github.com/joho/godotenv"
)

// OIDCConfig holds the OIDC settings read from the environment when
// OIDC_ISSUER is set.
type OIDCConfig struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	SuccessURL   string
}

// Porte returns the shared auth kit's configuration. The environment variable
// names are the suite convention and do not change; this only reshapes what
// Load already read.
func (c Config) Porte() porte.Config {
	if c.OIDC == nil {
		return porte.Config{SSOOnly: c.SSOOnly}
	}
	return porte.Config{
		Issuer:      c.OIDC.Issuer,
		ClientID:    c.OIDC.ClientID,
		ClientSecret: c.OIDC.ClientSecret,
		RedirectURL: c.OIDC.RedirectURL,
		SuccessURL:  c.OIDC.SuccessURL,
		SSOOnly:     c.SSOOnly,
	}
}

// Config is the fully-loaded application configuration produced by Load.
type Config struct {
	troncenv.Core

	// LegacyAPI is the transitional TypeScript API every /api route the Go
	// backend does not own yet is proxied to. nil when not configured.
	LegacyAPI *url.URL

	// UploadDir files uploaded assets under, workspace logos today.
	UploadDir string

	// AuthBridgeSecret is the shared secret the Go proxy and the legacy TS API
	// both read. The proxy stamps proxied requests with an HMAC derived from it
	// so the legacy API can trust the injected user headers without exposing
	// the action to a caller who cannot name the user. Empty disables the
	// bridge: no identity headers are injected and proxied routes keep their
	// legacy 401 behaviour.
	AuthBridgeSecret string

	// OIDC and SSOOnly shape the porte auth kit.
	OIDC    *OIDCConfig
	SSOOnly bool
}

func parseLegacyAPI() (*url.URL, error) {
	raw := troncenv.String("LEGACY_API_URL", "")
	if raw == "" {
		return nil, nil
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("LEGACY_API_URL must be a valid URL")
	}
	return parsed, nil
}

func parseOIDCConfig(corsOrigins []string) (*OIDCConfig, error) {
	issuer := troncenv.String("OIDC_ISSUER", "")
	if issuer == "" {
		return nil, nil
	}
	clientID := troncenv.String("OIDC_CLIENT_ID", "")
	clientSecret := troncenv.String("OIDC_CLIENT_SECRET", "")
	redirectURL := troncenv.String("OIDC_REDIRECT_URL", "")
	if clientID == "" || clientSecret == "" || redirectURL == "" {
		return nil, fmt.Errorf("OIDC_CLIENT_ID, OIDC_CLIENT_SECRET, and OIDC_REDIRECT_URL are required when OIDC_ISSUER is set")
	}
	successURL := troncenv.String("OIDC_SUCCESS_URL", "")
	if successURL == "" && len(corsOrigins) > 0 {
		successURL = corsOrigins[0]
	}
	return &OIDCConfig{
		Issuer:       issuer,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		SuccessURL:   successURL,
	}, nil
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
	legacyAPI, err := parseLegacyAPI()
	if err != nil {
		return Config{}, err
	}
	ssoOnly, err := troncenv.Bool("SSO_ONLY", false)
	if err != nil {
		return Config{}, err
	}
	oidcConfig, err := parseOIDCConfig(core.CORSAllowedOrigins)
	if err != nil {
		return Config{}, err
	}
	return Config{
		Core:             core,
		LegacyAPI:        legacyAPI,
		UploadDir:        troncenv.String("UPLOAD_DIR", "./data/uploads"),
		AuthBridgeSecret: troncenv.String("AUTH_BRIDGE_SECRET", ""),
		SSOOnly:          ssoOnly,
		OIDC:             oidcConfig,
	}, nil
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
