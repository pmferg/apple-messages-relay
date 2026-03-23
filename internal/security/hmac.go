package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
)

// CanonicalInput builds the canonical string for HMAC verification.
// Format: destination + "\n" + payload + "\n" + timestamp + "\n" + nonce
func CanonicalInput(destination, payload string, timestamp int64, nonce string) string {
	return destination + "\n" + payload + "\n" + strconv.FormatInt(timestamp, 10) + "\n" + nonce
}

// ComputeHMAC computes HMAC-SHA256 of the canonical input using the secret.
func ComputeHMAC(secret string, canonical string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(canonical))
	return hex.EncodeToString(h.Sum(nil))
}

// VerifyHMAC verifies the provided hash against the computed HMAC using constant-time comparison.
func VerifyHMAC(secret, canonical, providedHash string) error {
	expected := ComputeHMAC(secret, canonical)
	if !constantTimeCompare(expected, providedHash) {
		return fmt.Errorf("hmac verification failed")
	}
	return nil
}

// constantTimeCompare performs a constant-time comparison to prevent timing attacks.
func constantTimeCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}
