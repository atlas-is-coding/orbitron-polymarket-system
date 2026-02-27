package webui

import (
	"testing"

	"github.com/atlasdev/polytrade-bot/internal/tui"
)

func TestStateBalance(t *testing.T) {
	s := newWebState()
	s.SetBalance(42.5)
	if got := s.Balance(); got != 42.5 {
		t.Fatalf("want 42.5, got %v", got)
	}
}

func TestStateOrders(t *testing.T) {
	s := newWebState()
	rows := []tui.OrderRow{{ID: "abc", Market: "BTC"}}
	s.SetOrders(rows)
	got := s.Orders()
	if len(got) != 1 || got[0].ID != "abc" {
		t.Fatalf("unexpected orders: %v", got)
	}
}

func TestStateLogsBuffer(t *testing.T) {
	s := newWebState()
	for range 300 {
		s.AddLog("info", "msg")
	}
	if len(s.Logs()) > 200 {
		t.Fatal("logs buffer not capped")
	}
}
