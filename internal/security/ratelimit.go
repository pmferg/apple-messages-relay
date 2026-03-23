package security

import (
	"sync"
	"time"
)

// RateLimiter enforces per-minute and per-day message limits.
type RateLimiter struct {
	mu           sync.Mutex
	perMinute    int
	perDay       int
	minuteCount  int
	minuteWindow time.Time
	dayCount     int
	dayWindow    time.Time
}

// NewRateLimiter creates a rate limiter with the given limits.
func NewRateLimiter(maxPerMinute, maxPerDay int) *RateLimiter {
	now := time.Now()
	return &RateLimiter{
		perMinute:   maxPerMinute,
		perDay:      maxPerDay,
		minuteCount: 0,
		minuteWindow: now.Truncate(time.Minute),
		dayCount:   0,
		dayWindow:  now.Truncate(24 * time.Hour),
	}
}

// Allow checks if a message is allowed. Returns false if rate limit exceeded.
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Reset minute window if we've crossed into a new minute
	if now.Sub(rl.minuteWindow) >= time.Minute {
		rl.minuteWindow = now.Truncate(time.Minute)
		rl.minuteCount = 0
	}

	// Reset day window if we've crossed into a new day
	if now.Sub(rl.dayWindow) >= 24*time.Hour {
		rl.dayWindow = now.Truncate(24 * time.Hour)
		rl.dayCount = 0
	}

	if rl.minuteCount >= rl.perMinute {
		return false
	}
	if rl.dayCount >= rl.perDay {
		return false
	}

	rl.minuteCount++
	rl.dayCount++
	return true
}
