package utils

import (
	"sync"
	"testing"
)

// IDs taken concurrently in a tight loop must never collide. A bare
// timestamp did on Windows, leaking event-bus subscribers.
func TestUniqueIDNeverCollides(t *testing.T) {
	const n = 10000
	ids := make(chan string, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); ids <- UniqueID("x-") }()
	}
	wg.Wait()
	close(ids)
	seen := make(map[string]bool, n)
	for id := range ids {
		if seen[id] {
			t.Fatalf("duplicate id %q", id)
		}
		seen[id] = true
	}
}
