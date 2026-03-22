// internal/builder/logger.go
package builder

import (
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// OrderLogEntry holds metadata for a single order submission.
type OrderLogEntry struct {
	OrderID       string
	BuilderKeySet bool
	Timestamp     time.Time
	Success       bool
}

// OrderExecutionLogger tracks per-order builder key attribution and logs summaries.
type OrderExecutionLogger struct {
	mu           sync.Mutex
	totalOrders  int64
	withKey      int64
	withoutKey   int64
	log          zerolog.Logger
}

// NewOrderExecutionLogger creates a logger. Call LogOrder after each order submission.
func NewOrderExecutionLogger(log zerolog.Logger) *OrderExecutionLogger {
	return &OrderExecutionLogger{log: log}
}

// LogOrder records a single order. Thread-safe. Logs summary every 100 orders.
func (l *OrderExecutionLogger) LogOrder(entry OrderLogEntry) {
	l.mu.Lock()
	l.totalOrders++
	if entry.BuilderKeySet {
		l.withKey++
	} else {
		l.withoutKey++
	}
	total := l.totalOrders
	withKey := l.withKey
	withoutKey := l.withoutKey
	l.mu.Unlock()

	l.log.Debug().
		Str("order_id", entry.OrderID).
		Bool("builder_key_set", entry.BuilderKeySet).
		Bool("success", entry.Success).
		Msg("builder: order submitted")

	if withoutKey > 0 {
		l.log.Error().
			Int64("without_key", withoutKey).
			Msg("builder: order submitted WITHOUT builder key — attribution missing")
	}

	if total%100 == 0 {
		l.log.Info().
			Int64("total", total).
			Int64("with_key", withKey).
			Int64("without_key", withoutKey).
			Msg("builder: order attribution summary")
	}
}

// Summary returns current counters. Thread-safe.
func (l *OrderExecutionLogger) Summary() (total, withKey, withoutKey int64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.totalOrders, l.withKey, l.withoutKey
}
