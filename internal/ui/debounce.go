package ui

import (
	"sync"
	"time"
)

// Debouncer prevents rapid successive actions by enforcing a minimum time interval.
type Debouncer struct {
	mu        sync.Mutex
	lastPress time.Time
	delay     time.Duration
}

// NewDebouncer creates a new Debouncer with the specified delay.
func NewDebouncer(delay time.Duration) *Debouncer {
	return &Debouncer{
		delay:     delay,
		lastPress: time.Time{}, // zero time, allows first press immediately
	}
}

// Allow returns true if enough time has passed since the last allowed action.
// If the action is allowed, it updates lastPress and returns true.
// If too little time has passed, it returns false without updating lastPress.
func (d *Debouncer) Allow() bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	if now.Sub(d.lastPress) < d.delay {
		return false
	}

	d.lastPress = now
	return true
}
