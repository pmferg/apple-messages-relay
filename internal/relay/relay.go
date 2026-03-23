package relay

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/example/messages-relay/internal/config"
)

// Relay sends validated messages via macOS Messages.app.
type Relay struct {
	cfg *config.Config
}

// New creates a new Relay.
func New(cfg *config.Config) *Relay {
	return &Relay{cfg: cfg}
}

// Send invokes the AppleScript to send the message. In test mode, does nothing.
func (r *Relay) Send(destination, payload string) error {
	if r.cfg.Relay.TestMode {
		// Test mode: no AppleScript, just validate path would work
		return nil
	}

	scriptPath := r.cfg.Relay.AppleScriptPath
	if scriptPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cannot determine home directory: %w", err)
		}
		scriptPath = filepath.Join(home, "Library", "Application Support", "messages-relay", "send-message.applescript")
	}

	cmd := exec.Command("osascript", scriptPath, destination, payload)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("osascript failed: %w (output: %s)", err, string(output))
	}
	return nil
}
