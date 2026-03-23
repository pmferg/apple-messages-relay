package relay

import (
	"testing"

	"github.com/example/messages-relay/internal/config"
)

func TestRelay_TestMode(t *testing.T) {
	cfg := &config.Config{
		Relay: config.RelayConfig{
			TestMode: true,
		},
	}
	r := New(cfg)
	err := r.Send("+447700900123", "Hello")
	if err != nil {
		t.Errorf("Send in test mode: %v", err)
	}
}
