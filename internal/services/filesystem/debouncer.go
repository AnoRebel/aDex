package filesystem

import (
	"context"
	"sync"
	"time"

	"aDex-UI/internal/models"
)

// Debouncer prevents excessive event processing by collecting events
// and emitting them after a specified delay
type Debouncer struct {
	delay     time.Duration
	eventChan chan *models.FileSystemEvent
	outputChan chan *models.FileSystemEvent
	mutex      sync.RWMutex
	pending   map[string]*PendingEvent
	ctx       context.Context
	cancel    context.CancelFunc
	running   bool
}

// PendingEvent represents an event that is pending debouncing
type PendingEvent struct {
	Event      *models.FileSystemEvent
	LastUpdate time.Time
	Timer      *time.Timer
	Count      int
}

// DebounceConfig holds configuration for debouncing
type DebounceConfig struct {
	Delay           time.Duration `json:"delay"`
	MaxPending      int          `json:"max_pending"`
	BatchSize       int          `json:"batch_size"`
	GroupByPath     bool         `json:"group_by_path"`
	GroupByType     bool         `json:"group_by_type"`
}

// DefaultDebounceConfig returns default debouncing configuration
func DefaultDebounceConfig() *DebounceConfig {
	return &DebounceConfig{
		Delay:       300 * time.Millisecond,
		MaxPending:  1000,
		BatchSize:   50,
		GroupByPath: true,
		GroupByType: false,
	}
}

// NewDebouncer creates a new event debouncer
func NewDebouncer(config *DebounceConfig) *Debouncer {
	if config == nil {
		config = DefaultDebounceConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	debouncer := &Debouncer{
		delay:     config.Delay,
		eventChan: make(chan *models.FileSystemEvent, config.MaxPending),
		outputChan: make(chan *models.FileSystemEvent, config.MaxPending),
		pending:   make(map[string]*PendingEvent),
		ctx:       ctx,
		cancel:    cancel,
		running:   true,
	}

	// Start the debouncing goroutine
	go debouncer.processPendingEvents(config)

	return debouncer
}

// processPendingEvents handles the debouncing logic
func (d *Debouncer) processPendingEvents(config *DebounceConfig) {
	for {
		select {
		case <-d.ctx.Done():
			return
		case event := <-d.eventChan:
			d.handleEvent(event, config)
		case <-time.After(config.Delay):
			d.flushAllPending(config)
		}
	}
}

// handleEvent processes a single incoming event
func (d *Debouncer) handleEvent(event *models.FileSystemEvent, config *DebounceConfig) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	// Check for nil event
	if event == nil {
		return
	}

	// Create key for grouping events
	key := d.createEventKey(event, config)

	// Check if we already have a pending event for this key
	if pending, exists := d.pending[key]; exists {
		// Update the existing pending event
		pending.Event = event
		pending.LastUpdate = time.Now()
		pending.Count++

		// Reset the timer
		if pending.Timer != nil {
			pending.Timer.Stop()
		}

		// Set a new timer to emit the event after the delay
		pending.Timer = time.AfterFunc(config.Delay, func() {
			d.emitPendingEvent(key)
		})
	} else {
		// Create new pending event
		timer := time.AfterFunc(config.Delay, func() {
			d.emitPendingEvent(key)
		})

		d.pending[key] = &PendingEvent{
			Event:      event,
			LastUpdate: time.Now(),
			Timer:      timer,
			Count:      1,
		}
	}
}

// createEventKey creates a key for grouping events
func (d *Debouncer) createEventKey(event *models.FileSystemEvent, config *DebounceConfig) string {
	var key string

	if config.GroupByPath {
		key = event.Path
	}

	if config.GroupByType {
		if key != "" {
			key += "|"
		}
		key += event.Operation
	}

	if key == "" {
		// Fallback to using path if no grouping
		key = event.Path
	}

	return key
}

// emitPendingEvent emits a pending event
func (d *Debouncer) emitPendingEvent(key string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if pending, exists := d.pending[key]; exists {
		// Stop the timer if it's still running
		if pending.Timer != nil {
			pending.Timer.Stop()
		}

		// Send the event to output channel
		select {
		case d.outputChan <- pending.Event:
		case <-d.ctx.Done():
			// Debouncer is shutting down
		}

		// Remove from pending map
		delete(d.pending, key)
	}
}

// flushAllPending emits all pending events immediately
func (d *Debouncer) flushAllPending(config *DebounceConfig) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if len(d.pending) == 0 {
		return
	}

	// Collect all pending events
	events := make([]*models.FileSystemEvent, 0, len(d.pending))
	for _, pending := range d.pending {
		// Stop the timer if it's still running
		if pending.Timer != nil {
			pending.Timer.Stop()
		}
		events = append(events, pending.Event)
	}

	// Clear pending map
	d.pending = make(map[string]*PendingEvent)

	// Emit events in batches
	go func() {
		for i := 0; i < len(events); i += config.BatchSize {
			end := i + config.BatchSize
			if end > len(events) {
				end = len(events)
			}

			batch := events[i:end]
			for _, event := range batch {
				select {
				case d.outputChan <- event:
				case <-d.ctx.Done():
					return
				}
			}

			// Small delay between batches
			if i+config.BatchSize < len(events) {
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()
}

// AddEvent adds an event to the debouncer
func (d *Debouncer) AddEvent(event *models.FileSystemEvent) {
	if !d.running {
		return
	}

	select {
	case d.eventChan <- event:
	case <-d.ctx.Done():
		// Debouncer is shutting down
	}
}

// Events returns the output channel for debounced events
func (d *Debouncer) Events() <-chan *models.FileSystemEvent {
	return d.outputChan
}

// GetStats returns debouncing statistics
func (d *Debouncer) GetStats() map[string]interface{} {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	stats := map[string]interface{}{
		"running":        d.running,
		"pending_count":  len(d.pending),
		"delay":          d.delay.String(),
	}

	// Add details about pending events
	pendingDetails := make([]map[string]interface{}, 0, len(d.pending))
	for key, pending := range d.pending {
		pendingDetails = append(pendingDetails, map[string]interface{}{
			"key":         key,
			"path":        pending.Event.Path,
			"operation":   pending.Event.Operation,
			"count":       pending.Count,
			"last_update": pending.LastUpdate,
			"age":         time.Since(pending.LastUpdate).String(),
		})
	}
	stats["pending_events"] = pendingDetails

	return stats
}

// Stop stops the debouncer and flushes all pending events
func (d *Debouncer) Stop() {
	if !d.running {
		return
	}

	d.running = false
	d.cancel()

	// Flush any remaining pending events
	d.flushAllPending(DefaultDebounceConfig())

	// Close channels
	close(d.eventChan)
	close(d.outputChan)
}

// SetDelay updates the debounce delay
func (d *Debouncer) SetDelay(delay time.Duration) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.delay = delay
}

// Flush manually flushes all pending events
func (d *Debouncer) Flush() {
	d.flushAllPending(DefaultDebounceConfig())
}

// GetPendingCount returns the number of pending events
func (d *Debouncer) GetPendingCount() int {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	return len(d.pending)
}

// ClearPending clears all pending events without emitting them
func (d *Debouncer) ClearPending() {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	// Stop all timers
	for _, pending := range d.pending {
		if pending.Timer != nil {
			pending.Timer.Stop()
		}
	}

	// Clear pending map
	d.pending = make(map[string]*PendingEvent)
}

// EventGroup represents a group of related events that were debounced
type EventGroup struct {
	Path        string                    `json:"path"`
	Operation   string                    `json:"operation"`
	Count       int                       `json:"count"`
	FirstEvent  *models.FileSystemEvent   `json:"first_event"`
	LastEvent   *models.FileSystemEvent   `json:"last_event"`
	Duration    time.Duration             `json:"duration"`
}

// GetEventGroups returns information about grouped events
func (d *Debouncer) GetEventGroups() []EventGroup {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	groups := make([]EventGroup, 0, len(d.pending))
	for _, pending := range d.pending {
		group := EventGroup{
			Path:       pending.Event.Path,
			Operation:  pending.Event.Operation,
			Count:      pending.Count,
			FirstEvent: pending.Event, // In a real implementation, we'd track the first event
			LastEvent:  pending.Event,
			Duration:   time.Since(pending.LastUpdate),
		}
		groups = append(groups, group)
	}

	return groups
}