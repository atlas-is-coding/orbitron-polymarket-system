# Nexus: Centralized Data Synchronization System

Nexus is the central coordinator for all data synchronization in polytrade-bot. It provides broadcast messaging, request-response commands, and consistent state management for WebUI, TUI, and TelegramUI.

## Architecture

```
Nexus
├── EventBus (Pub/Sub)          - Broadcast events with glob pattern matching
├── CommandProcessor (RPC)      - Sync/async command execution with timeouts
├── StateStore (Cache)          - In-memory state for fast queries
├── AuditLog (Persistence)      - Critical event/command logging
└── RPCServer (Remote UIs)      - JSON RPC + WebSocket for WebUI/TelegramUI
```

## Core Components

### EventBus
- **Publish**: `nexus.PublishEvent(event)`
- **Subscribe**: `eventChan := nexus.Subscribe("order.*")`
- **Pattern Matching**: `"order.*"`, `"balance.*"`, `"*"` (wildcard)
- **Non-blocking**: Drops events if subscriber is slow

### CommandProcessor
- **Sync Execution**: `result, err := nexus.ExecuteCommand(ctx, cmd)`
- **Async Execution**: `cmdID, err := nexus.ExecuteCommandAsync(ctx, cmd)`
- **Status Polling**: `cmd, err := nexus.GetCommandStatus(cmdID)`
- **Timeouts**: 30s default, configurable per command type

### StateStore
- **In-Memory Cache**: Fast read-only access to current state
- **Auto-Synced**: Updated by event handler loop
- **State Types**: Wallets, Orders, Positions, Strategies, Markets, Health
- **Query Methods**: `GetAllWallets()`, `GetOrdersByWallet()`, etc.

### RPCServer
- **JSON RPC Endpoint**: `/api/rpc` (HTTP POST)
- **WebSocket Streaming**: `/api/events` (subscribe to patterns)
- **Methods**: ExecuteCommand, ExecuteCommandAsync, GetCommandStatus, GetState, GetStats, Health

## Usage Examples

### Publish Events
```go
nexus.PublishEvent(nexus.Event{
    Type:     nexus.EventOrderPlaced,
    Source:   "trades_monitor",
    Critical: true,  // Persist to audit log
    Payload: &nexus.OrderPlacedPayload{
        OrderID:  "ord_123",
        TokenID:  "tok_1",
        WalletID: "w1",
        Side:     "YES",
        Price:    0.5,
        SizeUSD:  100,
    },
})
```

### Subscribe to Events (TUI)
```go
eventChan := nexus.Subscribe("order.*")
for event := range eventChan {
    switch event.Type {
    case nexus.EventOrderPlaced:
        orders := nexus.GetState("orders").([]*nexus.OrderState)
        // Update UI with orders
    }
}
```

### Execute Commands (WebUI)
```go
// Async command with tracking
cmdID, err := nexus.ExecuteCommandAsync(ctx, &nexus.Command{
    Type: nexus.CommandPlaceOrder,
    Payload: &nexus.PlaceOrderPayload{...},
})

// Poll for result
cmd, _ := nexus.GetCommandStatus(cmdID)
if cmd.Status == nexus.StatusCompleted {
    result := cmd.Result
}
```

### RPC from JavaScript
```javascript
// Sync command
const res = await fetch('/api/rpc', {
    method: 'POST',
    body: JSON.stringify({
        method: 'ExecuteCommand',
        params: {type: 'cancel.order', payload: {order_id: 'ord_1'}},
        id: 'req_1'
    })
});

// Async event streaming
const ws = new WebSocket('ws://localhost:9000/api/events');
ws.send(JSON.stringify({pattern: 'order.*'}));
ws.onmessage = (e) => {
    const event = JSON.parse(e.data);
    // Update UI
};
```

## Event Types

**Orders**: placed, filled, canceled
**Positions**: opened, closed
**Wallets**: added, removed, changed, balance_updated
**Strategies**: started, stopped, alert
**Markets**: updated, price_alert
**System**: config_reloaded, health_snapshot

## Command Types

**Orders**: place, cancel, cancel_all
**Wallets**: add, remove, toggle, update
**Strategies**: start, stop
**Config**: reload

## Integration

Each subsystem initializes with Nexus:

```go
// Create Nexus
nex, _ := nexus.NewNexus(auditLog, 9000, log)

// Subsystems
monitor := monitor.NewTradesMonitor(nex, cfg)
tuiApp := tui.NewAppModel(nex, cfg)
webServer := webui.NewServer(cfg, cfgPath, password, nex)
telegramBot := telegrambot.NewBot(nex, cfg)
```

## Performance

- **EventBus**: <1ms publish, 10,000+ events/sec
- **CommandProcessor**: Handler-dependent (typically 100-500ms)
- **StateStore**: <1ms query, 100,000+ queries/sec
- **Memory**: ~10MB for 1000 wallets + 10000 orders

## Monitoring

### Statistics
```go
stats := nexus.GetStats()
// {
//   "event_bus": {"sent": 1234, "dropped": 0, "subscribers": 3},
//   "command_proc": {"in_flight": 2, "queue_len": 5, "workers": 4},
//   "state_store": {"wallets": 5, "orders": 23, ...},
//   "rpc_clients": 2
// }
```

### Health Check
```go
health := nexus.Health()
// {"status": "healthy", "timestamp": "...", "event_bus_drops": 0, ...}
```

## Testing

```bash
# Run all Nexus tests
go test ./internal/nexus -v

# Run specific test
go test ./internal/nexus -v -run TestFullNexusFlow

# With coverage
go test ./internal/nexus -cover
```

## Shutdown

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
if err := nexus.Shutdown(ctx); err != nil {
    log.Error().Err(err).Msg("Shutdown error")
}
```

Nexus gracefully closes all components, waits for goroutines, and persists critical events to audit log.
