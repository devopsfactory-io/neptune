package verify

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func hmacSHA256Hex(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func TestWebhookSignature_Valid(t *testing.T) {
	body := []byte(`{"foo":1}`)
	secret := "mysecret"
	sig := "sha256=" + hmacSHA256Hex(body, secret)
	if err := WebhookSignature(body, sig, secret); err != nil {
		t.Errorf("expected nil error for valid signature, got %v", err)
	}
}

func TestWebhookSignature_MissingHeader(t *testing.T) {
	body := []byte(`{"foo":1}`)
	secret := "mysecret"
	err := WebhookSignature(body, "", secret)
	if err == nil {
		t.Fatal("expected error for missing header")
	}
	if err.Error() != "missing X-Hub-Signature-256" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestWebhookSignature_EmptySecret(t *testing.T) {
	body := []byte(`{"foo":1}`)
	sig := "sha256=" + hmacSHA256Hex(body, "x")
	err := WebhookSignature(body, sig, "")
	if err == nil {
		t.Fatal("expected error for empty secret")
	}
	if err.Error() != "webhook secret is empty" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestWebhookSignature_InvalidPrefix(t *testing.T) {
	body := []byte(`{"foo":1}`)
	secret := "mysecret"
	err := WebhookSignature(body, "invalid=abc123def456", secret)
	if err == nil {
		t.Fatal("expected error for invalid prefix")
	}
	if err.Error() != "signature must start with sha256=" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestWebhookSignature_Mismatch(t *testing.T) {
	body := []byte(`{"foo":1}`)
	secret := "mysecret"
	wrongSig := "sha256=deadbeef"
	err := WebhookSignature(body, wrongSig, secret)
	if err == nil {
		t.Fatal("expected error for signature mismatch")
	}
	if err.Error() != "signature mismatch" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestWebhookSignature_WrongSecret(t *testing.T) {
	body := []byte(`{"foo":1}`)
	sig := "sha256=" + hmacSHA256Hex(body, "secret1")
	err := WebhookSignature(body, sig, "secret2")
	if err == nil {
		t.Fatal("expected error when secret does not match")
	}
	if err.Error() != "signature mismatch" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestWebhookSignature_HeaderTrimSpace(t *testing.T) {
	body := []byte(`{"foo":1}`)
	secret := "mysecret"
	sig := "  sha256=" + hmacSHA256Hex(body, secret) + "  "
	if err := WebhookSignature(body, sig, secret); err != nil {
		t.Errorf("expected nil for header with trim; got %v", err)
	}
}
