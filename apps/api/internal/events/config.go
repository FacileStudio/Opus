package events

import (
	"os"
)

// Config carries Antenne client connection settings.
type Config struct {
	URL        string
	Secret     string
	InstanceID string
}

// LoadConfig reads Antenne configuration from standard environment variables.
func LoadConfig() Config {
	return Config{
		URL:        os.Getenv("ANTENNE_URL"),
		Secret:     os.Getenv("ANTENNE_SECRET"),
		InstanceID: os.Getenv("ANTENNE_INSTANCE_ID"),
	}
}

// Enabled reports whether Antenne publishing is configured.
func (c Config) Enabled() bool {
	return c.URL != "" && c.Secret != ""
}
