package nexus

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

// TestFullNexusFlow tests complete event-to-state-to-command flow
func TestFullNexusFlow(t *testing.T) {
	log := zerolog.New(nil)
	auditLog := NewMockAuditLog()
	nex, _ := NewNexus(auditLog, 0, log)
	defer nex.Shutdown(context.Background())

	nex.RegisterCommandHandler(CommandPlaceOrder, func(ctx context.Context, cmd *Command) (interface{}, error) {
		time.Sleep(50 * time.Millisecond)
		nex.PublishEvent(Event{
			Type: EventOrderPlaced,
			Source: "handler",
			Critical: true,
			Payload: &OrderPlacedPayload{OrderID: "ord_123", TokenID: "tok_1", WalletID: "w1", Side: "YES", Price: 0.5, SizeUSD: 100},
		})
		return map[string]string{"order_id": "ord_123"}, nil
	})

	eventChan := nex.Subscribe("order.*")
	cmdID, _ := nex.ExecuteCommandAsync(context.Background(), &Command{
		Type: CommandPlaceOrder,
		Payload: &PlaceOrderPayload{ConditionID: "cond_1", WalletID: "w1", Side: "YES", Price: 0.5, SizeUSD: 100},
	})

	select {
	case event := <-eventChan:
		if event.Type != EventOrderPlaced {
			t.Errorf("got %v, want order.placed", event.Type)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout")
	}

	time.Sleep(100 * time.Millisecond)
	orders := nex.GetState("orders").([]*OrderState)
	if len(orders) == 0 {
		t.Fatal("no orders in state")
	}

	cmd, _ := nex.GetCommandStatus(cmdID)
	if cmd.Status != StatusCompleted {
		t.Errorf("status %v, want completed", cmd.Status)
	}
}
