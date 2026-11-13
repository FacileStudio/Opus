// Package env loads the Opus API configuration from the environment.
package env

import (
	"crypto/sha256"
	"fmt"
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

	// CLIAudience is the client id the suite's CLIs hold their device-grant
	// token under. Setting it mounts porte's device exchange; empty leaves
	// the route unmounted and answering 404. It is never the machine
	// audience, which names this app.
	CLIAudience     string
	MachineAudience string
	ScopePrefix     string
}

// Porte returns the shared auth kit's configuration. The environment variable
// names are the suite convention and do not change; this only reshapes what
// Load already read.
func (c Config) Porte() porte.Config {
	if c.OIDC == nil {
		return porte.Config{}
	}
	return porte.Config{
		Issuer:       c.OIDC.Issuer,
		ClientID:     c.OIDC.ClientID,
		ClientSecret: c.OIDC.ClientSecret,
		RedirectURL:  c.OIDC.RedirectURL,
		SuccessURL:   c.OIDC.SuccessURL,

		CLIAudience:     c.OIDC.CLIAudience,
		MachineAudience: c.OIDC.MachineAudience,
		ScopePrefix:     c.OIDC.ScopePrefix,
		ExtraScopes:     []string{"connections"},
		ExtraClaims:     []string{"connections"},
	}
}

// Config is the fully-loaded application configuration produced by Load.
type Config struct {
	troncenv.Core

	// UploadDir files uploaded assets under, workspace logos today.
	UploadDir string

	// RedisURL configures the Redis pub/sub broker for realtime broadcasts.
	RedisURL string

	// GitHubWebhookSecret validates incoming GitHub webhooks.
	GitHubWebhookSecret string

	// GitHubApp is the Opus GitHub App. Both fields empty turn the linked
	// repositories and the issues view off; they then answer 412.
	GitHubApp GitHubAppConfig

	// BulleURL is Bulle's origin. Empty leaves the Ideas routes unmounted.
	BulleURL string

	// OIDC shapes the porte auth kit. Sign-in goes through Registre only.
	OIDC *OIDCConfig

	// MCP shapes the MCP server. It is off unless OIDC is on and
	// MCP_RESOURCE is set.
	MCP MCPConfig
}

// GitHubAppConfig holds the App's client id, the JWT issuer, and its
// private key, the PEM file base64-encoded.
type GitHubAppConfig struct {
	ClientID   string
	PrivateKey string
}

// MCPConfig holds the MCP server settings.
type MCPConfig struct {
	// Resource is MCP_RESOURCE, the canonical URI of the endpoint, such as
	// https://opus.facile.studio/mcp. Registre binds tokens to it.
	Resource string

	// StateKey signs the confirmation questions handed to clients. It is
	// derived from the OIDC client secret, so it is secret, stable across
	// restarts and needs no variable of its own.
	StateKey []byte
}

// UploadKey signs local upload URLs. It is derived from the OIDC client
// secret, so it is secret and stable across restarts and replicas without a
// variable of its own; nil without OIDC, where a per-process key is used.
func (c Config) UploadKey() []byte {
	if c.OIDC == nil || c.OIDC.ClientSecret == "" {
		return nil
	}
	key := sha256.Sum256([]byte("opus-upload-sign:" + c.OIDC.ClientSecret))
	return key[:]
}

// mcpConfig reads MCP_RESOURCE and derives the state key.
func mcpConfig(oidc *OIDCConfig) MCPConfig {
	resource := strings.TrimSpace(troncenv.String("MCP_RESOURCE", ""))
	if oidc == nil || resource == "" {
		return MCPConfig{}
	}
	key := sha256.Sum256([]byte("opus-mcp-state:" + oidc.ClientSecret))
	return MCPConfig{Resource: resource, StateKey: key[:]}
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

		CLIAudience:     troncenv.String("OIDC_CLI_AUDIENCE", ""),
		MachineAudience: troncenv.String("OIDC_MACHINE_AUDIENCE", ""),
		ScopePrefix:     troncenv.String("OIDC_SCOPE_PREFIX", ""),
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
	oidcConfig, err := parseOIDCConfig(core.CORSAllowedOrigins)
	if err != nil {
		return Config{}, err
	}
	return Config{
		Core:                core,
		UploadDir:           troncenv.String("UPLOAD_DIR", "./data/uploads"),
		RedisURL:            troncenv.String("REDIS_URL", ""),
		GitHubWebhookSecret: troncenv.String("GITHUB_WEBHOOK_SECRET", ""),
		GitHubApp: GitHubAppConfig{
			ClientID:   troncenv.String("OPUS_GITHUB_APP_CLIENT_ID", ""),
			PrivateKey: troncenv.String("OPUS_GITHUB_APP_PRIVATE_KEY", ""),
		},
		BulleURL: strings.TrimSuffix(troncenv.String("BULLE_URL", ""), "/"),
		OIDC:     oidcConfig,
		MCP:      mcpConfig(oidcConfig),
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
