package webui

import (
	"encoding/json"
	"testing"
)

func TestHubBroadcast(t *testing.T) {
	h := newHub()
	msgs := make(chan []byte, 1)
	// register a fake client
	h.register("test", msgs)

	h.broadcast(WsEvent{Type: "balance", Data: map[string]any{"usdc": 99.0}})

	select {
	case raw := <-msgs:
		var ev WsEvent
		if err := json.Unmarshal(raw, &ev); err != nil {
			t.Fatal(err)
		}
		if ev.Type != "balance" {
			t.Fatalf("expected balance, got %s", ev.Type)
		}
	default:
		t.Fatal("no message received")
	}
}
