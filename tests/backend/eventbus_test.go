package tests

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"aDex-UI/internal/events"
)

func TestEventBus_PublishDelivers(t *testing.T) {
	bus := events.NewEventBus()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var got int32
	sub, err := bus.Subscribe(ctx, []string{"x.fired"}, func(_ context.Context, ev events.Event) error {
		atomic.AddInt32(&got, 1)
		return nil
	})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	defer bus.Unsubscribe(sub.ID)

	for i := 0; i < 5; i++ {
		if err := bus.Publish(ctx, "x.fired", i, "test"); err != nil {
			t.Fatalf("Publish: %v", err)
		}
	}

	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) && atomic.LoadInt32(&got) < 5 {
		time.Sleep(5 * time.Millisecond)
	}
	if got := atomic.LoadInt32(&got); got != 5 {
		t.Fatalf("got %d events, want 5", got)
	}
}

func TestEventBus_FiltersByType(t *testing.T) {
	bus := events.NewEventBus()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var got int32
	_, err := bus.Subscribe(ctx, []string{"x.fired"}, func(_ context.Context, _ events.Event) error {
		atomic.AddInt32(&got, 1)
		return nil
	})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	bus.Publish(ctx, "y.other", nil, "test") // wrong type — should not deliver
	bus.Publish(ctx, "x.fired", nil, "test")

	time.Sleep(50 * time.Millisecond)
	if got := atomic.LoadInt32(&got); got != 1 {
		t.Fatalf("got %d events, want 1", got)
	}
}

func TestEventBus_UnsubscribeStopsDelivery(t *testing.T) {
	bus := events.NewEventBus()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var got int32
	sub, _ := bus.Subscribe(ctx, []string{"e"}, func(_ context.Context, _ events.Event) error {
		atomic.AddInt32(&got, 1)
		return nil
	})
	bus.Publish(ctx, "e", nil, "")
	time.Sleep(20 * time.Millisecond)

	if err := bus.Unsubscribe(sub.ID); err != nil {
		t.Fatalf("Unsubscribe: %v", err)
	}
	bus.Publish(ctx, "e", nil, "")
	time.Sleep(20 * time.Millisecond)
	if got := atomic.LoadInt32(&got); got != 1 {
		t.Fatalf("got %d events after unsubscribe, want 1", got)
	}
}

func TestEventBus_ShutdownDrainsGoroutines(t *testing.T) {
	bus := events.NewEventBus()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 0; i < 10; i++ {
		_, err := bus.Subscribe(ctx, []string{"x"}, func(_ context.Context, _ events.Event) error {
			return nil
		})
		if err != nil {
			t.Fatalf("Subscribe: %v", err)
		}
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancelShutdown()
	if err := bus.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}

	// After Shutdown, Publish should error.
	if err := bus.Publish(ctx, "x", nil, ""); err == nil {
		t.Fatalf("Publish after Shutdown should error")
	}
	// Idempotent.
	if err := bus.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("second Shutdown: %v", err)
	}
}

func TestEventBus_HandlesManyEventTypes(t *testing.T) {
	// The previous bus implementation hardcoded channels[0]/channels[1]
	// in eventListener and silently dropped events for subscriptions with
	// 3+ event types. Verify the refactor handles arbitrary counts.
	bus := events.NewEventBus()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	types := []string{"a", "b", "c", "d", "e"}
	var got int32
	_, _ = bus.Subscribe(ctx, types, func(_ context.Context, _ events.Event) error {
		atomic.AddInt32(&got, 1)
		return nil
	})
	for _, tp := range types {
		bus.Publish(ctx, tp, nil, "")
	}
	time.Sleep(50 * time.Millisecond)
	if got := atomic.LoadInt32(&got); got != int32(len(types)) {
		t.Fatalf("got %d events across %d types, want %d", got, len(types), len(types))
	}
}
