package events

import (
	"os"
	"strconv"
	"strings"
)

// Config carries Antenne client connection settings.
type Config struct {
	URL        string
	Secret     string
	InstanceID string

	SyncWorkspaceID int64
}

// LoadConfig reads Antenne configuration from standard environment variables.
func LoadConfig() Config {
	return Config{
		URL:        os.Getenv("ANTENNE_URL"),
		Secret:     os.Getenv("ANTENNE_SECRET"),
		InstanceID: os.Getenv("ANTENNE_INSTANCE_ID"),

		SyncWorkspaceID: parseWorkspaceID(os.Getenv("ANTENNE_SYNC_WORKSPACE_ID")),
	}
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
