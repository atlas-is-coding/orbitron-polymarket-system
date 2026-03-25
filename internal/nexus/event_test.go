package nexus

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEventBusSubscribeAndPublish tests basic subscribe and publish functionality
func TestEventBusSubscribeAndPublish(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := zerolog.New(zerolog.NewTestWriter(t))
	bus := NewEventBus(ctx, log)
	defer bus.Close()

	// Subscribe to order.placed
	ch := bus.Subscribe("order.placed")
	require.NotNil(t, ch)

	// Create and publish event
	event := Event{
		ID:        "test-1",
		Type:      EventOrderPlaced,
		Timestamp: time.Now(),
		Source:    "test",
		Payload: OrderPlacedPayload{
			OrderID: "order-1",
		},
	}

	bus.Publish(event)

	// Receive event
	select {
	case received := <-ch:
		assert.Equal(t, event.ID, received.ID)
		assert.Equal(t, EventOrderPlaced, received.Type)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for event")
	}

	stats := bus.Stats()
	assert.Equal(t, uint64(1), stats["sent"])
	assert.Equal(t, uint64(0), stats["dropped"])
}

// TestEventBusGlobPatternMatching tests pattern matching with glob-like patterns
func TestEventBusGlobPatternMatching(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := zerolog.New(zerolog.NewTestWriter(t))
	bus := NewEventBus(ctx, log)
	defer bus.Close()

	// Subscribe to order.* pattern
	ch := bus.Subscribe("order.*")
	require.NotNil(t, ch)

	// Publish order.placed
	bus.Publish(Event{
		ID:        "test-1",
		Type:      EventOrderPlaced,
		Timestamp: time.Now(),
		Source:    "test",
	})

	// Publish order.filled
	bus.Publish(Event{
		ID:        "test-2",
		Type:      EventOrderFilled,
		Timestamp: time.Now(),
		Source:    "test",
	})

	// Publish order.canceled
	bus.Publish(Event{
		ID:        "test-3",
		Type:      EventOrderCanceled,
		Timestamp: time.Now(),
		Source:    "test",
	})

	// All three should be received
	received := make([]string, 0, 3)
	timeout := time.After(1 * time.Second)

	for len(received) < 3 {
		select {
		case event := <-ch:
			received = append(received, event.ID)
		case <-timeout:
			t.Fatalf("timeout waiting for events, got %d events", len(received))
		}
	}

	assert.ElementsMatch(t, []string{"test-1", "test-2", "test-3"}, received)

	// Now publish a non-matching event: wallet.added
	bus.Publish(Event{
		ID:        "test-4",
		Type:      EventWalletAdded,
		Timestamp: time.Now(),
		Source:    "test",
	})

	// Should NOT receive wallet.added
	select {
	case event := <-ch:
		t.Fatalf("should not receive wallet event, got %s", event.ID)
	case <-time.After(100 * time.Millisecond):
		// Expected - no event received
	}

	stats := bus.Stats()
	assert.Equal(t, uint64(4), stats["sent"])
	assert.Equal(t, uint64(0), stats["dropped"]) // no drops
}

// TestEventBusWildcardPattern tests wildcard "*" pattern matching
func TestEventBusWildcardPattern(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := zerolog.New(zerolog.NewTestWriter(t))
	bus := NewEventBus(ctx, log)
	defer bus.Close()

	// Subscribe to "*" (wildcard - all events)
	ch := bus.Subscribe("*")
	require.NotNil(t, ch)

	// Publish various event types
	events := []Event{
		{ID: "evt-1", Type: EventOrderPlaced, Timestamp: time.Now(), Source: "test"},
		{ID: "evt-2", Type: EventWalletAdded, Timestamp: time.Now(), Source: "test"},
		{ID: "evt-3", Type: EventMarketsUpdated, Timestamp: time.Now(), Source: "test"},
		{ID: "evt-4", Type: EventBalanceUpdated, Timestamp: time.Now(), Source: "test"},
	}

	for _, event := range events {
		bus.Publish(event)
	}

	// All should be received
	received := make(map[string]bool)
	timeout := time.After(1 * time.Second)

	for len(received) < 4 {
		select {
		case event := <-ch:
			received[event.ID] = true
		case <-timeout:
			t.Fatalf("timeout waiting for events, got %d", len(received))
		}
	}

	for _, event := range events {
		assert.True(t, received[event.ID], "missing event %s", event.ID)
	}
}

// TestEventBusMultipleSubscribers tests multiple subscribers to same pattern
func TestEventBusMultipleSubscribers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := zerolog.New(zerolog.NewTestWriter(t))
	bus := NewEventBus(ctx, log)
	defer bus.Close()

	// Create 2 subscribers to same pattern
	ch1 := bus.Subscribe("order.*")
	ch2 := bus.Subscribe("order.*")

	require.NotNil(t, ch1)
	require.NotNil(t, ch2)

	// Publish event
	event := Event{
		ID:        "multi-1",
		Type:      EventOrderPlaced,
		Timestamp: time.Now(),
		Source:    "test",
	}

	bus.Publish(event)

	// Both subscribers should receive independently
	timeout := time.After(1 * time.Second)

	// Receive on ch1
	var received1 Event
	select {
	case received1 = <-ch1:
		assert.Equal(t, event.ID, received1.ID)
	case <-timeout:
		t.Fatal("timeout on ch1")
	}

	// Receive on ch2
	var received2 Event
	select {
	case received2 = <-ch2:
		assert.Equal(t, event.ID, received2.ID)
	case <-timeout:
		t.Fatal("timeout on ch2")
	}

	stats := bus.Stats()
	assert.Equal(t, uint64(2), stats["subscribers"])
	assert.Equal(t, uint64(1), stats["sent"])
}

// TestEventBusBackpressure tests channel backpressure handling
func TestEventBusBackpressure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := zerolog.New(zerolog.NewTestWriter(t))
	bus := NewEventBus(ctx, log)
	defer bus.Close()

	// Subscribe to "*"
	ch := bus.Subscribe("*")
	require.NotNil(t, ch)

	// Fill the channel (1024 elements)
	for i := 0; i < 1024; i++ {
		bus.Publish(Event{
			ID:        "fill-" + string(rune(i)),
			Type:      EventOrderPlaced,
			Timestamp: time.Now(),
			Source:    "test",
		})
	}

	// Verify sent count
	stats := bus.Stats()
	assert.Equal(t, uint64(1024), stats["sent"])
	assert.Equal(t, uint64(0), stats["dropped"])

	// Next publish should drop (channel is full)
	bus.Publish(Event{
		ID:        "drop-1",
		Type:      EventOrderPlaced,
		Timestamp: time.Now(),
		Source:    "test",
	})

	// Check dropped count increased
	stats = bus.Stats()
	assert.Equal(t, uint64(1025), stats["sent"])
	assert.Equal(t, uint64(1), stats["dropped"])

	// Verify dropped count via method
	assert.Equal(t, uint64(1), bus.DroppedCount())

	// Drain one event and publish again
	<-ch
	bus.Publish(Event{
		ID:        "accept-1",
		Type:      EventOrderPlaced,
		Timestamp: time.Now(),
		Source:    "test",
	})

	// Should have room now
	stats = bus.Stats()
	assert.Equal(t, uint64(1026), stats["sent"])
	assert.Equal(t, uint64(1), stats["dropped"]) // still 1 dropped total
}

// TestEventBusUnsubscribe tests unsubscribe functionality
func TestEventBusUnsubscribe(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := zerolog.New(zerolog.NewTestWriter(t))
	bus := NewEventBus(ctx, log)
	defer bus.Close()

	// Subscribe
	ch := bus.Subscribe("order.*")
	require.NotNil(t, ch)

	// Verify subscriber count
	stats := bus.Stats()
	assert.Equal(t, uint64(1), stats["subscribers"])

	// Unsubscribe
	bus.Unsubscribe("order.*", ch)

	// Verify subscriber count decreased
	stats = bus.Stats()
	assert.Equal(t, uint64(0), stats["subscribers"])

	// Publish event
	bus.Publish(Event{
		ID:        "unsub-1",
		Type:      EventOrderPlaced,
		Timestamp: time.Now(),
		Source:    "test",
	})

	// Channel should be closed (closed channels always receive zero value with ok=false)
	// Verify the channel is closed
	select {
	case event, ok := <-ch:
		if ok {
			t.Fatal("should not receive on unsubscribed channel")
		}
		// Expected - channel is closed (ok=false)
		assert.Equal(t, Event{}, event)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("receive on closed channel should not block")
	}
}

// TestEventBusClose tests proper shutdown of EventBus
func TestEventBusClose(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := zerolog.New(zerolog.NewTestWriter(t))
	bus := NewEventBus(ctx, log)

	// Create multiple subscribers
	ch1 := bus.Subscribe("order.*")
	ch2 := bus.Subscribe("*")

	// Verify subscribers are registered
	stats := bus.Stats()
	assert.Equal(t, uint64(2), stats["subscribers"])

	// Close bus
	bus.Close()

	// All channels should be closed (receiving on closed channel returns zero value)
	// Send on closed channel should panic, but we're testing receive behavior

	// Verify no more subscribers
	stats = bus.Stats()
	assert.Equal(t, uint64(0), stats["subscribers"])

	// Verify channels are closed by attempting to receive
	var zero Event
	received1, ok1 := <-ch1
	received2, ok2 := <-ch2
	assert.Equal(t, zero, received1)
	assert.Equal(t, zero, received2)
	assert.False(t, ok1)
	assert.False(t, ok2)
}

// TestEventBusPatternEdgeCases tests edge cases in pattern matching
func TestEventBusPatternEdgeCases(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := zerolog.New(zerolog.NewTestWriter(t))
	bus := NewEventBus(ctx, log)
	defer bus.Close()

	tests := []struct {
		name       string
		pattern    string
		eventType  EventType
		shouldMatch bool
	}{
		{"exact match", "order.placed", EventOrderPlaced, true},
		{"exact no match", "order.placed", EventOrderFilled, false},
		{"prefix wildcard", "order.*", EventOrderPlaced, true},
		{"prefix wildcard match filled", "order.*", EventOrderFilled, true},
		{"prefix no match", "order.*", EventWalletAdded, false},
		{"full wildcard", "*", EventOrderPlaced, true},
		{"full wildcard wallet", "*", EventWalletAdded, true},
		{"balance prefix", "balance.*", EventBalanceUpdated, true},
		{"balance no match", "balance.*", EventOrderPlaced, false},
		{"position wildcard", "position.*", EventPositionOpened, true},
		{"position wildcard closed", "position.*", EventPositionClosed, true},
		{"price alert", "price.alert", EventPriceAlert, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := bus.Subscribe(tt.pattern)
			defer bus.Unsubscribe(tt.pattern, ch)

			bus.Publish(Event{
				ID:        "test-" + tt.name,
				Type:      tt.eventType,
				Timestamp: time.Now(),
				Source:    "test",
			})

			timeout := time.After(100 * time.Millisecond)
			received := false

			select {
			case <-ch:
				received = true
			case <-timeout:
				received = false
			}

			assert.Equal(t, tt.shouldMatch, received,
				"pattern %q should %s match %q",
				tt.pattern,
				map[bool]string{true: "not ", false: ""}[!tt.shouldMatch],
				tt.eventType)
		})
	}
}

// TestEventBusConcurrentPublish tests concurrent publishes
func TestEventBusConcurrentPublish(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := zerolog.New(zerolog.NewTestWriter(t))
	bus := NewEventBus(ctx, log)
	defer bus.Close()

	// Subscribe to drain events to prevent backpressure
	drainCh := bus.Subscribe("*")

	var wg sync.WaitGroup
	numGoroutines := 10
	eventsPerGoroutine := 100

	// Drain events in background
	go func() {
		for range drainCh {
			// Just drain
		}
	}()

	// Publish from multiple goroutines
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < eventsPerGoroutine; j++ {
				bus.Publish(Event{
					ID:        "concurrent-" + string(rune(id)) + "-" + string(rune(j)),
					Type:      EventOrderPlaced,
					Timestamp: time.Now(),
					Source:    "test",
				})
			}
		}(i)
	}

	wg.Wait()

	// Verify stats
	stats := bus.Stats()
	expectedCount := uint64(numGoroutines * eventsPerGoroutine)
	assert.Equal(t, expectedCount, stats["sent"])
	assert.LessOrEqual(t, stats["dropped"], expectedCount) // May have some drops due to buffering
}

// TestEventBusConcurrentSubscribe tests concurrent subscribe/unsubscribe
func TestEventBusConcurrentSubscribe(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := zerolog.New(zerolog.NewTestWriter(t))
	bus := NewEventBus(ctx, log)
	defer bus.Close()

	var wg sync.WaitGroup
	numSubscribers := 10

	// Subscribe from multiple goroutines
	channels := make([]<-chan Event, 0, numSubscribers)
	mu := &sync.Mutex{}

	for i := 0; i < numSubscribers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ch := bus.Subscribe("*")
			mu.Lock()
			channels = append(channels, ch)
			mu.Unlock()
		}()
	}

	wg.Wait()

	// All subscribers should be registered
	stats := bus.Stats()
	assert.Equal(t, uint64(numSubscribers), stats["subscribers"])

	// Publish an event
	bus.Publish(Event{
		ID:        "concurrent-sub-test",
		Type:      EventOrderPlaced,
		Timestamp: time.Now(),
		Source:    "test",
	})

	// All should receive
	timeout := time.After(1 * time.Second)
	received := 0

	for received < numSubscribers {
		select {
		case <-channels[received]:
			received++
		case <-timeout:
			t.Fatalf("timeout waiting for subscribers, got %d/%d", received, numSubscribers)
		}
	}

	// Now unsubscribe all
	for _, ch := range channels {
		bus.Unsubscribe("*", ch)
	}

	stats = bus.Stats()
	assert.Equal(t, uint64(0), stats["subscribers"])
}
