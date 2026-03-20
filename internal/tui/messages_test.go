package tui

import (
	"testing"
)

func TestEventBusDropCounter(t *testing.T) {
	bus := NewEventBus()

	// Send more messages than the buffer holds (16384) without a reader.
	for i := 0; i < 20000; i++ {
		bus.Send(BotEventMsg{Message: "overflow"})
	}

	dropped := bus.DroppedCount()
	if dropped == 0 {
		t.Error("expected some messages to be dropped, got 0")
	}
	t.Logf("dropped %d / 20000 messages (%.1f%%)", dropped, float64(dropped)/200.0)
}
