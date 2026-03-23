package security

import (
	"sync"
	"time"
)

// NonceStore maintains a cache of seen nonces to prevent replay attacks.
type NonceStore struct {
	mu     sync.RWMutex
	nonces map[string]time.Time
	ttl    time.Duration
}

// NewNonceStore creates a nonce store with the given TTL.
// Entries expire after TTL and are pruned opportunistically.
func NewNonceStore(ttl time.Duration) *NonceStore {
	ns := &NonceStore{
		nonces: make(map[string]time.Time),
		ttl:    ttl,
	}
	go ns.pruneLoop()
	return ns
}

// Seen checks if the nonce has been seen before. If not, records it and returns false (not a replay).
// Returns true if the nonce was already seen (replay attack).
func (ns *NonceStore) Seen(nonce string) bool {
	if nonce == "" {
		return true // Empty nonce treated as invalid
	}
	ns.mu.Lock()
	defer ns.mu.Unlock()

	now := time.Now()
	if exp, ok := ns.nonces[nonce]; ok {
		if now.Before(exp) {
			return true // Replay
		}
		// Expired, remove and allow (will re-add below)
		delete(ns.nonces, nonce)
	}

	ns.nonces[nonce] = now.Add(ns.ttl)
	return false
}

func (ns *NonceStore) pruneLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		ns.prune()
	}
}

func (ns *NonceStore) prune() {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	now := time.Now()
	for nonce, exp := range ns.nonces {
		if now.After(exp) {
			delete(ns.nonces, nonce)
		}
	}
}
