package validator

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/example/messages-relay/internal/config"
	"github.com/example/messages-relay/internal/security"
)

const maxPayloadLen = 1000

// E.164: optional +, then 1-15 digits
var e164Re = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)

// Message is the parsed and validated message structure.
type Message struct {
	Destination string
	Payload     string
	Timestamp   int64
	Nonce       string
	Hash        string
}

// Validator validates incoming MQTT messages.
type Validator struct {
	cfg         *config.Config
	nonceStore  *security.NonceStore
	rateLimiter *security.RateLimiter
}

// New creates a new Validator.
func New(cfg *config.Config) *Validator {
	return &Validator{
		cfg:         cfg,
		nonceStore:  security.NewNonceStore(2 * time.Minute),
		rateLimiter: security.NewRateLimiter(cfg.Limits.MaxPerMinute, cfg.Limits.MaxPerDay),
	}
}

// Validate parses the raw payload and validates it. Returns nil error and valid Message on success.
func (v *Validator) Validate(raw []byte) (*Message, error) {
	// Size check before parsing
	if len(raw) > 10*1024 { // 10KB max raw
		return nil, fmt.Errorf("payload too large")
	}

	var incoming struct {
		Destination string `json:"destination"`
		Payload     string `json:"payload"`
		Timestamp   int64  `json:"timestamp"`
		Nonce       string `json:"nonce"`
		Hash        string `json:"hash"`
	}

	if err := json.Unmarshal(raw, &incoming); err != nil {
		return nil, fmt.Errorf("invalid json: %w", err)
	}

	if incoming.Destination == "" {
		return nil, fmt.Errorf("missing destination")
	}
	if incoming.Payload == "" {
		return nil, fmt.Errorf("missing payload")
	}
	if incoming.Nonce == "" {
		return nil, fmt.Errorf("missing nonce")
	}
	if incoming.Hash == "" {
		return nil, fmt.Errorf("missing hash")
	}

	if len(incoming.Payload) > maxPayloadLen {
		return nil, fmt.Errorf("payload exceeds max length (%d)", maxPayloadLen)
	}

	if !e164Re.MatchString(incoming.Destination) {
		return nil, fmt.Errorf("invalid destination format (must be E.164)")
	}

	// Allowed destinations whitelist (if configured)
	if len(v.cfg.Security.AllowedDestinations) > 0 {
		allowed := false
		for _, d := range v.cfg.Security.AllowedDestinations {
			if d == incoming.Destination {
				allowed = true
				break
			}
		}
		if !allowed {
			return nil, fmt.Errorf("destination not in allowed list")
		}
	}

	// Timestamp window
	skew := time.Duration(v.cfg.Security.MaxSkewSeconds) * time.Second
	now := time.Now().Unix()
	if d := now - incoming.Timestamp; d < 0 {
		d = -d
		if int64(d) > int64(skew.Seconds()) {
			return nil, fmt.Errorf("timestamp too far in future")
		}
	} else if d > int64(skew.Seconds()) {
		return nil, fmt.Errorf("timestamp expired")
	}

	// Nonce replay check
	if v.nonceStore.Seen(incoming.Nonce) {
		return nil, fmt.Errorf("nonce already used (replay)")
	}

	// HMAC verification
	canonical := security.CanonicalInput(
		incoming.Destination,
		incoming.Payload,
		incoming.Timestamp,
		incoming.Nonce,
	)
	if err := security.VerifyHMAC(v.cfg.Security.SharedSecret, canonical, incoming.Hash); err != nil {
		return nil, fmt.Errorf("hmac verification failed: %w", err)
	}

	// Rate limit
	if !v.rateLimiter.Allow() {
		return nil, fmt.Errorf("rate limit exceeded")
	}

	return &Message{
		Destination: incoming.Destination,
		Payload:     incoming.Payload,
		Timestamp:   incoming.Timestamp,
		Nonce:       incoming.Nonce,
		Hash:        incoming.Hash,
	}, nil
}

