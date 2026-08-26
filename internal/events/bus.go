package events

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// EventBus handles event publishing and subscription.
//
// Each subscription owns one buffered channel and one goroutine. Publish
// fans out to every subscription whose Types include the event type. The
// per-subscription goroutine exits when its channel is closed (via
// Unsubscribe or Shutdown) or when the parent context is cancelled.
//
// The previous design held one channel per event-type and a single
// goroutine that hard-coded `channels[0]`/`channels[1]` polling. That
// silently dropped events for subscriptions registered with anything
// other than 1 or 2 types and never let goroutines wind down on shutdown
// — which is what made `go test` hang for 30 s on every test that
// initialized the coordinator.
type EventBus struct {
	mu          sync.RWMutex
	subs        map[string]*subscription // keyed by sub ID
	wg          sync.WaitGroup
	closed      bool
}

type subscription struct {
	id      string
	types   map[string]struct{}
	handler EventHandler
	ch      chan Event
	cancel  context.CancelFunc
}

// Event represents an application event.
type Event struct {
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
	Source    string      `json:"source"`
}

// EventSubscription is the public handle returned to subscribers. The
// Active flag is purely informational — actual cancellation goes through
// EventBus.Unsubscribe(id).
type EventSubscription struct {
	ID     string
	Types  []string
	Active bool
}

// EventHandler handles incoming events.
type EventHandler func(ctx context.Context, event Event) error

// IEventBus is defined in contracts.go.

// Global event bus instance.
var (
	globalEventBus *EventBus
	globalOnce     sync.Once
)

// GetEventBus returns the global event bus instance, initialized lazily.
func GetEventBus() *EventBus {
	globalOnce.Do(func() {
		globalEventBus = NewEventBus()
	})
	return globalEventBus
}

// NewEventBus constructs a fresh, isolated EventBus. Tests should use this
// instead of the global instance to keep state from leaking between cases.
func NewEventBus() *EventBus {
	return &EventBus{
		subs: make(map[string]*subscription),
	}
}

// Subscribe registers a handler for the given event types and starts a
// goroutine that delivers matching events to it.
func (eb *EventBus) Subscribe(ctx context.Context, eventTypes []string, handler EventHandler) (*EventSubscription, error) {
	if handler == nil {
		return nil, fmt.Errorf("events: handler must be non-nil")
	}
	if len(eventTypes) == 0 {
		return nil, fmt.Errorf("events: at least one event type required")
	}

	eb.mu.Lock()
	if eb.closed {
		eb.mu.Unlock()
		return nil, fmt.Errorf("events: bus is shut down")
	}

	subCtx, cancel := context.WithCancel(ctx)
	id := fmt.Sprintf("sub-%d", time.Now().UnixNano())
	types := make(map[string]struct{}, len(eventTypes))
	for _, t := range eventTypes {
		types[t] = struct{}{}
	}

	s := &subscription{
		id:      id,
		types:   types,
		handler: handler,
		ch:      make(chan Event, 100),
		cancel:  cancel,
	}
	eb.subs[id] = s
	eb.wg.Add(1)
	eb.mu.Unlock()

	go eb.deliver(subCtx, s)

	return &EventSubscription{ID: id, Types: append([]string(nil), eventTypes...), Active: true}, nil
}

// SubscribeOnce delivers exactly one matching event then unsubscribes.
func (eb *EventBus) SubscribeOnce(ctx context.Context, eventType string, handler EventHandler) (*EventSubscription, error) {
	var ref *EventSubscription
	wrapped := func(ctx context.Context, ev Event) error {
		err := handler(ctx, ev)
		if ref != nil {
			_ = eb.Unsubscribe(ref.ID)
		}
		return err
	}
	sub, err := eb.Subscribe(ctx, []string{eventType}, wrapped)
	if err != nil {
		return nil, err
	}
	ref = sub
	return sub, nil
}

// deliver is the per-subscription goroutine.
func (eb *EventBus) deliver(ctx context.Context, s *subscription) {
	defer eb.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-s.ch:
			if !ok {
				return
			}
			if _, want := s.types[ev.Type]; !want && ev.Type != "broadcast" {
				continue
			}
			// Don't propagate handler errors — they shouldn't take the
			// goroutine down. We log/swallow at the bus boundary.
			_ = s.handler(ctx, ev)
		}
	}
}

// Publish delivers an event to every subscription that wants its type.
func (eb *EventBus) Publish(_ context.Context, eventType string, data interface{}, source string) error {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	if eb.closed {
		return fmt.Errorf("events: bus is shut down")
	}
	ev := Event{Type: eventType, Timestamp: time.Now(), Data: data, Source: source}
	for _, s := range eb.subs {
		if _, want := s.types[eventType]; !want {
			continue
		}
		select {
		case s.ch <- ev:
		default:
			// Subscriber is slow; drop rather than block the publisher.
		}
	}
	return nil
}

// Broadcast sends a synthetic "broadcast" event to every subscription
// regardless of declared type.
func (eb *EventBus) Broadcast(_ context.Context, data interface{}, source string) error {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	if eb.closed {
		return fmt.Errorf("events: bus is shut down")
	}
	ev := Event{Type: "broadcast", Timestamp: time.Now(), Data: data, Source: source}
	for _, s := range eb.subs {
		select {
		case s.ch <- ev:
		default:
		}
	}
	return nil
}

// Unsubscribe stops delivery to the given subscription, closes its
// channel, and waits for its goroutine to exit.
func (eb *EventBus) Unsubscribe(subscriptionID string) error {
	eb.mu.Lock()
	s, ok := eb.subs[subscriptionID]
	if !ok {
		eb.mu.Unlock()
		return nil
	}
	delete(eb.subs, subscriptionID)
	eb.mu.Unlock()

	s.cancel()
	close(s.ch)
	return nil
}

// GetSubscribers returns the number of subscriptions whose Types include
// the given event type.
func (eb *EventBus) GetSubscribers(eventType string) int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	n := 0
	for _, s := range eb.subs {
		if _, ok := s.types[eventType]; ok {
			n++
		}
	}
	return n
}

// Emit is an alias for Publish that uses a background context. Provided
// for callers that don't have a context handy (e.g. boot-time hooks).
func (eb *EventBus) Emit(eventType string, data interface{}) error {
	return eb.Publish(context.Background(), eventType, data, "system")
}

// Shutdown drains every subscription and waits up to ctx's deadline for
// the per-subscription goroutines to exit. Returns ctx.Err() if the
// deadline fires first.
func (eb *EventBus) Shutdown(ctx context.Context) error {
	eb.mu.Lock()
	if eb.closed {
		eb.mu.Unlock()
		return nil
	}
	eb.closed = true
	subs := make([]*subscription, 0, len(eb.subs))
	for _, s := range eb.subs {
		subs = append(subs, s)
	}
	eb.subs = map[string]*subscription{}
	eb.mu.Unlock()

	for _, s := range subs {
		s.cancel()
		close(s.ch)
	}

	done := make(chan struct{})
	go func() {
		eb.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Common event types and data structs live in types.go to keep bus.go
// focused on the bus mechanics.
