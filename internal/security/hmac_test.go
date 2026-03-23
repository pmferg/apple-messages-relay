package security

import (
	"testing"
)

func TestCanonicalInput(t *testing.T) {
	got := CanonicalInput("+447700900123", "Hello", 1700000000, "abc-uuid")
	want := "+447700900123\nHello\n1700000000\nabc-uuid"
	if got != want {
		t.Errorf("CanonicalInput = %q, want %q", got, want)
	}
}

func TestComputeHMAC(t *testing.T) {
	secret := "my-secret"
	canonical := CanonicalInput("+447700900123", "Hello", 1700000000, "abc-uuid")
	hash := ComputeHMAC(secret, canonical)
	if len(hash) != 64 {
		t.Errorf("HMAC hex length = %d, want 64", len(hash))
	}
	// Deterministic
	hash2 := ComputeHMAC(secret, canonical)
	if hash != hash2 {
		t.Error("HMAC should be deterministic")
	}
}

func TestVerifyHMAC(t *testing.T) {
	secret := "my-secret"
	canonical := CanonicalInput("+447700900123", "Hello", 1700000000, "abc-uuid")
	hash := ComputeHMAC(secret, canonical)

	if err := VerifyHMAC(secret, canonical, hash); err != nil {
		t.Errorf("VerifyHMAC valid: %v", err)
	}
	if err := VerifyHMAC(secret, canonical, hash+"x"); err == nil {
		t.Error("VerifyHMAC invalid: expected error")
	}
	if err := VerifyHMAC(secret, canonical+"x", hash); err == nil {
		t.Error("VerifyHMAC wrong canonical: expected error")
	}
	if err := VerifyHMAC("wrong-secret", canonical, hash); err == nil {
		t.Error("VerifyHMAC wrong secret: expected error")
	}
}

func TestConstantTimeCompare(t *testing.T) {
	if constantTimeCompare("", "x") {
		t.Error("empty vs x should be false")
	}
	if constantTimeCompare("abc", "ab") {
		t.Error("different length should be false")
	}
	if !constantTimeCompare("abc", "abc") {
		t.Error("same string should be true")
	}
	if constantTimeCompare("abc", "abd") {
		t.Error("different char should be false")
	}
}
