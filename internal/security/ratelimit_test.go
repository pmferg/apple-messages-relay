package security

import (
	"testing"
)

func TestRateLimiter_AllowsWithinLimit(t *testing.T) {
	rl := NewRateLimiter(5, 50)
	for i := 0; i < 5; i++ {
		if !rl.Allow() {
			t.Errorf("message %d should be allowed", i+1)
		}
	}
	if rl.Allow() {
		t.Error("6th message should be denied (per-minute limit)")
	}
}

func TestRateLimiter_DayLimit(t *testing.T) {
	rl := NewRateLimiter(100, 3)
	for i := 0; i < 3; i++ {
		if !rl.Allow() {
			t.Errorf("message %d should be allowed", i+1)
		}
	}
	if rl.Allow() {
		t.Error("4th message should be denied (per-day limit)")
	}
}
