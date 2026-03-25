package nexus

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// RPCServer represents an optional RPC server for remote Nexus access
type RPCServer interface {
	Close() error
}

// Nexus is the central coordinator that manages events, commands, and state
type Nexus struct {
	eventBus      *EventBus
	cmdProcessor  *CommandProcessor
	state         *StateStore
	auditLog      AuditLog
	rpcServer     RPCServer

	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup

	eventHandler  chan Event
	log           zerolog.Logger
}

// NewNexus creates a new Nexus coordinator instance
// Parameters:
//   - auditLog: storage.AuditLog implementation for event/command persistence
//   - rpcPort: RPC server port (0 means no RPC server)
//   - log: zerolog logger instance
// Returns error if RPC creation fails
func NewNexus(auditLog AuditLog, rpcPort int, log zerolog.Logger) (*Nexus, error) {
	// Create context for lifecycle management
	ctx, cancel := context.WithCancel(context.Background())

	n := &Nexus{
		eventBus:     NewEventBus(ctx, log),
		cmdProcessor: NewCommandProcessor(ctx, auditLog, 4, log),
		state:        NewStateStore(log),
		auditLog:     auditLog,
		rpcServer:    nil,

		ctx:          ctx,
		cancel:       cancel,
		eventHandler: make(chan Event, 512),
		log:          log,
	}

	// Create RPC server if port specified
	if rpcPort > 0 {
		// For now, we'll create a placeholder RPC server that can be closed
		// Task 6 will implement the actual RPC server
		rpcServer := &basicRPCServer{}
		n.rpcServer = rpcServer
		n.log.Info().Int("port", rpcPort).Msg("RPC server created (placeholder)")
	}

	// Start event handler loop
	n.wg.Add(1)
	go n.eventHandlerLoop()

	n.log.Info().Msg("Nexus coordinator initialized")
	return n, nil
}

// PublishEvent enqueues an event for processing
// Non-blocking: if channel is full, logs warning and returns
// Generates UUID and timestamp if not set
func (n *Nexus) PublishEvent(event Event) {
	// Generate UUID if not set
	if event.ID == "" {
		event.ID = uuid.New().String()
	}

	// Set timestamp if not set
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Non-blocking send to event handler channel
	select {
	case n.eventHandler <- event:
		// Successfully enqueued
	default:
		// Channel full - log warning
		n.log.Warn().
			Str("event_type", string(event.Type)).
			Str("event_id", event.ID).
			Msg("event handler channel full, event dropped")
	}
}

// eventHandlerLoop processes events sequentially
// This runs in a dedicated goroutine and ensures no concurrent event processing
func (n *Nexus) eventHandlerLoop() {
	defer n.wg.Done()

	for {
		select {
		case <-n.ctx.Done():
			n.log.Debug().Msg("event handler loop shutting down")
			return
		case event := <-n.eventHandler:
			n.processEvent(&event)
		}
	}
}

// processEvent handles a single event
func (n *Nexus) processEvent(event *Event) {
	// Recover from panics in event processing
	defer func() {
		if r := recover(); r != nil {
			n.log.Error().
				Interface("panic", r).
				Str("event_id", event.ID).
				Str("event_type", string(event.Type)).
				Msg("panic in event processing")
		}
	}()

	// Broadcast to subscribers
	n.eventBus.Publish(*event)

	// Save critical events to audit log
	if event.Critical {
		_ = n.auditLog.SaveEvent(context.Background(), event)
	}

	// Update state from event
	n.updateStateFromEvent(event)

	// Log event details
	n.log.Debug().
		Str("event_id", event.ID).
		Str("event_type", string(event.Type)).
		Str("source", event.Source).
		Bool("critical", event.Critical).
		Msg("event processed")
}

// updateStateFromEvent updates the state store based on event type
// Handles all 16 event types
func (n *Nexus) updateStateFromEvent(event *Event) {
	switch event.Type {
	// Wallet events
	case EventWalletAdded:
		if payload, ok := event.Payload.(WalletAddedPayload); ok {
			wallet := &WalletState{
				ID:      payload.ID,
				Address: payload.Address,
				Label:   payload.Label,
				Enabled: payload.Enabled,
				Primary: payload.Primary,
			}
			n.state.UpdateWallet(wallet)
		}

	case EventWalletRemoved:
		if payload, ok := event.Payload.(WalletRemovedPayload); ok {
			n.state.RemoveWallet(payload.ID)
		}

	case EventWalletChanged:
		if payload, ok := event.Payload.(WalletChangedPayload); ok {
			wallet := n.state.GetWallet(payload.ID)
			if wallet != nil {
				wallet.Enabled = payload.Enabled
				wallet.Primary = payload.Primary
				n.state.UpdateWallet(wallet)
			}
		}

	case EventBalanceUpdated:
		if payload, ok := event.Payload.(BalanceUpdatedPayload); ok {
			wallet := n.state.GetWallet(payload.WalletID)
			if wallet != nil {
				wallet.BalanceUSD = payload.BalanceUSD
				n.state.UpdateWallet(wallet)
			}
		}

	// Order events
	case EventOrderPlaced:
		if payload, ok := event.Payload.(OrderPlacedPayload); ok {
			order := &OrderState{
				ID:       payload.OrderID,
				WalletID: payload.WalletID,
				TokenID:  payload.TokenID,
				Side:     payload.Side,
				Price:    payload.Price,
				SizeUSD:  payload.SizeUSD,
				Status:   "pending",
			}
			n.state.UpdateOrder(order)
		}

	case EventOrderFilled:
		if payload, ok := event.Payload.(OrderFilledPayload); ok {
			order := n.state.GetOrder(payload.OrderID)
			if order != nil {
				order.FilledSize = payload.FilledSize
				order.Status = "filled"
				n.state.UpdateOrder(order)
			}
		}

	case EventOrderCanceled:
		if payload, ok := event.Payload.(OrderCanceledPayload); ok {
			order := n.state.GetOrder(payload.OrderID)
			if order != nil {
				order.Status = "canceled"
				n.state.UpdateOrder(order)
			}
		}

	// Position events
	case EventPositionOpened:
		if payload, ok := event.Payload.(PositionOpenedPayload); ok {
			position := &PositionState{
				ID:       payload.PositionID,
				WalletID: payload.WalletID,
				TokenID:  payload.TokenID,
				Outcome:  payload.Side,
				Size:     payload.Size,
			}
			n.state.UpdatePosition(position)
		}

	case EventPositionClosed:
		if payload, ok := event.Payload.(PositionClosedPayload); ok {
			n.state.RemovePosition(payload.PositionID)
		}

	// Strategy events
	case EventStrategyStarted:
		if payload, ok := event.Payload.(StrategyStartedPayload); ok {
			strategy := &StrategyState{
				Name:   payload.Strategy,
				Status: "running",
			}
			n.state.UpdateStrategy(strategy)
		}

	case EventStrategyStopped:
		if payload, ok := event.Payload.(StrategyStoppedPayload); ok {
			strategy := n.state.GetStrategy(payload.Strategy)
			if strategy != nil {
				strategy.Status = "stopped"
				n.state.UpdateStrategy(strategy)
			}
		}

	// Market events
	case EventMarketsUpdated:
		// Markets updated - no state update needed, just log
		n.log.Debug().Msg("markets updated")

	case EventPriceAlert:
		// Price alert - no state update needed, just log
		n.log.Debug().Msg("price alert triggered")

	// Strategy alert
	case EventStrategyAlert:
		// Strategy alert - no state update needed
		n.log.Debug().Msg("strategy alert triggered")

	// System events
	case EventConfigReloaded:
		n.log.Debug().Msg("config reloaded")

	case EventHealthSnapshot:
		if payload, ok := event.Payload.(HealthSnapshotPayload); ok {
			health := &HealthState{
				Name:    "system",
				Status:  payload.Status,
				Message: fmt.Sprintf("%v", payload.Data),
			}
			n.state.UpdateHealth(health)
		}
	}
}

// RegisterCommandHandler registers a handler for a specific command type
func (n *Nexus) RegisterCommandHandler(cmdType CommandType, handler CommandHandler) {
	n.cmdProcessor.RegisterHandler(cmdType, handler)
}

// SetCommandTimeout sets a custom timeout for a command type
func (n *Nexus) SetCommandTimeout(cmdType CommandType, timeout time.Duration) {
	n.cmdProcessor.SetCommandTimeout(cmdType, timeout)
}

// Subscribe creates a new subscription for events matching the pattern
func (n *Nexus) Subscribe(pattern string) <-chan Event {
	return n.eventBus.Subscribe(pattern)
}

// Unsubscribe removes a subscription
func (n *Nexus) Unsubscribe(pattern string, ch <-chan Event) {
	n.eventBus.Unsubscribe(pattern, ch)
}

// ExecuteCommand synchronously executes a command
func (n *Nexus) ExecuteCommand(ctx context.Context, cmd *Command) (*Command, error) {
	return n.cmdProcessor.Execute(ctx, cmd)
}

// ExecuteCommandAsync asynchronously executes a command
// Returns the command ID for status polling
func (n *Nexus) ExecuteCommandAsync(ctx context.Context, cmd *Command) (string, error) {
	return n.cmdProcessor.ExecuteAsync(ctx, cmd)
}

// GetCommandStatus returns the current status of a command by ID
func (n *Nexus) GetCommandStatus(cmdID string) (*Command, error) {
	return n.cmdProcessor.GetStatus(cmdID)
}

// GetState retrieves state by type
// Valid types: "wallets", "orders", "positions", "strategies", "markets", "health"
// Returns nil if unknown type
func (n *Nexus) GetState(stateType string) interface{} {
	switch stateType {
	case "wallets":
		return n.state.GetAllWallets()
	case "orders":
		return n.state.GetAllOrders()
	case "positions":
		return n.state.GetAllPositions()
	case "strategies":
		return n.state.GetAllStrategies()
	case "markets":
		return n.state.GetAllMarkets()
	case "health":
		return n.state.GetAllHealth()
	default:
		return nil
	}
}

// GetStats returns aggregated statistics from all components
func (n *Nexus) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"event_bus":           n.eventBus.Stats(),
		"command_processor":   n.cmdProcessor.Stats(),
		"state_store":         n.state.Snapshot(),
	}
}

// Health returns the system health status
func (n *Nexus) Health() map[string]interface{} {
	stats := n.GetStats()

	// Calculate overall health
	health := "healthy"
	if eventBusStats, ok := stats["event_bus"].(map[string]interface{}); ok {
		if dropped, ok := eventBusStats["dropped"].(uint64); ok && dropped > 100 {
			health = "degraded"
		}
	}

	return map[string]interface{}{
		"status": health,
		"stats":  stats,
	}
}

// GetRPCServer returns the RPC server instance (can be nil if not configured)
func (n *Nexus) GetRPCServer() RPCServer {
	return n.rpcServer
}

// Shutdown gracefully shuts down the Nexus coordinator
// Waits for all operations to complete with timeout
func (n *Nexus) Shutdown(ctx context.Context) error {
	n.log.Info().Msg("shutting down Nexus coordinator")

	// Close RPC server if exists
	if n.rpcServer != nil {
		if err := n.rpcServer.Close(); err != nil {
			n.log.Warn().Err(err).Msg("error closing RPC server")
		}
	}

	// Cancel context to signal shutdown
	n.cancel()

	// Close event bus
	n.eventBus.Close()

	// Close command processor
	if err := n.cmdProcessor.Close(); err != nil {
		n.log.Warn().Err(err).Msg("error closing command processor")
	}

	// Wait for event handler loop with timeout
	done := make(chan struct{})
	go func() {
		n.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		n.log.Info().Msg("Nexus coordinator shut down successfully")
		return nil
	case <-ctx.Done():
		return fmt.Errorf("shutdown timeout exceeded")
	}
}

// basicRPCServer is a placeholder RPC server for now
// Will be replaced by actual implementation in Task 6
type basicRPCServer struct{}

func (r *basicRPCServer) Close() error {
	return nil
}
