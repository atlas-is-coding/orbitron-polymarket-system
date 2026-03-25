package nexus

import (
	"testing"
	"time"
)

func TestEventTypeValidation(t *testing.T) {
	tests := []struct {
		name      string
		eventType EventType
		valid     bool
	}{
		// Valid order events
		{"EventOrderPlaced", EventOrderPlaced, true},
		{"EventOrderFilled", EventOrderFilled, true},
		{"EventOrderCanceled", EventOrderCanceled, true},
		// Valid position events
		{"EventPositionOpened", EventPositionOpened, true},
		{"EventPositionClosed", EventPositionClosed, true},
		// Valid wallet events
		{"EventWalletAdded", EventWalletAdded, true},
		{"EventWalletRemoved", EventWalletRemoved, true},
		{"EventWalletChanged", EventWalletChanged, true},
		{"EventBalanceUpdated", EventBalanceUpdated, true},
		// Valid strategy events
		{"EventStrategyStarted", EventStrategyStarted, true},
		{"EventStrategyStopped", EventStrategyStopped, true},
		{"EventStrategyAlert", EventStrategyAlert, true},
		// Valid market events
		{"EventMarketsUpdated", EventMarketsUpdated, true},
		{"EventPriceAlert", EventPriceAlert, true},
		// Valid system events
		{"EventConfigReloaded", EventConfigReloaded, true},
		{"EventHealthSnapshot", EventHealthSnapshot, true},
		// Invalid
		{"InvalidEventType", EventType("invalid"), false},
		{"EmptyEventType", EventType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidEventType(tt.eventType); got != tt.valid {
				t.Errorf("IsValidEventType(%q) = %v, want %v", tt.eventType, got, tt.valid)
			}
		})
	}
}

func TestCommandTypeValidation(t *testing.T) {
	tests := []struct {
		name        string
		commandType CommandType
		valid       bool
	}{
		// Valid order commands
		{"CommandPlaceOrder", CommandPlaceOrder, true},
		{"CommandCancelOrder", CommandCancelOrder, true},
		{"CommandCancelAllOrders", CommandCancelAllOrders, true},
		// Valid wallet commands
		{"CommandAddWallet", CommandAddWallet, true},
		{"CommandRemoveWallet", CommandRemoveWallet, true},
		{"CommandToggleWallet", CommandToggleWallet, true},
		{"CommandUpdateWallet", CommandUpdateWallet, true},
		// Valid strategy commands
		{"CommandStartStrategy", CommandStartStrategy, true},
		{"CommandStopStrategy", CommandStopStrategy, true},
		// Valid config commands
		{"CommandReloadConfig", CommandReloadConfig, true},
		// Invalid
		{"InvalidCommandType", CommandType("invalid"), false},
		{"EmptyCommandType", CommandType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidCommandType(tt.commandType); got != tt.valid {
				t.Errorf("IsValidCommandType(%q) = %v, want %v", tt.commandType, got, tt.valid)
			}
		})
	}
}

func TestCommandStatusValidation(t *testing.T) {
	tests := []struct {
		name   string
		status CommandStatus
		valid  bool
	}{
		{"StatusPending", StatusPending, true},
		{"StatusProcessing", StatusProcessing, true},
		{"StatusCompleted", StatusCompleted, true},
		{"StatusFailed", StatusFailed, true},
		{"StatusTimedOut", StatusTimedOut, true},
		{"InvalidStatus", CommandStatus("invalid"), false},
		{"EmptyStatus", CommandStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidCommandStatus(tt.status); got != tt.valid {
				t.Errorf("IsValidCommandStatus(%q) = %v, want %v", tt.status, got, tt.valid)
			}
		})
	}
}

func TestEventStructure(t *testing.T) {
	t.Run("EventCreation", func(t *testing.T) {
		e := &Event{
			ID:        "test-id",
			Type:      EventOrderPlaced,
			Timestamp: time.Now(),
			Source:    "trades_monitor",
			Payload: &OrderPlacedPayload{
				OrderID:  "order-1",
				TokenID:  "123",
				WalletID: "wallet-1",
				Side:     "buy",
				Price:    0.5,
				SizeUSD:  100,
			},
			Critical: false,
		}

		if e.ID != "test-id" {
			t.Errorf("Event.ID = %q, want %q", e.ID, "test-id")
		}
		if e.Type != EventOrderPlaced {
			t.Errorf("Event.Type = %v, want %v", e.Type, EventOrderPlaced)
		}
		if e.Source != "trades_monitor" {
			t.Errorf("Event.Source = %q, want %q", e.Source, "trades_monitor")
		}
	})
}

func TestCommandStructure(t *testing.T) {
	t.Run("CommandCreation", func(t *testing.T) {
		c := &Command{
			ID:       "cmd-1",
			Type:     CommandPlaceOrder,
			Timestamp: time.Now(),
			SourceUI: "webui",
			Status:   StatusPending,
			Timeout:  30 * time.Second,
			Payload: &PlaceOrderPayload{
				ConditionID: "cond-1",
				WalletID:    "wallet-1",
				Side:        "buy",
				Price:       0.5,
				SizeUSD:     100,
				OrderType:   "limit",
			},
		}

		if c.ID != "cmd-1" {
			t.Errorf("Command.ID = %q, want %q", c.ID, "cmd-1")
		}
		if c.Type != CommandPlaceOrder {
			t.Errorf("Command.Type = %v, want %v", c.Type, CommandPlaceOrder)
		}
		if c.Status != StatusPending {
			t.Errorf("Command.Status = %v, want %v", c.Status, StatusPending)
		}
	})
}

func TestEventPayloadStructures(t *testing.T) {
	tests := []struct {
		name    string
		payload interface{}
	}{
		{"OrderPlacedPayload", &OrderPlacedPayload{OrderID: "o1"}},
		{"OrderFilledPayload", &OrderFilledPayload{OrderID: "o1"}},
		{"OrderCanceledPayload", &OrderCanceledPayload{OrderID: "o1"}},
		{"PositionOpenedPayload", &PositionOpenedPayload{PositionID: "p1"}},
		{"PositionClosedPayload", &PositionClosedPayload{PositionID: "p1"}},
		{"WalletAddedPayload", &WalletAddedPayload{ID: "w1"}},
		{"WalletRemovedPayload", &WalletRemovedPayload{ID: "w1"}},
		{"WalletChangedPayload", &WalletChangedPayload{ID: "w1"}},
		{"BalanceUpdatedPayload", &BalanceUpdatedPayload{WalletID: "w1"}},
		{"StrategyStartedPayload", &StrategyStartedPayload{Strategy: "s1"}},
		{"StrategyStoppedPayload", &StrategyStoppedPayload{Strategy: "s1"}},
		{"StrategyAlertPayload", &StrategyAlertPayload{Strategy: "s1"}},
		{"MarketsUpdatedPayload", &MarketsUpdatedPayload{Count: 10}},
		{"PriceAlertPayload", &PriceAlertPayload{ConditionID: "c1"}},
		{"ConfigReloadedPayload", &ConfigReloadedPayload{Timestamp: time.Now()}},
		{"HealthSnapshotPayload", &HealthSnapshotPayload{Status: "ok"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.payload == nil {
				t.Errorf("payload is nil")
			}
		})
	}
}

func TestCommandPayloadStructures(t *testing.T) {
	tests := []struct {
		name    string
		payload interface{}
	}{
		{"PlaceOrderPayload", &PlaceOrderPayload{ConditionID: "c1"}},
		{"CancelOrderPayload", &CancelOrderPayload{OrderID: "o1"}},
		{"AddWalletPayload", &AddWalletPayload{Address: "0x123"}},
		{"RemoveWalletPayload", &RemoveWalletPayload{WalletID: "w1"}},
		{"ToggleWalletPayload", &ToggleWalletPayload{WalletID: "w1"}},
		{"StartStrategyPayload", &StartStrategyPayload{StrategyName: "s1"}},
		{"StopStrategyPayload", &StopStrategyPayload{StrategyName: "s1"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.payload == nil {
				t.Errorf("payload is nil")
			}
		})
	}
}
