package security

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestNonceStore_FirstUse(t *testing.T) {
	ns := NewNonceStore(2 * time.Minute)

	if ns.Seen("nonce-1") {
		t.Error("first use of nonce should not be replay")
	}
}

func TestNonceStore_Replay(t *testing.T) {
	ns := NewNonceStore(2 * time.Minute)

	if ns.Seen("nonce-2") {
		t.Error("first use should not be replay")
	}
	if !ns.Seen("nonce-2") {
		t.Error("second use should be replay")
	}
}

func TestNonceStore_EmptyNonce(t *testing.T) {
	ns := NewNonceStore(2 * time.Minute)

	if !ns.Seen("") {
		t.Error("empty nonce should be treated as invalid (replay)")
	}
}

func TestNonceStore_Concurrent(t *testing.T) {
	ns := NewNonceStore(2 * time.Minute)
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			ns.Seen(fmt.Sprintf("nonce-%d", n))
		}(i)
	}
	wg.Wait()
}
