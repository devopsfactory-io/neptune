package verify

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

const signaturePrefix = "sha256="

// WebhookSignature verifies the GitHub webhook signature in X-Hub-Signature-256.
// Body is the raw request body (e.g. []byte). Secret is the webhook secret.
// Returns nil if the signature is valid.
func WebhookSignature(body []byte, signatureHeader, secret string) error {
	if signatureHeader == "" {
		return errors.New("missing X-Hub-Signature-256")
	}
	if secret == "" {
		return errors.New("webhook secret is empty")
	}
	sig := strings.TrimSpace(signatureHeader)
	if !strings.HasPrefix(strings.ToLower(sig), signaturePrefix) {
		return errors.New("signature must start with sha256=")
	}
	sigHex := sig[len(signaturePrefix):]
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(sigHex)) {
		return errors.New("signature mismatch")
	}
	return nil
}
