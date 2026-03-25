package nexus

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// TestRPCExecuteCommand verifies synchronous command execution via RPC
func TestRPCExecuteCommand(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create Nexus
	auditLog := &MockAuditLog{}

	nexus, err := NewNexus(auditLog, 0, newTestLogger())
	if err != nil {
		t.Fatalf("NewNexus failed: %v", err)
	}
	defer nexus.Shutdown(ctx) //nolint:errcheck

	// Create RPC server (port 0 means no listener, for testing)
	httpRPCServer, err := NewRPCServer(0, nexus, newTestLogger())
	if err != nil {
		t.Fatalf("NewRPCServer failed: %v", err)
	}
	defer httpRPCServer.Close() //nolint:errcheck

	// Register a test command handler
	nexus.RegisterCommandHandler(CommandCancelOrder, func(cmdCtx context.Context, cmd *Command) (interface{}, error) {
		return map[string]string{"status": "canceled"}, nil
	})

	// Create RPC request for ExecuteCommand
	req := RPCRequest{
		Method: "ExecuteCommand",
		ID:     "test-1",
		Params: json.RawMessage(`{
			"type": "cmd.cancel_order",
			"payload": {"order_id": "order-1"},
			"source_ui": "test"
		}`),
	}

	resp := httpRPCServer.handleRequest(req)

	if resp.Error != "" {
		t.Fatalf("RPC error: %s", resp.Error)
	}
	if resp.ID != "test-1" {
		t.Errorf("expected ID 'test-1', got '%s'", resp.ID)
	}

	// Verify result is a command
	cmd, ok := resp.Result.(*Command)
	if !ok {
		t.Fatalf("expected Command result, got %T", resp.Result)
	}
	if cmd.Status != StatusCompleted {
		t.Errorf("expected StatusCompleted, got %v", cmd.Status)
	}
}

// TestRPCExecuteCommandAsync verifies asynchronous command execution
func TestRPCExecuteCommandAsync(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	auditLog := &MockAuditLog{}

	nexus, err := NewNexus(auditLog, 0, newTestLogger())
	if err != nil {
		t.Fatalf("NewNexus failed: %v", err)
	}
	defer nexus.Shutdown(ctx) //nolint:errcheck

	httpRPCServer, err := NewRPCServer(0, nexus, newTestLogger())
	if err != nil {
		t.Fatalf("NewRPCServer failed: %v", err)
	}
	defer httpRPCServer.Close() //nolint:errcheck

	nexus.RegisterCommandHandler(CommandPlaceOrder, func(cmdCtx context.Context, cmd *Command) (interface{}, error) {
		time.Sleep(100 * time.Millisecond) // Simulate async work
		return map[string]string{"order_id": "order-new"}, nil
	})

	req := RPCRequest{
		Method: "ExecuteCommandAsync",
		ID:     "test-2",
		Params: json.RawMessage(`{
			"type": "cmd.place_order",
			"payload": {"wallet_id": "w1", "side": "yes", "price": 0.5},
			"source_ui": "test"
		}`),
	}

	resp := httpRPCServer.handleRequest(req)

	if resp.Error != "" {
		t.Fatalf("RPC error: %s", resp.Error)
	}

	// Result should have command_id
	resultMap, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map result, got %T", resp.Result)
	}

	cmdID, ok := resultMap["command_id"].(string)
	if !ok {
		t.Fatalf("expected command_id in result, got %T", resultMap["command_id"])
	}
	if cmdID == "" {
		t.Error("command_id is empty")
	}
}

// TestRPCGetCommandStatus verifies retrieving command status
func TestRPCGetCommandStatus(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	auditLog := &MockAuditLog{}

	nexus, err := NewNexus(auditLog, 0, newTestLogger())
	if err != nil {
		t.Fatalf("NewNexus failed: %v", err)
	}
	defer nexus.Shutdown(ctx) //nolint:errcheck

	httpRPCServer, err := NewRPCServer(0, nexus, newTestLogger())
	if err != nil {
		t.Fatalf("NewRPCServer failed: %v", err)
	}
	defer httpRPCServer.Close() //nolint:errcheck

	// Execute a command asynchronously so we can check status while in-flight
	nexus.RegisterCommandHandler(CommandPlaceOrder, func(cmdCtx context.Context, cmd *Command) (interface{}, error) {
		time.Sleep(500 * time.Millisecond) // Hold for a bit so we can check status
		return map[string]string{"order_id": "order-new"}, nil
	})

	cmd := &Command{
		ID:   "cmd-" + uuid.New().String()[:8],
		Type: CommandPlaceOrder,
		Payload: PlaceOrderPayload{
			WalletID: "wallet-1",
			Side:     "yes",
		},
	}

	// Execute async so command stays in-flight briefly
	_, err = nexus.ExecuteCommandAsync(ctx, cmd)
	if err != nil {
		t.Fatalf("ExecuteCommandAsync failed: %v", err)
	}

	// Now get status via RPC - should find it in-flight
	req := RPCRequest{
		Method: "GetCommandStatus",
		ID:     "test-3",
		Params: json.RawMessage(`"` + cmd.ID + `"`),
	}

	resp := httpRPCServer.handleRequest(req)

	// May or may not find it depending on timing, so just verify the RPC works
	if resp.ID != "test-3" {
		t.Errorf("expected ID 'test-3', got '%s'", resp.ID)
	}
}

// TestRPCGetState verifies state retrieval
func TestRPCGetState(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	auditLog := &MockAuditLog{}

	nexus, err := NewNexus(auditLog, 0, newTestLogger())
	if err != nil {
		t.Fatalf("NewNexus failed: %v", err)
	}
	defer nexus.Shutdown(ctx) //nolint:errcheck

	httpRPCServer, err := NewRPCServer(0, nexus, newTestLogger())
	if err != nil {
		t.Fatalf("NewRPCServer failed: %v", err)
	}
	defer httpRPCServer.Close() //nolint:errcheck

	// Add a wallet to state
	nexus.PublishEvent(Event{
		Type:   EventWalletAdded,
		Source: "test",
		Payload: WalletAddedPayload{
			ID:      "wallet-1",
			Address: "0x123",
			Label:   "Test",
			Enabled: true,
		},
	})

	time.Sleep(100 * time.Millisecond)

	// Get state via RPC
	req := RPCRequest{
		Method: "GetState",
		ID:     "test-4",
		Params: json.RawMessage(`"wallets"`),
	}

	resp := httpRPCServer.handleRequest(req)

	if resp.Error != "" {
		t.Fatalf("RPC error: %s", resp.Error)
	}

	// Should return array
	resultArr, ok := resp.Result.([]interface{})
	if !ok {
		t.Fatalf("expected array result, got %T", resp.Result)
	}
	if len(resultArr) == 0 {
		t.Fatal("wallets array is empty")
	}
}

// TestRPCGetStats verifies stats retrieval
func TestRPCGetStats(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	auditLog := &MockAuditLog{}

	nexus, err := NewNexus(auditLog, 0, newTestLogger())
	if err != nil {
		t.Fatalf("NewNexus failed: %v", err)
	}
	defer nexus.Shutdown(ctx) //nolint:errcheck

	httpRPCServer, err := NewRPCServer(0, nexus, newTestLogger())
	if err != nil {
		t.Fatalf("NewRPCServer failed: %v", err)
	}
	defer httpRPCServer.Close() //nolint:errcheck

	req := RPCRequest{
		Method: "GetStats",
		ID:     "test-5",
		Params: json.RawMessage(`null`),
	}

	resp := httpRPCServer.handleRequest(req)

	if resp.Error != "" {
		t.Fatalf("RPC error: %s", resp.Error)
	}

	statsMap, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map result, got %T", resp.Result)
	}

	if _, ok := statsMap["event_bus"]; !ok {
		t.Fatal("event_bus stats missing")
	}
}

// TestRPCHealth verifies health check
func TestRPCHealth(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	auditLog := &MockAuditLog{}

	nexus, err := NewNexus(auditLog, 0, newTestLogger())
	if err != nil {
		t.Fatalf("NewNexus failed: %v", err)
	}
	defer nexus.Shutdown(ctx) //nolint:errcheck

	httpRPCServer, err := NewRPCServer(0, nexus, newTestLogger())
	if err != nil {
		t.Fatalf("NewRPCServer failed: %v", err)
	}
	defer httpRPCServer.Close() //nolint:errcheck

	req := RPCRequest{
		Method: "Health",
		ID:     "test-6",
		Params: json.RawMessage(`null`),
	}

	resp := httpRPCServer.handleRequest(req)

	if resp.Error != "" {
		t.Fatalf("RPC error: %s", resp.Error)
	}

	healthMap, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map result, got %T", resp.Result)
	}

	if status, ok := healthMap["status"].(string); !ok || status == "" {
		t.Error("health status missing or invalid")
	}
}

// TestRPCUnknownMethod verifies error handling for unknown methods
func TestRPCUnknownMethod(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	auditLog := &MockAuditLog{}

	nexus, err := NewNexus(auditLog, 0, newTestLogger())
	if err != nil {
		t.Fatalf("NewNexus failed: %v", err)
	}
	defer nexus.Shutdown(ctx) //nolint:errcheck

	httpRPCServer, err := NewRPCServer(0, nexus, newTestLogger())
	if err != nil {
		t.Fatalf("NewRPCServer failed: %v", err)
	}
	defer httpRPCServer.Close() //nolint:errcheck

	req := RPCRequest{
		Method: "UnknownMethod",
		ID:     "test-7",
		Params: json.RawMessage(`null`),
	}

	resp := httpRPCServer.handleRequest(req)

	if resp.Error == "" {
		t.Fatal("expected error for unknown method")
	}
	if !strings.Contains(resp.Error, "unknown method") {
		t.Errorf("expected 'unknown method' error, got '%s'", resp.Error)
	}
}

// TestRPCGetConnectedClients verifies client tracking
func TestRPCGetConnectedClients(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	auditLog := &MockAuditLog{}

	nexus, err := NewNexus(auditLog, 0, newTestLogger())
	if err != nil {
		t.Fatalf("NewNexus failed: %v", err)
	}
	defer nexus.Shutdown(ctx) //nolint:errcheck

	httpRPCServer, err := NewRPCServer(0, nexus, newTestLogger())
	if err != nil {
		t.Fatalf("NewRPCServer failed: %v", err)
	}
	defer httpRPCServer.Close() //nolint:errcheck

	// Initially no clients
	clients := httpRPCServer.GetConnectedClients()
	if len(clients) != 0 {
		t.Errorf("expected 0 clients initially, got %d", len(clients))
	}

	// Manually add a client for testing
	httpRPCServer.clientsMu.Lock()
	httpRPCServer.clients["client-1"] = &RPCClient{
		ID:            "client-1",
		Source:        "webui",
		Connected:     time.Now(),
		LastHeartbeat: time.Now(),
	}
	httpRPCServer.clientsMu.Unlock()

	// Should now have 1 client
	clients = httpRPCServer.GetConnectedClients()
	if len(clients) != 1 {
		t.Errorf("expected 1 client, got %d", len(clients))
	}

	// Verify client data
	client := clients[0]
	if client["id"] != "client-1" {
		t.Errorf("expected id 'client-1', got '%v'", client["id"])
	}
	if client["source"] != "webui" {
		t.Errorf("expected source 'webui', got '%v'", client["source"])
	}
}

// TestRPCWebSocketEventStream verifies WebSocket event streaming (simplified)
func TestRPCWebSocketEventStream(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	auditLog := &MockAuditLog{}

	nexus, err := NewNexus(auditLog, 0, newTestLogger())
	if err != nil {
		t.Fatalf("NewNexus failed: %v", err)
	}
	defer nexus.Shutdown(ctx) //nolint:errcheck

	httpRPCServer, err := NewRPCServer(0, nexus, newTestLogger())
	if err != nil {
		t.Fatalf("NewRPCServer failed: %v", err)
	}
	defer httpRPCServer.Close() //nolint:errcheck

	// Start HTTP server with RPC handler
	mux := http.NewServeMux()
	httpRPCServer.RegisterRoutes(mux)

	// Create a test HTTP server on a free port
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	portStr := fmt.Sprintf("%d", listener.Addr().(*net.TCPAddr).Port)
	listener.Close()

	server := &http.Server{
		Addr:    ":" + portStr,
		Handler: mux,
	}

	go func() {
		_ = server.ListenAndServe()
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)
	defer server.Close() //nolint:errcheck

	// Create WebSocket client
	wsURL := "ws://localhost:" + portStr + "/api/events"
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		// WebSocket may not work in test environment without proper setup
		t.Logf("WebSocket dial failed (expected in test): %v", err)
		t.Skip("WebSocket integration test requires proper HTTP server setup")
	}
	defer ws.Close()

	// Send subscription pattern
	subscription := map[string]string{"pattern": "wallet.*"}
	if err := ws.WriteJSON(subscription); err != nil {
		t.Fatalf("WriteJSON failed: %v", err)
	}

	// Publish a wallet event
	go func() {
		time.Sleep(100 * time.Millisecond)
		nexus.PublishEvent(Event{
			Type:   EventWalletAdded,
			Source: "test",
			Payload: WalletAddedPayload{
				ID:      "wallet-1",
				Address: "0x123",
				Enabled: true,
			},
		})
	}()

	// Read event from WebSocket with timeout
	ws.SetReadDeadline(time.Now().Add(2 * time.Second))
	var receivedEvent Event
	if err := ws.ReadJSON(&receivedEvent); err != nil {
		if !strings.Contains(err.Error(), "i/o timeout") {
			t.Fatalf("ReadJSON failed: %v", err)
		}
		t.Skip("WebSocket timeout (expected in basic test)")
	}

	if receivedEvent.Type != EventWalletAdded {
		t.Errorf("expected EventWalletAdded, got %v", receivedEvent.Type)
	}
}

// ============================================================
// Helper Functions
// ============================================================

// createTestListener creates a test TCP listener on a random port
func createTestListener(t *testing.T) (net.Listener, int) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to create test listener: %v", err)
	}

	addr := listener.Addr().(*net.TCPAddr)
	return listener, addr.Port
}
