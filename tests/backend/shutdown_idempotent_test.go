package tests

import (
	"context"
	"testing"
	"time"

	"aDex-UI/backend/services/coordinator"
)

// Verifies ServiceShutdown is idempotent and does not deadlock when called
// repeatedly (spec: application-lifecycle — "Cleanup is requested more than once").
func TestShutdownIdempotentNoDeadlock(t *testing.T) {
	sc := coordinator.NewServiceCoordinator()
	if err := sc.Initialize(context.Background()); err != nil {
		t.Fatalf("initialize: %v", err)
	}
	done := make(chan struct{})
	go func() {
		for i := 0; i < 3; i++ {
			if err := sc.ServiceShutdown(); err != nil {
				t.Errorf("shutdown %d returned error: %v", i, err)
			}
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(30 * time.Second):
		t.Fatal("ServiceShutdown deadlocked (3 sequential calls did not complete)")
	}
}
