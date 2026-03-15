package tui_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/atlasdev/orbitron/internal/tui"
)

func TestEventBus_SendAndReceive(t *testing.T) {
	bus := tui.NewEventBus()
	msg := tui.BotEventMsg{Level: "info", Message: "hello"}
	bus.Send(msg)
	cmd := bus.WaitForEvent()
	received := cmd()
	assert.Equal(t, msg, received)
}

func TestEventBus_Tap_ReceivesCopy(t *testing.T) {
	bus := tui.NewEventBus()
	tap := bus.Tap()
	msg := tui.BotEventMsg{Level: "warn", Message: "tap test"}
	bus.Send(msg)
	select {
	case received := <-tap:
		assert.Equal(t, msg, received)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("tap did not receive message")
	}
}

func TestEventBus_MultipleTaps(t *testing.T) {
	bus := tui.NewEventBus()
	tap1 := bus.Tap()
	tap2 := bus.Tap()
	msg := tui.BalanceMsg{USDC: 100.5}
	bus.Send(msg)
	select {
	case got := <-tap1:
		assert.Equal(t, msg, got)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("tap1 did not receive")
	}
	select {
	case got := <-tap2:
		assert.Equal(t, msg, got)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("tap2 did not receive")
	}
}

func TestEventBus_Send_NonBlocking_WhenFull(t *testing.T) {
	bus := tui.NewEventBus()
	for i := 0; i < 512; i++ {
		bus.Send(tui.BotEventMsg{Message: "fill"})
	}
	done := make(chan struct{})
	go func() {
		bus.Send(tui.BotEventMsg{Message: "overflow"})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Send blocked on full buffer")
	}
}

func TestEventBus_DifferentMessageTypes(t *testing.T) {
	bus := tui.NewEventBus()
	tap := bus.Tap()
	msgs := []interface{}{
		tui.BotEventMsg{Level: "info", Message: "test"},
		tui.BalanceMsg{USDC: 50.0},
		tui.SubsystemStatusMsg{Name: "monitor", Active: true},
		tui.LanguageChangedMsg{},
	}
	for _, m := range msgs {
		bus.Send(m)
	}
	for _, expected := range msgs {
		select {
		case got := <-tap:
			assert.Equal(t, expected, got)
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("did not receive %T", expected)
		}
	}
}

func TestEventBus_WaitForEvent_ReturnsCmd(t *testing.T) {
	bus := tui.NewEventBus()
	msg := tui.SubsystemStatusMsg{Name: "test", Active: true}
	bus.Send(msg)
	cmd := bus.WaitForEvent()
	require.NotNil(t, cmd)
	got := cmd()
	assert.Equal(t, msg, got)
}
