package nexus

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

// RPCRequest represents a JSON RPC request
type RPCRequest struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
	ID     string          `json:"id"`
}

// RPCResponse represents a JSON RPC response
type RPCResponse struct {
	ID     string      `json:"id"`
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
}

// RPCClient represents a connected RPC client
type RPCClient struct {
	ID            string
	Source        string // "webui", "telegram", etc.
	Connected     time.Time
	LastHeartbeat time.Time
}

// HTTPRPCServer provides JSON RPC and WebSocket interfaces to Nexus
type HTTPRPCServer struct {
	nexus     *Nexus
	listener  net.Listener
	log       zerolog.Logger
	ctx       context.Context
	cancel    context.CancelFunc
	clients   map[string]*RPCClient
	clientsMu sync.RWMutex
	wg        sync.WaitGroup
}

// NewRPCServer creates and starts a new RPC server
// Parameters:
//   - port: TCP port to listen on (0 means don't listen, for tests)
//   - nexus: Nexus coordinator instance
//   - log: zerolog logger instance
// Returns error if port binding fails
func NewRPCServer(port int, nexus *Nexus, log zerolog.Logger) (*HTTPRPCServer, error) {
	// Create context for lifecycle
	ctx, cancel := context.WithCancel(context.Background())

	rs := &HTTPRPCServer{
		nexus:   nexus,
		log:     log,
		ctx:     ctx,
		cancel:  cancel,
		clients: make(map[string]*RPCClient),
	}

	// Only create listener if port > 0
	if port > 0 {
		addr := fmt.Sprintf("localhost:%d", port)
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to bind RPC port %d: %w", port, err)
		}
		rs.listener = listener
		log.Info().Int("port", port).Msg("RPC server initialized with listener")
	} else {
		log.Info().Msg("RPC server initialized without listener (port 0)")
	}

	return rs, nil
}

// RegisterRoutes registers HTTP handlers with an http.ServeMux
// This registers:
//   - POST /api/rpc - JSON RPC endpoint
//   - GET /api/events - WebSocket event stream
func (rs *HTTPRPCServer) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/rpc", rs.handleRPCRequest)
	mux.HandleFunc("GET /api/events", rs.handleEventStream)
	rs.log.Info().Msg("RPC routes registered")
}

// handleRPCRequest handles HTTP JSON RPC requests
func (rs *HTTPRPCServer) handleRPCRequest(w http.ResponseWriter, r *http.Request) {
	// Check method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode request
	var req RPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Process request
	resp := rs.handleRequest(req)

	// Encode response
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// handleRequest processes an RPC request and returns a response
func (rs *HTTPRPCServer) handleRequest(req RPCRequest) RPCResponse {
	switch req.Method {
	case "ExecuteCommand":
		return rs.executeCommand(req)
	case "ExecuteCommandAsync":
		return rs.executeCommandAsync(req)
	case "GetCommandStatus":
		return rs.getCommandStatus(req)
	case "GetState":
		return rs.getState(req)
	case "GetStats":
		return rs.getStats(req)
	case "Health":
		return rs.health(req)
	default:
		return RPCResponse{
			ID:    req.ID,
			Error: fmt.Sprintf("unknown method: %s", req.Method),
		}
	}
}

// executeCommand handles synchronous command execution
func (rs *HTTPRPCServer) executeCommand(req RPCRequest) RPCResponse {
	var cmd Command
	if err := json.Unmarshal(req.Params, &cmd); err != nil {
		return RPCResponse{
			ID:    req.ID,
			Error: fmt.Sprintf("invalid command params: %v", err),
		}
	}

	// Set defaults if not provided
	if cmd.ID == "" {
		cmd.ID = uuid.New().String()
	}
	if cmd.Timestamp.IsZero() {
		cmd.Timestamp = time.Now()
	}

	// Execute synchronously
	ctx, cancel := context.WithTimeout(rs.ctx, 30*time.Second)
	defer cancel()

	result, err := rs.nexus.ExecuteCommand(ctx, &cmd)
	if err != nil {
		return RPCResponse{
			ID:    req.ID,
			Error: fmt.Sprintf("command execution failed: %v", err),
		}
	}

	return RPCResponse{
		ID:     req.ID,
		Result: result,
	}
}

// executeCommandAsync handles asynchronous command execution
func (rs *HTTPRPCServer) executeCommandAsync(req RPCRequest) RPCResponse {
	var cmd Command
	if err := json.Unmarshal(req.Params, &cmd); err != nil {
		return RPCResponse{
			ID:    req.ID,
			Error: fmt.Sprintf("invalid command params: %v", err),
		}
	}

	// Set defaults if not provided
	if cmd.ID == "" {
		cmd.ID = uuid.New().String()
	}
	if cmd.Timestamp.IsZero() {
		cmd.Timestamp = time.Now()
	}

	// Execute asynchronously
	ctx, cancel := context.WithTimeout(rs.ctx, 30*time.Second)
	defer cancel()

	cmdID, err := rs.nexus.ExecuteCommandAsync(ctx, &cmd)
	if err != nil {
		return RPCResponse{
			ID:    req.ID,
			Error: fmt.Sprintf("async command execution failed: %v", err),
		}
	}

	return RPCResponse{
		ID: req.ID,
		Result: map[string]interface{}{
			"command_id": cmdID,
		},
	}
}

// getCommandStatus handles getting command status
func (rs *HTTPRPCServer) getCommandStatus(req RPCRequest) RPCResponse {
	var cmdID string
	if err := json.Unmarshal(req.Params, &cmdID); err != nil {
		return RPCResponse{
			ID:    req.ID,
			Error: fmt.Sprintf("invalid command ID: %v", err),
		}
	}

	status, err := rs.nexus.GetCommandStatus(cmdID)
	if err != nil {
		return RPCResponse{
			ID:    req.ID,
			Error: fmt.Sprintf("command not found: %v", err),
		}
	}

	return RPCResponse{
		ID:     req.ID,
		Result: status,
	}
}

// getState handles state retrieval
func (rs *HTTPRPCServer) getState(req RPCRequest) RPCResponse {
	var stateType string
	if err := json.Unmarshal(req.Params, &stateType); err != nil {
		return RPCResponse{
			ID:    req.ID,
			Error: fmt.Sprintf("invalid state type: %v", err),
		}
	}

	state := rs.nexus.GetState(stateType)
	if state == nil {
		return RPCResponse{
			ID:    req.ID,
			Error: fmt.Sprintf("unknown state type: %s", stateType),
		}
	}

	// Convert to interface slice for JSON marshaling
	var result []interface{}
	if slice, ok := state.([]interface{}); ok {
		result = slice
	} else {
		result = []interface{}{state}
	}

	return RPCResponse{
		ID:     req.ID,
		Result: result,
	}
}

// getStats handles stats retrieval
func (rs *HTTPRPCServer) getStats(req RPCRequest) RPCResponse {
	stats := rs.nexus.GetStats()
	return RPCResponse{
		ID:     req.ID,
		Result: stats,
	}
}

// health handles health checks
func (rs *HTTPRPCServer) health(req RPCRequest) RPCResponse {
	health := rs.nexus.Health()
	return RPCResponse{
		ID:     req.ID,
		Result: health,
	}
}

// handleEventStream handles WebSocket connections for event streaming
func (rs *HTTPRPCServer) handleEventStream(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// Allow all origins for now
			return true
		},
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		rs.log.Warn().Err(err).Msg("WebSocket upgrade failed")
		return
	}
	
	defer ws.Close()

	// Generate client ID
	clientID := uuid.New().String()
	client := &RPCClient{
		ID:            clientID,
		Source:        r.Header.Get("X-Client-Source"),
		Connected:     time.Now(),
		LastHeartbeat: time.Now(),
	}
	if client.Source == "" {
		client.Source = "unknown"
	}

	// Register client
	rs.clientsMu.Lock()
	rs.clients[clientID] = client
	rs.clientsMu.Unlock()

	defer func() {
		rs.clientsMu.Lock()
		delete(rs.clients, clientID)
		rs.clientsMu.Unlock()
		rs.log.Debug().Str("client_id", clientID).Msg("client disconnected")
	}()

	rs.log.Debug().
		Str("client_id", clientID).
		Str("source", client.Source).
		Msg("client connected to event stream")

	// Wait for subscription message with pattern
	var subscription map[string]string
	if err := ws.ReadJSON(&subscription); err != nil {
		rs.log.Warn().Err(err).Msg("failed to read subscription")
		return
	}

	pattern := subscription["pattern"]
	if pattern == "" {
		pattern = "*" // Default to all events
	}

	rs.log.Debug().
		Str("client_id", clientID).
		Str("pattern", pattern).
		Msg("client subscribed to pattern")

	// Subscribe to events from Nexus
	eventChan := rs.nexus.Subscribe(pattern)
	defer rs.nexus.Unsubscribe(pattern, eventChan)

	// Stream events to WebSocket
	for {
		select {
		case <-rs.ctx.Done():
			return
		case event := <-eventChan:
			// Update last heartbeat
			rs.clientsMu.Lock()
			if c, ok := rs.clients[clientID]; ok {
				c.LastHeartbeat = time.Now()
			}
			rs.clientsMu.Unlock()

			// Send event to client
			if err := ws.WriteJSON(event); err != nil {
				rs.log.Debug().
					Str("client_id", clientID).
					Err(err).
					Msg("failed to write event to WebSocket")
				return
			}
		}
	}
}

// GetConnectedClients returns a list of all connected RPC clients
func (rs *HTTPRPCServer) GetConnectedClients() []map[string]interface{} {
	rs.clientsMu.RLock()
	defer rs.clientsMu.RUnlock()

	clients := make([]map[string]interface{}, 0, len(rs.clients))
	for _, client := range rs.clients {
		clients = append(clients, map[string]interface{}{
			"id":               client.ID,
			"source":           client.Source,
			"connected_at":     client.Connected,
			"last_heartbeat":   client.LastHeartbeat,
			"uptime_seconds":   time.Since(client.Connected).Seconds(),
		})
	}
	return clients
}

// GetConnectedClientsCount returns the number of connected clients
func (rs *HTTPRPCServer) GetConnectedClientsCount() int {
	rs.clientsMu.RLock()
	defer rs.clientsMu.RUnlock()
	return len(rs.clients)
}

// Close gracefully shuts down the RPC server
func (rs *HTTPRPCServer) Close() error {
	rs.log.Info().Msg("shutting down RPC server")

	// Cancel context to stop event streaming
	rs.cancel()

	// Close all client connections
	rs.clientsMu.Lock()
	for _, client := range rs.clients {
		rs.log.Debug().Str("client_id", client.ID).Msg("closing client connection")
	}
	rs.clients = make(map[string]*RPCClient)
	rs.clientsMu.Unlock()

	// Close listener
	if rs.listener != nil {
		if err := rs.listener.Close(); err != nil {
			rs.log.Warn().Err(err).Msg("error closing RPC listener")
		}
	}

	// Wait for any pending operations
	rs.wg.Wait()

	rs.log.Info().Msg("RPC server shut down successfully")
	return nil
}
