// Package githubapp talks to GitHub as the Opus GitHub App. It signs the App
// JWT, finds the installation that covers a repository and mints installation
// tokens scoped to that one repository and the one permission a call needs.
// Tokens and keys never leave this package and are never logged.
package githubapp

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/go-jose/go-jose/v4"
)

var (
	// ErrNotInstalled means the App has no installation covering the repository.
	ErrNotInstalled = errors.New("the GitHub App is not installed on this repository")
	// ErrNotFound means GitHub answered 404.
	ErrNotFound = errors.New("not found on GitHub")
	// ErrInvalidRepo means the name is not owner/repo.
	ErrInvalidRepo = errors.New("a repository is written owner/repo")
)

const (
	defaultBaseURL = "https://api.github.com"
	tokenMargin    = 5 * time.Minute
)

var repoPattern = regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9-]{0,38})/[A-Za-z0-9._-]{1,100}$`)

// Client is the App's GitHub client. A nil *Client means the App is not
// configured; callers check for nil before using it.
type Client struct {
	clientID string
	key      *rsa.PrivateKey
	baseURL  string
	http     *http.Client
	now      func() time.Time

	mu     sync.Mutex
	tokens map[string]cachedToken
	access map[string]cachedPermission
	issues map[string]cachedIssues
}

type cachedToken struct {
	value   string
	expires time.Time
}

// New builds the client from the App's client id and its private key, the
// PEM file base64-encoded (a raw PEM is accepted too). Both empty returns a
// nil client and no error: the App is simply off.
func New(clientID, encodedKey string) (*Client, error) {
	clientID, encodedKey = strings.TrimSpace(clientID), strings.TrimSpace(encodedKey)
	if clientID == "" && encodedKey == "" {
		return nil, nil
	}
	if clientID == "" || encodedKey == "" {
		return nil, errors.New("OPUS_GITHUB_APP_CLIENT_ID and OPUS_GITHUB_APP_PRIVATE_KEY must be set together")
	}
	key, err := parseKey(encodedKey)
	if err != nil {
		return nil, err
	}
	return &Client{
		clientID: clientID,
		key:      key,
		baseURL:  defaultBaseURL,
		http:     &http.Client{Timeout: 15 * time.Second},
		now:      time.Now,
		tokens:   map[string]cachedToken{},
		access:   map[string]cachedPermission{},
		issues:   map[string]cachedIssues{},
	}, nil
}

// WithBaseURL points the client at another API root, a test fake.
func (c *Client) WithBaseURL(baseURL string) *Client {
	c.baseURL = strings.TrimSuffix(baseURL, "/")
	return c
}

// ValidRepo reports whether name is a well-formed owner/repo.
func ValidRepo(name string) bool {
	return repoPattern.MatchString(name)
}

func parseKey(encoded string) (*rsa.PrivateKey, error) {
	raw := []byte(encoded)
	if !strings.HasPrefix(encoded, "-----BEGIN") {
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return nil, errors.New("OPUS_GITHUB_APP_PRIVATE_KEY is not base64")
		}
		raw = decoded
	}
	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, errors.New("OPUS_GITHUB_APP_PRIVATE_KEY holds no PEM block")
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.New("OPUS_GITHUB_APP_PRIVATE_KEY is not an RSA private key")
	}
	key, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("OPUS_GITHUB_APP_PRIVATE_KEY is not an RSA private key")
	}
	return key, nil
}

func (c *Client) appJWT() (string, error) {
	now := c.now()
	claims, err := json.Marshal(map[string]any{
		"iat": now.Add(-time.Minute).Unix(),
		"exp": now.Add(9 * time.Minute).Unix(),
		"iss": c.clientID,
	})
	if err != nil {
		return "", err
	}
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: c.key}, (&jose.SignerOptions{}).WithType("JWT"))
	if err != nil {
		return "", err
	}
	signed, err := signer.Sign(claims)
	if err != nil {
		return "", err
	}
	return signed.CompactSerialize()
}

func (c *Client) token(ctx context.Context, repo, permission, level string) (string, error) {
	if !ValidRepo(repo) {
		return "", ErrInvalidRepo
	}
	cacheKey := strings.ToLower(repo) + "|" + permission + ":" + level
	c.mu.Lock()
	cached, ok := c.tokens[cacheKey]
	c.mu.Unlock()
	if ok && c.now().Before(cached.expires.Add(-tokenMargin)) {
		return cached.value, nil
	}
	minted, err := c.mint(ctx, repo, permission, level)
	if err != nil {
		return "", err
	}
	c.mu.Lock()
	c.tokens[cacheKey] = minted
	c.mu.Unlock()
	return minted.value, nil
}

func (c *Client) mint(ctx context.Context, repo, permission, level string) (cachedToken, error) {
	jwt, err := c.appJWT()
	if err != nil {
		return cachedToken{}, fmt.Errorf("sign the GitHub App JWT: %w", err)
	}
	var installation struct {
		ID int64 `json:"id"`
	}
	if err := c.call(ctx, request{path: "/repos/" + repo + "/installation", auth: "Bearer " + jwt}, &installation); err != nil {
		if errors.Is(err, ErrNotFound) {
			return cachedToken{}, ErrNotInstalled
		}
		return cachedToken{}, err
	}
	_, name, _ := strings.Cut(repo, "/")
	body := map[string]any{
		"repositories": []string{name},
		"permissions":  map[string]string{permission: level},
	}
	var minted struct {
		Token     string    `json:"token"`
		ExpiresAt time.Time `json:"expires_at"`
	}
	path := fmt.Sprintf("/app/installations/%d/access_tokens", installation.ID)
	if err := c.call(ctx, request{method: http.MethodPost, path: path, auth: "Bearer " + jwt, body: body}, &minted); err != nil {
		return cachedToken{}, err
	}
	return cachedToken{value: minted.Token, expires: minted.ExpiresAt}, nil
}
