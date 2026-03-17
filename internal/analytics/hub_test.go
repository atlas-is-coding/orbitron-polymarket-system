package analytics

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestAnalyticsHub_RecordAndFlush(t *testing.T) {
	hub := NewAnalyticsHub()

	trade1 := TradeReport{ID: "t1", Volume: 100}
	trade2 := TradeReport{ID: "t2", Volume: 200}

	hub.RecordTrade(trade1)
	hub.RecordTrade(trade2)

	trades := hub.Flush()
	assert.Len(t, trades, 2)
	assert.Equal(t, "t1", trades[0].ID)
	assert.Equal(t, "t2", trades[1].ID)

	// Flush should clear the buffer
	trades2 := hub.Flush()
	assert.Empty(t, trades2)
}
