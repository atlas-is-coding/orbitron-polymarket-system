package order

import (
	"time"

	"github.com/atlasdev/orbitron/internal/api/clob"
)

// IsOrderExpired checks if a GTD (Good Till Date) order has expired.
// Returns false for non-GTD orders.
// Returns false if ExpiresAt is 0 (no expiration set).
// Returns true if the current time exceeds the expiration timestamp.
func IsOrderExpired(order *clob.Order) bool {
	// Only GTD orders can expire
	if order.OrderType != clob.OrderTypeGTD {
		return false
	}

	// No expiration set
	if order.ExpiresAt == 0 {
		return false
	}

	// Check if current time exceeds expiration time
	return time.Now().UnixMilli() > order.ExpiresAt
}

// MarkExpired updates the order status to EXPIRED if the order has expired.
// This is a helper function for processing expired orders in the trades monitor.
func MarkExpired(order *clob.Order) {
	if IsOrderExpired(order) {
		order.Status = clob.StatusExpired
	}
}
