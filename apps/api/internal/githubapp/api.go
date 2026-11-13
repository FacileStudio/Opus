package githubapp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const accessTTL = 10 * time.Minute

var githubIDPattern = regexp.MustCompile(`^[0-9]{1,20}$`)

type request struct {
	method string
	path   string
	auth   string
	body   any
	etag   string
}

// Repository is what Opus stores about a linked repository.
type Repository struct {
	FullName      string `json:"full_name"`
	DefaultBranch string `json:"default_branch"`
}

type cachedPermission struct {
	level   string
	fetched time.Time
}

func (c *Client) send(ctx context.Context, req request) (*http.Response, error) {
	method := req.method
	if method == "" {
		method = http.MethodGet
	}
	var body io.Reader
	if req.body != nil {
		encoded, err := json.Marshal(req.body)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(encoded)
	}
	httpReq, err := http.NewRequestWithContext(ctx, method, c.baseURL+req.path, body)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Accept", "application/vnd.github+json")
	httpReq.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	httpReq.Header.Set("User-Agent", "opus-github-app")
	httpReq.Header.Set("Authorization", req.auth)
	if req.etag != "" {
		httpReq.Header.Set("If-None-Match", req.etag)
	}
	if body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	return c.http.Do(httpReq)
}

func (c *Client) call(ctx context.Context, req request, out any) error {
	resp, err := c.send(ctx, req)
	if err != nil {
		return fmt.Errorf("github %s: %w", req.path, err)
	}
	defer resp.Body.Close()
	if err := statusError(resp, req.path); err != nil {
		return err
	}
	return json.NewDecoder(io.LimitReader(resp.Body, 8<<20)).Decode(out)
}

func statusError(resp *http.Response, path string) error {
	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("github %s answered %d", path, resp.StatusCode)
	}
	return nil
}

// Repository returns the canonical name and default branch of repo.
func (c *Client) Repository(ctx context.Context, repo string) (Repository, error) {
	token, err := c.token(ctx, repo, "metadata", "read")
	if err != nil {
		return Repository{}, err
	}
	var out Repository
	err = c.call(ctx, request{path: "/repos/" + repo, auth: "Bearer " + token}, &out)
	return out, err
}

// Permission returns the GitHub account's permission on repo: admin,
// write, read or none. The answer is cached ten minutes per account and
// repository.
func (c *Client) Permission(ctx context.Context, repo, githubID string) (string, error) {
	if !githubIDPattern.MatchString(githubID) {
		return "none", nil
	}
	cacheKey := githubID + "|" + strings.ToLower(repo)
	c.mu.Lock()
	cached, ok := c.access[cacheKey]
	c.mu.Unlock()
	if ok && c.now().Sub(cached.fetched) < accessTTL {
		return cached.level, nil
	}
	level, err := c.fetchPermission(ctx, repo, githubID)
	if err != nil {
		return "", err
	}
	c.mu.Lock()
	c.access[cacheKey] = cachedPermission{level: level, fetched: c.now()}
	c.mu.Unlock()
	return level, nil
}

func (c *Client) fetchPermission(ctx context.Context, repo, githubID string) (string, error) {
	token, err := c.token(ctx, repo, "metadata", "read")
	if err != nil {
		return "", err
	}
	auth := "Bearer " + token
	var user struct {
		Login string `json:"login"`
	}
	if err := c.call(ctx, request{path: "/user/" + githubID, auth: auth}, &user); err != nil {
		if errors.Is(err, ErrNotFound) {
			return "none", nil
		}
		return "", err
	}
	var permission struct {
		Permission string `json:"permission"`
	}
	path := "/repos/" + repo + "/collaborators/" + user.Login + "/permission"
	if err := c.call(ctx, request{path: path, auth: auth}, &permission); err != nil {
		if errors.Is(err, ErrNotFound) {
			return "none", nil
		}
		return "", err
	}
	return permission.Permission, nil
}

// CanWrite reports whether a permission level allows writing.
func CanWrite(level string) bool {
	return level == "admin" || level == "write" || level == "maintain"
}
