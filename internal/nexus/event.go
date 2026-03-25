package nexus

import (
	"context"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/rs/zerolog"
)

// EventBus manages event publishing and subscription with glob pattern matching.
// It routes events to subscribers based on pattern matching (e.g., "order.*", "*").
type EventBus struct {
	mu              sync.RWMutex
	subscribers     map[string][]chan Event // pattern -> channels
	sentCount       atomic.Uint64
	droppedCount    atomic.Uint64
	log             zerolog.Logger
	ctx             context.Context
	regexCache      map[string]*regexp.Regexp // cache compiled regexes
	regexCacheMutex sync.RWMutex
}

// NewEventBus creates a new EventBus instance.
func NewEventBus(ctx context.Context, log zerolog.Logger) *EventBus {
	return &EventBus{
		subscribers: make(map[string][]chan Event),
		log:         log,
		ctx:         ctx,
		regexCache:  make(map[string]*regexp.Regexp),
	}
}

// Subscribe registers a new subscriber for events matching the given pattern.
// Returns a buffered channel (1024 elements) that will receive matching events.
// Patterns support glob-like matching:
//   - "order.*" matches "order.placed", "order.filled", "order.canceled"
//   - "*" matches any event type
//   - "order.placed" exact match only
func (eb *EventBus) Subscribe(pattern string) <-chan Event {
	ch := make(chan Event, 1024)

	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.subscribers[pattern] = append(eb.subscribers[pattern], ch)
	return ch
}

// Unsubscribe removes a subscriber from the given pattern.
// The channel is closed after removal.
func (eb *EventBus) Unsubscribe(pattern string, ch <-chan Event) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	chans, exists := eb.subscribers[pattern]
	if !exists {
		return
	}

	// Find and remove the channel
	for i, c := range chans {
		if c == ch {
			// Close the channel
			close(c)

			// Remove from slice
			eb.subscribers[pattern] = append(chans[:i], chans[i+1:]...)

			// Clean up empty pattern entries
			if len(eb.subscribers[pattern]) == 0 {
				delete(eb.subscribers, pattern)
			}
			return
		}
	}
}

// Publish sends an event to all subscribers whose patterns match the event type.
// Publishing is non-blocking: if a subscriber's channel is full, the event is dropped
// for that subscriber and droppedCount is incremented.
func (eb *EventBus) Publish(event Event) {
	eb.sentCount.Add(1)

	eb.mu.RLock()
	// Create a snapshot of subscribers to avoid holding lock during send
	subscribers := make(map[string][]chan Event)
	for pattern, chans := range eb.subscribers {
		// Copy the slice
		subscribers[pattern] = append([]chan Event{}, chans...)
	}
	eb.mu.RUnlock()

	// Send to matching subscribers
	eventTypeStr := string(event.Type)
	for pattern, chans := range subscribers {
		if eb.matchesPattern(pattern, eventTypeStr) {
			for _, ch := range chans {
				// Non-blocking send: drop if channel is full
				select {
				case ch <- event:
					// Successfully sent
				default:
					// Channel full - drop and log
					eb.droppedCount.Add(1)
					eb.log.Warn().
						Str("pattern", pattern).
						Str("event_type", eventTypeStr).
						Uint64("dropped_total", eb.droppedCount.Load()).
						Msg("event dropped - subscriber channel full")
				}
			}
		}
	}
}

// matchesPattern determines if an event type matches a subscription pattern.
// Supports glob-like patterns:
//   - "*" matches anything
//   - "prefix.*" matches "prefix.anything"
//   - exact string matches require exact equality
func (eb *EventBus) matchesPattern(pattern, eventType string) bool {
	// Exact match (no wildcards)
	if !strings.Contains(pattern, "*") {
		return pattern == eventType
	}

	// Wildcard only
	if pattern == "*" {
		return true
	}

	// Convert glob pattern to regex and cache it
	regex := eb.getCompiledRegex(pattern)
	if regex == nil {
		// Invalid pattern - never matches
		return false
	}

	return regex.MatchString(eventType)
}

// getCompiledRegex returns a cached compiled regex for the pattern, or compiles and caches it.
func (eb *EventBus) getCompiledRegex(pattern string) *regexp.Regexp {
	eb.regexCacheMutex.RLock()
	regex, exists := eb.regexCache[pattern]
	eb.regexCacheMutex.RUnlock()

	if exists {
		return regex
	}

	// Compile and cache
	regexStr := globToRegex(pattern)
	regex, err := regexp.Compile("^" + regexStr + "$")
	if err != nil {
		return nil
	}

	eb.regexCacheMutex.Lock()
	eb.regexCache[pattern] = regex
	eb.regexCacheMutex.Unlock()

	return regex
}

// globToRegex converts a glob pattern to a regex string.
// Only handles "." as literal and "*" as wildcard.
func globToRegex(pattern string) string {
	var result strings.Builder
	for i := 0; i < len(pattern); i++ {
		c := pattern[i]
		switch c {
		case '*':
			result.WriteString(".*")
		case '.':
			result.WriteString("\\.")
		case '?', '[', ']', '(', ')', '+', '^', '$', '|', '{', '}', '\\':
			// Escape regex special characters
			result.WriteByte('\\')
			result.WriteByte(c)
		default:
			result.WriteByte(c)
		}
	}
	return result.String()
}

// Stats returns current statistics about the EventBus.
// Returns a map with keys: "sent", "dropped", "subscribers"
func (eb *EventBus) Stats() map[string]interface{} {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	// Count total subscribers
	totalSubs := uint64(0)
	for _, chans := range eb.subscribers {
		totalSubs += uint64(len(chans))
	}

	return map[string]interface{}{
		"sent":        eb.sentCount.Load(),
		"dropped":     eb.droppedCount.Load(),
		"subscribers": totalSubs,
	}
}

// DroppedCount returns the total number of events dropped due to full channels.
func (eb *EventBus) DroppedCount() uint64 {
	return eb.droppedCount.Load()
}

// Close gracefully shuts down the EventBus by closing all subscriber channels
// and clearing the subscribers map.
func (eb *EventBus) Close() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	// Close all channels
	for _, chans := range eb.subscribers {
		for _, ch := range chans {
			close(ch)
		}
	}

	// Clear all subscribers
	eb.subscribers = make(map[string][]chan Event)
}
