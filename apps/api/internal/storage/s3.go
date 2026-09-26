package storage

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/FacileStudio/tronc/errors"
)

// S3Config holds connection and bucket details for S3-compatible storage.
type S3Config struct {
	Endpoint            string
	Bucket              string
	AccessKeyID         string
	SecretAccessKey     string
	Region              string
	ForcePathStyle      bool
	PublicBaseURL       string
	MaxImageUploadBytes int64
	PresignTTLSeconds   int
}

// LoadS3Config reads S3 settings from environment variables.
// Returns nil if the core S3 credentials or endpoint are unset.
func LoadS3Config() *S3Config {
	endpoint := strings.TrimSpace(os.Getenv("S3_ENDPOINT"))
	bucket := strings.TrimSpace(os.Getenv("S3_BUCKET"))
	accessKey := strings.TrimSpace(os.Getenv("S3_ACCESS_KEY_ID"))
	secretKey := strings.TrimSpace(os.Getenv("S3_SECRET_ACCESS_KEY"))
	if endpoint == "" || bucket == "" || accessKey == "" || secretKey == "" {
		return nil
	}
	region := strings.TrimSpace(os.Getenv("S3_REGION"))
	if region == "" {
		region = "us-east-1"
	}
	forcePathStyle := true
	if val := strings.TrimSpace(os.Getenv("S3_FORCE_PATH_STYLE")); val != "" {
		forcePathStyle = strings.ToLower(val) == "true"
	}
	maxBytes := int64(10 * 1024 * 1024)
	if parsed, err := strconv.ParseInt(os.Getenv("S3_MAX_IMAGE_UPLOAD_BYTES"), 10, 64); err == nil && parsed > 0 {
		maxBytes = parsed
	}
	ttl := 300
	if parsed, err := strconv.Atoi(os.Getenv("S3_PRESIGN_TTL_SECONDS")); err == nil && parsed > 0 {
		ttl = parsed
	}
	return &S3Config{
		Endpoint:            endpoint,
		Bucket:              bucket,
		AccessKeyID:         accessKey,
		SecretAccessKey:     secretKey,
		Region:              region,
		ForcePathStyle:      forcePathStyle,
		PublicBaseURL:       strings.TrimSpace(os.Getenv("S3_PUBLIC_BASE_URL")),
		MaxImageUploadBytes: maxBytes,
		PresignTTLSeconds:   ttl,
	}
}

func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func deriveSigningKey(secretKey, dateStamp, region string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secretKey), []byte(dateStamp))
	kRegion := hmacSHA256(kDate, []byte(region))
	kService := hmacSHA256(kRegion, []byte("s3"))
	return hmacSHA256(kService, []byte("aws4_request"))
}

func buildCanonicalQueryString(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		val := params[k]
		escapedKey := strings.ReplaceAll(url.QueryEscape(k), "+", "%20")
		escapedVal := strings.ReplaceAll(url.QueryEscape(val), "+", "%20")
		pairs = append(pairs, fmt.Sprintf("%s=%s", escapedKey, escapedVal))
	}
	return strings.Join(pairs, "&")
}

type s3Target struct {
	Scheme  string
	Host    string
	RawPath string
}

func resolveHostAndPath(cfg *S3Config, objectKey string) (*s3Target, error) {
	parsed, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return nil, err
	}
	cleanKey := strings.TrimPrefix(objectKey, "/")
	escapedSegments := make([]string, 0)
	for _, seg := range strings.Split(cleanKey, "/") {
		escapedSegments = append(escapedSegments, url.PathEscape(seg))
	}
	keyPath := strings.Join(escapedSegments, "/")
	if cfg.ForcePathStyle {
		return &s3Target{
			Scheme:  parsed.Scheme,
			Host:    parsed.Host,
			RawPath: fmt.Sprintf("/%s/%s", cfg.Bucket, keyPath),
		}, nil
	}
	return &s3Target{
		Scheme:  parsed.Scheme,
		Host:    fmt.Sprintf("%s.%s", cfg.Bucket, parsed.Host),
		RawPath: fmt.Sprintf("/%s", keyPath),
	}, nil
}

// PresignURL generates an AWS SigV4 presigned URL for S3 object operations.
func PresignURL(cfg *S3Config, method, objectKey, contentType string, ttlSeconds int) (string, error) {
	target, err := resolveHostAndPath(cfg, objectKey)
	if err != nil {
		return "", err
	}
	scheme, host, rawPath := target.Scheme, target.Host, target.RawPath
	now := time.Now().UTC()
	dateStamp := now.Format("20060102")
	amzDate := now.Format("20060102T150405Z")
	scope := fmt.Sprintf("%s/%s/s3/aws4_request", dateStamp, cfg.Region)
	signedHeaders := "host"
	canonicalHeaders := fmt.Sprintf("host:%s\n", host)
	if contentType != "" && method == http.MethodPut {
		signedHeaders = "content-type;host"
		canonicalHeaders = fmt.Sprintf("content-type:%s\nhost:%s\n", strings.TrimSpace(contentType), host)
	}
	queryParams := map[string]string{
		"X-Amz-Algorithm":     "AWS4-HMAC-SHA256",
		"X-Amz-Credential":    fmt.Sprintf("%s/%s", cfg.AccessKeyID, scope),
		"X-Amz-Date":          amzDate,
		"X-Amz-Expires":       strconv.Itoa(ttlSeconds),
		"X-Amz-SignedHeaders": signedHeaders,
	}
	canonicalQuery := buildCanonicalQueryString(queryParams)
	canonicalReq := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\nUNSIGNED-PAYLOAD", method, rawPath, canonicalQuery, canonicalHeaders, signedHeaders)
	stringToSign := fmt.Sprintf("AWS4-HMAC-SHA256\n%s\n%s\n%s", amzDate, scope, sha256Hex([]byte(canonicalReq)))
	signingKey := deriveSigningKey(cfg.SecretAccessKey, dateStamp, cfg.Region)
	signature := hex.EncodeToString(hmacSHA256(signingKey, []byte(stringToSign)))
	return fmt.Sprintf("%s://%s%s?%s&X-Amz-Signature=%s", scheme, host, rawPath, canonicalQuery, signature), nil
}

// FetchS3Object retrieves an object stream from S3 using a presigned GET request.
func FetchS3Object(ctx context.Context, cfg *S3Config, objectKey string) (*ObjectStream, error) {
	presignedURL, err := PresignURL(cfg, http.MethodGet, objectKey, "", 300)
	if err != nil {
		return nil, errors.Internal("failed to presign s3 download url", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, presignedURL, nil)
	if err != nil {
		return nil, errors.Internal("failed to create s3 get request", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Internal("failed to execute s3 get request", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil, errors.NotFound("Asset object not found")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		resp.Body.Close()
		return nil, errors.Internal(fmt.Sprintf("s3 get request failed with status %d", resp.StatusCode), nil)
	}
	lastModified, _ := http.ParseTime(resp.Header.Get("Last-Modified"))
	return &ObjectStream{
		Body:          resp.Body,
		ContentType:   resp.Header.Get("Content-Type"),
		ContentLength: resp.ContentLength,
		ETag:          resp.Header.Get("ETag"),
		LastModified:  lastModified,
	}, nil
}
