package events

import (
	"errors"
	"os"
	"strconv"
	"strings"

	antenne "github.com/FacileStudio/antenne-client/go"
	envelope "github.com/FacileStudio/antenne-client/go/envelope"
)

// Config carries Antenne client connection settings. SyncWorkspaceID only
// places projects from old Sablier payloads that carry no space. Issuer is
// the OIDC issuer whose porte identities give owner and actor subjects.
type Config struct {
	URL        string
	Secret     string
	InstanceID string

	SyncWorkspaceID int64
	Issuer          string
}

// LoadConfig reads the Antenne settings through antenne.ConfigFromEnv:
// ANTENNE_URL, or its older synonym ANTENNE_INSTANCE, ANTENNE_SECRET and
// ANTENNE_INSTANCE_ID. No URL turns the bus off; a URL without a secret is an
// error. Opus always registers as the Opus app, whatever ANTENNE_APP says.
func LoadConfig() (Config, error) {
	cfg := Config{SyncWorkspaceID: parseWorkspaceID(os.Getenv("ANTENNE_SYNC_WORKSPACE_ID"))}
	client, err := antenne.ConfigFromEnv(string(envelope.AppOpus))
	switch {
	case errors.Is(err, antenne.ErrNotConfigured):
		return cfg, nil
	case err != nil:
		return cfg, err
	}
	cfg.URL, cfg.Secret, cfg.InstanceID = client.Instance, client.Secret, client.InstanceID
	return cfg, nil
}

// Enabled reports whether Antenne publishing is configured.
func (c Config) Enabled() bool {
	return c.URL != "" && c.Secret != ""
}

func parseWorkspaceID(raw string) int64 {
	id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || id <= 0 {
		return 0
	}
	return id
}
