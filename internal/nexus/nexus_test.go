package nexus

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestNexusEventPublishToStateStore verifies that published events update the state store
func TestNexusEventPublishToStateStore(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	auditLog := &MockAuditLog{}
	nexus, err := NewNexus(auditLog, 0, newTestLogger())
	if err != nil {
		t.Fatalf("NewNexus failed: %v", err)
	}
	defer nexus.Shutdown(ctx) //nolint:errcheck

	// Publish a wallet added event
	event := Event{
		Type:      EventWalletAdded,
		Source:    "test",
		Timestamp: time.Now(),
		Payload: WalletAddedPayload{
			ID:      "wallet-1",
			Address: "0x123",
			Label:   "Test Wallet",
			Enabled: true,
			Primary: true,
		},
	}

	nexus.PublishEvent(event)

	// Wait for event processing
	time.Sleep(100 * time.Millisecond)

	// Verify state was updated
	wallet := nexus.state.GetWallet("wallet-1")
	if wallet == nil {
		t.Fatal("wallet not found in state store")
	}
	if wallet.Address != "0x123" || wallet.Label != "Test Wallet" {
		t.Errorf("wallet data mismatch: got %+v", wallet)
	}
}

// TestNexusEventSubscription verifies event subscription and receipt
func TestNexusEventSubscription(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	auditLog := &MockAuditLog{}
	nexus, err := NewNexus(auditLog, 0, newTestLogger())
	if err != nil {
		t.Fatalf("NewNexus failed: %v", err)
	}
	defer nexus.Shutdown(ctx) //nolint:errcheck

	// Subscribe to order events
	ch := nexus.Subscribe("order.*")
	if ch == nil {
		t.Fatal("Subscribe returned nil channel")
	}
	defer nexus.Unsubscribe("order.*", ch)

	// Publish an order placed event
	event := Event{
		ID:        uuid.New().String(),
		Type:      EventOrderPlaced,
		Source:    "test",
		Timestamp: time.Now(),
		Payload: OrderPlacedPayload{
			OrderID:  "order-1",
			TokenID:  "token-123",
			WalletID: "wallet-1",
			Side:     "yes",
			Price:    0.65,
			SizeUSD:  100.0,
		},
	}

	nexus.PublishEvent(event)

	// Receive the event on subscription
	select {
	case received := <-ch:
		if received.Type != EventOrderPlaced {
			t.Errorf("expected EventOrderPlaced, got %v", received.Type)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("did not receive event on subscription")
	}
}

// TestNexusCommandExecution verifies synchronous command execution
func TestNexusCommandExecution(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	auditLog := &MockAuditLog{}
	nexus, err := NewNexus(auditLog, 0, newTestLogger())
	if err != nil {
		t.Fatalf("NewNexus failed: %v", err)
	}
	defer nexus.Shutdown(ctx) //nolint:errcheck

	// Register a handler for cancel order command
	nexus.RegisterCommandHandler(CommandCancelOrder, func(cmdCtx context.Context, cmd *Command) (interface{}, error) {
		return map[string]string{"canceled": "true"}, nil
	})

	// Execute the command
	cmd := &Command{
		Type: CommandCancelOrder,
		Payload: CancelOrderPayload{
			OrderID: "order-1",
		},
	}

	result, err := nexus.ExecuteCommand(ctx, cmd)
	if err != nil {
		t.Fatalf("ExecuteCommand failed: %v", err)
	}

	if result.Status != StatusCompleted {
		t.Errorf("expected StatusCompleted, got %v", result.Status)
	}
}

// TestNexusStats verifies GetStats returns aggregated statistics
func TestNexusStats(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	auditLog := &MockAuditLog{}
	nexus, err := NewNexus(auditLog, 0, newTestLogger())
	if err != nil {
		t.Fatalf("NewNexus failed: %v", err)
	}
	defer nexus.Shutdown(ctx) //nolint:errcheck

	// Publish an event
	nexus.PublishEvent(Event{
		Type:   EventWalletAdded,
		Source: "test",
		Payload: WalletAddedPayload{
			ID: "wallet-1",
		},
	})

	time.Sleep(100 * time.Millisecond)

	// Get stats
	stats := nexus.GetStats()
	if stats == nil {
		t.Fatal("GetStats returned nil")
	}

	// Verify some keys exist
	if _, ok := stats["event_bus"]; !ok {
		t.Fatal("event_bus stats missing")
	}
	if _, ok := stats["command_processor"]; !ok {
		t.Fatal("command_processor stats missing")
	}
	if _, ok := stats["state_store"]; !ok {
		t.Fatal("state_store stats missing")
	}
}

// TestNexusShutdown verifies graceful shutdown
func TestNexusShutdown(t *testing.T) {
	auditLog := &MockAuditLog{}
	nexus, err := NewNexus(auditLog, 0, newTestLogger())
	if err != nil {
		t.Fatalf("NewNexus failed: %v", err)
	}

	// Publish some events
	for i := 0; i < 3; i++ {
		nexus.PublishEvent(Event{
			Type:   EventWalletAdded,
			Source: "test",
			Payload: WalletAddedPayload{
				ID: "wallet-" + string(rune(i)),
			},
		})
	}

	// Shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer shutdownCancel()

	err = nexus.Shutdown(shutdownCtx)
	if err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}
}

// TestNexusGetState verifies state retrieval by type
func TestNexusGetState(t *testing.T) {
	auditLog := &MockAuditLog{}
	nexus, err := NewNexus(auditLog, 0, newTestLogger())
	if err != nil {
		t.Fatalf("NewNexus failed: %v", err)
	}
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer shutdownCancel()
		nexus.Shutdown(shutdownCtx) //nolint:errcheck
	}()

	// Add a wallet event
	nexus.PublishEvent(Event{
		Type:   EventWalletAdded,
		Source: "test",
		Payload: WalletAddedPayload{
			ID: "wallet-1",
		},
	})

	time.Sleep(100 * time.Millisecond)

	// Test GetState for wallets
	wallets := nexus.GetState("wallets")
	if wallets == nil {
		t.Fatal("GetState(wallets) returned nil")
	}

	walletsList, ok := wallets.([]*WalletState)
	if !ok {
		t.Fatalf("unexpected type for wallets: %T", wallets)
	}

	if len(walletsList) == 0 {
		t.Fatal("wallets list is empty")
	}

	// Test other state types
	orders := nexus.GetState("orders")
	if orders == nil {
		t.Fatal("GetState(orders) returned nil")
	}

	positions := nexus.GetState("positions")
	if positions == nil {
		t.Fatal("GetState(positions) returned nil")
	}

	strategies := nexus.GetState("strategies")
	if strategies == nil {
		t.Fatal("GetState(strategies) returned nil")
	}

	markets := nexus.GetState("markets")
	if markets == nil {
		t.Fatal("GetState(markets) returned nil")
	}

	health := nexus.GetState("health")
	if health == nil {
		t.Fatal("GetState(health) returned nil")
	}

	// Test unknown state type
	unknown := nexus.GetState("unknown")
	if unknown != nil {
		t.Fatal("GetState(unknown) should return nil")
	}
}

// TestFullNexusFlow is an integration test of complete Nexus flow
func TestFullNexusFlow(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	auditLog := &MockAuditLog{}
	nexus, err := NewNexus(auditLog, 0, newTestLogger())
	if err != nil {
		t.Fatalf("NewNexus failed: %v", err)
	}
	defer nexus.Shutdown(ctx) //nolint:errcheck

	// Register a command handler that publishes an event
	nexus.RegisterCommandHandler(CommandPlaceOrder, func(cmdCtx context.Context, cmd *Command) (interface{}, error) {
		// Simulate placing an order and publishing event
		orderID := "order-" + uuid.New().String()[:8]
		nexus.PublishEvent(Event{
			Type:   EventOrderPlaced,
			Source: "CommandHandler",
			Payload: OrderPlacedPayload{
				OrderID:  orderID,
				TokenID:  "token-123",
				WalletID: "wallet-1",
				Side:     "yes",
				Price:    0.65,
				SizeUSD:  100.0,
			},
		})
		return map[string]string{"order_id": orderID}, nil
	})

	// Subscribe to order events
	ch := nexus.Subscribe("order.*")
	defer nexus.Unsubscribe("order.*", ch)

	// Execute async command
	cmd := &Command{
		Type: CommandPlaceOrder,
		Payload: PlaceOrderPayload{
			WalletID: "wallet-1",
			Side:     "yes",
			Price:    0.65,
			SizeUSD:  100.0,
		},
	}

	cmdID, err := nexus.ExecuteCommandAsync(ctx, cmd)
	if err != nil {
		t.Fatalf("ExecuteCommandAsync failed: %v", err)
	}

	// Receive event on subscription
	var receivedEvent Event
	select {
	case receivedEvent = <-ch:
		if receivedEvent.Type != EventOrderPlaced {
			t.Fatalf("expected EventOrderPlaced, got %v", receivedEvent.Type)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("did not receive order placed event")
	}

	// Verify state has the order
	time.Sleep(100 * time.Millisecond)
	orders := nexus.state.GetAllOrders()
	if len(orders) == 0 {
		t.Fatal("no orders in state store after event")
	}

	// Poll command status until completed
	var completedCmd *Command
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		status, err := nexus.GetCommandStatus(cmdID)
		if err != nil {
			t.Logf("GetCommandStatus error (may be in progress): %v", err)
			time.Sleep(50 * time.Millisecond)
			continue
		}
		if status.Status == StatusCompleted || status.Status == StatusFailed || status.Status == StatusTimedOut {
			completedCmd = status
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	if completedCmd == nil {
		t.Fatal("command did not complete in time")
	}

	if completedCmd.Status != StatusCompleted {
		t.Errorf("expected StatusCompleted, got %v", completedCmd.Status)
	}
}
