package githubapp

import (
	"context"
	"strings"
	"time"
)

const installRetry = time.Minute

type installLink struct {
	url    string
	failed time.Time
}

// InstallURL returns the page where any GitHub account or organization
// installs the App, read once from GET /app and cached. A failed lookup
// returns "" and is retried after a minute. A nil client returns "".
func (c *Client) InstallURL(ctx context.Context) string {
	if c == nil {
		return ""
	}
	c.mu.Lock()
	cached := c.install
	c.mu.Unlock()
	if cached.url != "" {
		return cached.url
	}
	if !cached.failed.IsZero() && c.now().Sub(cached.failed) < installRetry {
		return ""
	}
	url := c.fetchInstallURL(ctx)
	c.mu.Lock()
	c.install = installLink{url: url}
	if url == "" {
		c.install.failed = c.now()
	}
	c.mu.Unlock()
	return url
}

func (c *Client) fetchInstallURL(ctx context.Context) string {
	jwt, err := c.appJWT()
	if err != nil {
		return ""
	}
	var app struct {
		HTMLURL string `json:"html_url"`
	}
	if err := c.call(ctx, request{path: "/app", auth: "Bearer " + jwt}, &app); err != nil {
		return ""
	}
	if !strings.HasPrefix(app.HTMLURL, "https://") {
		return ""
	}
	return strings.TrimSuffix(app.HTMLURL, "/") + "/installations/new"
}
