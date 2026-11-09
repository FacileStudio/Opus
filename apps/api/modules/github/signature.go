package github

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// verifySignature validates the GitHub HMAC SHA-256 signature header.
func verifySignature(payload []byte, signatureHeader, secret string) bool {
	if secret == "" {
		return true
	}
	if !strings.HasPrefix(signatureHeader, "sha256=") {
		return false
	}
	signature := strings.TrimPrefix(signatureHeader, "sha256=")
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expectedMAC))
}
