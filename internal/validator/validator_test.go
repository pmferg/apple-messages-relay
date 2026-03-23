package validator

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/example/messages-relay/internal/config"
	"github.com/example/messages-relay/internal/security"
)

func TestValidator_ValidMessage(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			SharedSecret:   "test-secret",
			MaxSkewSeconds: 60,
		},
		Limits: config.LimitsConfig{
			MaxPerMinute: 10,
			MaxPerDay:    100,
		},
	}
	v := New(cfg)

	ts := time.Now().Unix()
	nonce := "test-nonce-12345"
	dest := "+447700900123"
	payload := "Hello"
	canonical := security.CanonicalInput(dest, payload, ts, nonce)
	hash := security.ComputeHMAC(cfg.Security.SharedSecret, canonical)

	raw := mustMarshal(map[string]interface{}{
		"destination": dest,
		"payload":     payload,
		"timestamp":   ts,
		"nonce":       nonce,
		"hash":        hash,
	})

	msg, err := v.Validate(raw)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if msg.Destination != dest {
		t.Errorf("Destination = %q, want %q", msg.Destination, dest)
	}
	if msg.Payload != payload {
		t.Errorf("Payload = %q, want %q", msg.Payload, payload)
	}
}

func TestValidator_InvalidHMAC(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			SharedSecret:   "test-secret",
			MaxSkewSeconds: 60,
		},
		Limits: config.LimitsConfig{MaxPerMinute: 10, MaxPerDay: 100},
	}
	v := New(cfg)

	ts := time.Now().Unix()
	raw := mustMarshal(map[string]interface{}{
		"destination": "+447700900123",
		"payload":     "Hello",
		"timestamp":   ts,
		"nonce":       "test-nonce",
		"hash":        "wrong-hash",
	})

	_, err := v.Validate(raw)
	if err == nil {
		t.Error("expected HMAC validation error")
	}
}

func TestValidator_ExpiredTimestamp(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			SharedSecret:   "test-secret",
			MaxSkewSeconds: 60,
		},
		Limits: config.LimitsConfig{MaxPerMinute: 10, MaxPerDay: 100},
	}
	v := New(cfg)

	ts := time.Now().Unix() - 120 // 2 minutes ago
	dest := "+447700900123"
	payload := "Hello"
	nonce := "expired-nonce"
	canonical := security.CanonicalInput(dest, payload, ts, nonce)
	hash := security.ComputeHMAC(cfg.Security.SharedSecret, canonical)

	raw := mustMarshal(map[string]interface{}{
		"destination": dest,
		"payload":     payload,
		"timestamp":   ts,
		"nonce":       nonce,
		"hash":        hash,
	})

	_, err := v.Validate(raw)
	if err == nil {
		t.Error("expected timestamp expired error")
	}
}

func TestValidator_NonceReplay(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			SharedSecret:   "test-secret",
			MaxSkewSeconds: 60,
		},
		Limits: config.LimitsConfig{MaxPerMinute: 10, MaxPerDay: 100},
	}
	v := New(cfg)

	ts := time.Now().Unix()
	dest := "+447700900123"
	payload := "Hello"
	nonce := "replay-nonce"
	canonical := security.CanonicalInput(dest, payload, ts, nonce)
	hash := security.ComputeHMAC(cfg.Security.SharedSecret, canonical)

	raw := mustMarshal(map[string]interface{}{
		"destination": dest,
		"payload":     payload,
		"timestamp":   ts,
		"nonce":       nonce,
		"hash":        hash,
	})

	_, err := v.Validate(raw)
	if err != nil {
		t.Fatalf("first validation: %v", err)
	}
	_, err = v.Validate(raw)
	if err == nil {
		t.Error("expected replay error on second validation")
	}
}

func TestValidator_InvalidDestination(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			SharedSecret:   "test-secret",
			MaxSkewSeconds: 60,
		},
		Limits: config.LimitsConfig{MaxPerMinute: 10, MaxPerDay: 100},
	}
	v := New(cfg)

	ts := time.Now().Unix()
	dest := "not-e164"
	payload := "Hello"
	nonce := "dest-nonce"
	canonical := security.CanonicalInput(dest, payload, ts, nonce)
	hash := security.ComputeHMAC(cfg.Security.SharedSecret, canonical)

	raw := mustMarshal(map[string]interface{}{
		"destination": dest,
		"payload":     payload,
		"timestamp":   ts,
		"nonce":       nonce,
		"hash":        hash,
	})

	_, err := v.Validate(raw)
	if err == nil {
		t.Error("expected invalid destination error")
	}
}

func TestValidator_AllowedDestinations(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			SharedSecret:         "test-secret",
			MaxSkewSeconds:       60,
			AllowedDestinations:  []string{"+447700900123"},
		},
		Limits: config.LimitsConfig{MaxPerMinute: 10, MaxPerDay: 100},
	}
	v := New(cfg)

	ts := time.Now().Unix()
	dest := "+447700900999" // not in allowed list
	payload := "Hello"
	nonce := "allowed-nonce"
	canonical := security.CanonicalInput(dest, payload, ts, nonce)
	hash := security.ComputeHMAC(cfg.Security.SharedSecret, canonical)

	raw := mustMarshal(map[string]interface{}{
		"destination": dest,
		"payload":     payload,
		"timestamp":   ts,
		"nonce":       nonce,
		"hash":        hash,
	})

	_, err := v.Validate(raw)
	if err == nil {
		t.Error("expected destination not allowed error")
	}
}

func mustMarshal(v map[string]interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
