package githubapp

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	issuesTTL     = 60 * time.Second
	issuesPerRepo = 100
)

// Issue is one GitHub issue as Opus shows it.
type Issue struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	HTMLURL   string    `json:"html_url"`
	State     string    `json:"state"`
	Comments  int       `json:"comments"`
	Labels    []Label   `json:"labels"`
	User      *User     `json:"user"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Label is an issue label.
type Label struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// User is an issue author.
type User struct {
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
	HTMLURL   string `json:"html_url"`
}

type rawIssue struct {
	Issue
	PullRequest json.RawMessage `json:"pull_request"`
}

type cachedIssues struct {
	etag    string
	items   []Issue
	fetched time.Time
}

// Issues returns up to 100 issues of repo in the given state, open or
// closed, without pull requests. Answers are cached 60 seconds, then
// revalidated with the ETag GitHub returned.
func (c *Client) Issues(ctx context.Context, repo, state string) ([]Issue, error) {
	cacheKey := strings.ToLower(repo) + "|" + state
	c.mu.Lock()
	cached, ok := c.issues[cacheKey]
	c.mu.Unlock()
	if ok && c.now().Sub(cached.fetched) < issuesTTL {
		return cached.items, nil
	}
	fresh, err := c.fetchIssues(ctx, repo, state, cached)
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	c.issues[cacheKey] = fresh
	c.mu.Unlock()
	return fresh.items, nil
}

func (c *Client) fetchIssues(ctx context.Context, repo, state string, cached cachedIssues) (cachedIssues, error) {
	token, err := c.token(ctx, repo, "issues", "read")
	if err != nil {
		return cachedIssues{}, err
	}
	path := "/repos/" + repo + "/issues?state=" + state + "&per_page=100"
	resp, err := c.send(ctx, request{path: path, auth: "Bearer " + token, etag: cached.etag})
	if err != nil {
		return cachedIssues{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotModified && cached.etag != "" {
		cached.fetched = c.now()
		return cached, nil
	}
	if err := statusError(resp, "/repos/"+repo+"/issues"); err != nil {
		return cachedIssues{}, err
	}
	var raw []rawIssue
	if err := json.NewDecoder(io.LimitReader(resp.Body, 16<<20)).Decode(&raw); err != nil {
		return cachedIssues{}, err
	}
	return cachedIssues{etag: resp.Header.Get("ETag"), items: withoutPullRequests(raw), fetched: c.now()}, nil
}

func withoutPullRequests(raw []rawIssue) []Issue {
	items := make([]Issue, 0, len(raw))
	for _, item := range raw {
		if len(item.PullRequest) > 0 && string(item.PullRequest) != "null" {
			continue
		}
		if len(items) == issuesPerRepo {
			break
		}
		items = append(items, item.Issue)
	}
	return items
}
