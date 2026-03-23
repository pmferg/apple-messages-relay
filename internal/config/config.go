package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the full application configuration.
type Config struct {
	MQTT     MQTTConfig     `json:"mqtt"`
	Security SecurityConfig `json:"security"`
	Limits   LimitsConfig   `json:"limits"`
	Relay    RelayConfig    `json:"relay"`
}

// MQTTConfig holds MQTT broker connection settings.
type MQTTConfig struct {
	Broker   string `json:"broker"`
	Topic    string `json:"topic"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// SecurityConfig holds HMAC and validation settings.
type SecurityConfig struct {
	SharedSecret    string `json:"shared_secret"`
	MaxSkewSeconds  int    `json:"max_skew_seconds"`
	AllowedDestinations []string `json:"allowed_destinations,omitempty"` // Empty = allow all
}

// LimitsConfig holds rate limiting settings.
type LimitsConfig struct {
	MaxPerMinute int `json:"max_per_minute"`
	MaxPerDay    int `json:"max_per_day"`
}

// RelayConfig holds relay-specific settings.
type RelayConfig struct {
	AppleScriptPath string `json:"applescript_path"`
	TestMode        bool   `json:"test_mode"` // If true, don't invoke AppleScript
}

// DefaultConfigPath returns the default config file location.
func DefaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, "Library", "Application Support", "messages-relay", "config.json"), nil
}

// Load reads configuration from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// Validate checks that required fields are present and sane.
func (c *Config) Validate() error {
	if c.MQTT.Broker == "" {
		return fmt.Errorf("mqtt.broker is required")
	}
	if c.MQTT.Topic == "" {
		return fmt.Errorf("mqtt.topic is required")
	}
	if c.Security.SharedSecret == "" {
		return fmt.Errorf("security.shared_secret is required")
	}
	if c.Security.MaxSkewSeconds <= 0 {
		c.Security.MaxSkewSeconds = 60
	}
	if c.Limits.MaxPerMinute <= 0 {
		c.Limits.MaxPerMinute = 5
	}
	if c.Limits.MaxPerDay <= 0 {
		c.Limits.MaxPerDay = 50
	}
	return nil
}
