package order

import (
	"testing"
	"time"

	"github.com/atlasdev/orbitron/internal/api/clob"
)

func TestGTDOrderExpiration(t *testing.T) {
	tests := []struct {
		name     string
		order    *clob.Order
		expected bool
	}{
		{
			name: "GTD order with expired timestamp",
			order: &clob.Order{
				ID:        "test-order-1",
				Status:    clob.StatusLive,
				OrderType: clob.OrderTypeGTD,
				ExpiresAt: time.Now().Add(-1 * time.Hour).UnixMilli(), // 1 hour ago
			},
			expected: true,
		},
		{
			name: "GTD order with future timestamp",
			order: &clob.Order{
				ID:        "test-order-2",
				Status:    clob.StatusLive,
				OrderType: clob.OrderTypeGTD,
				ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli(), // 1 hour from now
			},
			expected: false,
		},
		{
			name: "GTD order with zero expiration",
			order: &clob.Order{
				ID:        "test-order-3",
				Status:    clob.StatusLive,
				OrderType: clob.OrderTypeGTD,
				ExpiresAt: 0,
			},
			expected: false,
		},
		{
			name: "GTC order with any expiration",
			order: &clob.Order{
				ID:        "test-order-4",
				Status:    clob.StatusLive,
				OrderType: clob.OrderTypeGTC,
				ExpiresAt: time.Now().Add(-1 * time.Hour).UnixMilli(),
			},
			expected: false,
		},
		{
			name: "FOK order with any expiration",
			order: &clob.Order{
				ID:        "test-order-5",
				Status:    clob.StatusLive,
				OrderType: clob.OrderTypeFOK,
				ExpiresAt: time.Now().Add(-1 * time.Hour).UnixMilli(),
			},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := IsOrderExpired(tc.order)
			if result != tc.expected {
				t.Errorf("IsOrderExpired(%v) = %v, want %v", tc.order, result, tc.expected)
			}
		})
	}
}
