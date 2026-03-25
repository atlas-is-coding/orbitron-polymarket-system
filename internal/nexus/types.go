package nexus

import "time"

// EventType represents the type of an event
type EventType string

// Event type constants - Orders (3)
const (
	EventOrderPlaced   EventType = "order.placed"
	EventOrderFilled   EventType = "order.filled"
	EventOrderCanceled EventType = "order.canceled"
)

// Event type constants - Positions (2)
const (
	EventPositionOpened EventType = "position.opened"
	EventPositionClosed EventType = "position.closed"
)

// Event type constants - Wallets (4)
const (
	EventWalletAdded     EventType = "wallet.added"
	EventWalletRemoved   EventType = "wallet.removed"
	EventWalletChanged   EventType = "wallet.changed"
	EventBalanceUpdated  EventType = "balance.updated"
)

// Event type constants - Strategies (3)
const (
	EventStrategyStarted EventType = "strategy.started"
	EventStrategyStopped EventType = "strategy.stopped"
	EventStrategyAlert   EventType = "strategy.alert"
)

// Event type constants - Markets (2)
const (
	EventMarketsUpdated EventType = "markets.updated"
	EventPriceAlert     EventType = "price.alert"
)

// Event type constants - System (2)
const (
	EventConfigReloaded  EventType = "config.reloaded"
	EventHealthSnapshot  EventType = "health.snapshot"
)

// CommandType represents the type of a command
type CommandType string

// Command type constants - Orders (3)
const (
	CommandPlaceOrder      CommandType = "cmd.place_order"
	CommandCancelOrder     CommandType = "cmd.cancel_order"
	CommandCancelAllOrders CommandType = "cmd.cancel_all_orders"
)

// Command type constants - Wallets (4)
const (
	CommandAddWallet    CommandType = "cmd.add_wallet"
	CommandRemoveWallet CommandType = "cmd.remove_wallet"
	CommandToggleWallet CommandType = "cmd.toggle_wallet"
	CommandUpdateWallet CommandType = "cmd.update_wallet"
)

// Command type constants - Strategies (2)
const (
	CommandStartStrategy CommandType = "cmd.start_strategy"
	CommandStopStrategy  CommandType = "cmd.stop_strategy"
)

// Command type constants - Config (1)
const (
	CommandReloadConfig CommandType = "cmd.reload_config"
)

// CommandStatus represents the status of a command
type CommandStatus string

// Command status constants (5)
const (
	StatusPending    CommandStatus = "pending"
	StatusProcessing CommandStatus = "processing"
	StatusCompleted  CommandStatus = "completed"
	StatusFailed     CommandStatus = "failed"
	StatusTimedOut   CommandStatus = "timed_out"
)

// Event represents an event in the system
type Event struct {
	ID        string      `json:"id"`
	Type      EventType   `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Source    string      `json:"source"`
	Payload   interface{} `json:"payload"`
	Critical  bool        `json:"critical"`
}

// Command represents a command in the system
type Command struct {
	ID         string      `json:"id"`
	Type       CommandType `json:"type"`
	Timestamp  time.Time   `json:"timestamp"`
	SourceUI   string      `json:"source_ui"`
	Status     CommandStatus `json:"status"`
	Payload    interface{} `json:"payload"`
	Result     interface{} `json:"result,omitempty"`
	Error      string      `json:"error,omitempty"`
	Timeout    time.Duration `json:"timeout"`
	DeadlineAt time.Time   `json:"deadline_at,omitempty"`
}

// ============================================================
// Event Payloads
// ============================================================

// OrderPlacedPayload - order.placed event
type OrderPlacedPayload struct {
	OrderID  string  `json:"order_id"`
	TokenID  string  `json:"token_id"`
	WalletID string  `json:"wallet_id"`
	Side     string  `json:"side"`
	Price    float64 `json:"price"`
	SizeUSD  float64 `json:"size_usd"`
}

// OrderFilledPayload - order.filled event
type OrderFilledPayload struct {
	OrderID    string  `json:"order_id"`
	FilledSize float64 `json:"filled_size"`
}

// OrderCanceledPayload - order.canceled event
type OrderCanceledPayload struct {
	OrderID string `json:"order_id"`
}

// PositionOpenedPayload - position.opened event
type PositionOpenedPayload struct {
	PositionID string  `json:"position_id"`
	WalletID   string  `json:"wallet_id"`
	TokenID    string  `json:"token_id"`
	Side       string  `json:"side"`
	Size       float64 `json:"size"`
}

// PositionClosedPayload - position.closed event
type PositionClosedPayload struct {
	PositionID string `json:"position_id"`
}

// WalletAddedPayload - wallet.added event
type WalletAddedPayload struct {
	ID       string `json:"id"`
	Address  string `json:"address"`
	Label    string `json:"label"`
	Enabled  bool   `json:"enabled"`
	Primary  bool   `json:"primary"`
}

// WalletRemovedPayload - wallet.removed event
type WalletRemovedPayload struct {
	ID string `json:"id"`
}

// WalletChangedPayload - wallet.changed event
type WalletChangedPayload struct {
	ID      string `json:"id"`
	Enabled bool   `json:"enabled"`
	Primary bool   `json:"primary"`
}

// BalanceUpdatedPayload - balance.updated event
type BalanceUpdatedPayload struct {
	WalletID    string  `json:"wallet_id"`
	BalanceUSD  float64 `json:"balance_usd"`
	PreviousUSD float64 `json:"previous_usd"`
}

// StrategyStartedPayload - strategy.started event
type StrategyStartedPayload struct {
	Strategy string `json:"strategy"`
}

// StrategyStoppedPayload - strategy.stopped event
type StrategyStoppedPayload struct {
	Strategy string `json:"strategy"`
}

// StrategyAlertPayload - strategy.alert event
type StrategyAlertPayload struct {
	Strategy    string  `json:"strategy"`
	ConditionID string  `json:"condition_id"`
	Question    string  `json:"question"`
	Signal      string  `json:"signal"`
	Price       float64 `json:"price"`
	EdgePct     float64 `json:"edge_pct"`
	Reason      string  `json:"reason"`
}

// MarketsUpdatedPayload - markets.updated event
type MarketsUpdatedPayload struct {
	Count int `json:"count"`
}

// PriceAlertPayload - price.alert event
type PriceAlertPayload struct {
	ConditionID  string  `json:"condition_id"`
	Question     string  `json:"question"`
	Threshold    float64 `json:"threshold"`
	Direction    string  `json:"direction"`
	CurrentPrice float64 `json:"current_price"`
}

// ConfigReloadedPayload - config.reloaded event
type ConfigReloadedPayload struct {
	Timestamp time.Time `json:"timestamp"`
}

// HealthSnapshotPayload - health.snapshot event
type HealthSnapshotPayload struct {
	Status string                 `json:"status"`
	Data   map[string]interface{} `json:"data"`
}

// ============================================================
// Command Payloads
// ============================================================

// PlaceOrderPayload - place_order command
type PlaceOrderPayload struct {
	ConditionID string  `json:"condition_id"`
	WalletID    string  `json:"wallet_id"`
	Side        string  `json:"side"`
	Price       float64 `json:"price"`
	SizeUSD     float64 `json:"size_usd"`
	OrderType   string  `json:"order_type"`
}

// CancelOrderPayload - cancel_order command
type CancelOrderPayload struct {
	OrderID string `json:"order_id"`
}

// AddWalletPayload - add_wallet command
type AddWalletPayload struct {
	Address    string `json:"address"`
	Label      string `json:"label"`
	PrivateKey string `json:"private_key"`
}

// RemoveWalletPayload - remove_wallet command
type RemoveWalletPayload struct {
	WalletID string `json:"wallet_id"`
}

// ToggleWalletPayload - toggle_wallet command
type ToggleWalletPayload struct {
	WalletID string `json:"wallet_id"`
	Enabled  bool   `json:"enabled"`
}

// StartStrategyPayload - start_strategy command
type StartStrategyPayload struct {
	StrategyName string   `json:"strategy_name"`
	WalletIDs    []string `json:"wallet_ids"`
}

// StopStrategyPayload - stop_strategy command
type StopStrategyPayload struct {
	StrategyName string `json:"strategy_name"`
}

// ============================================================
// Validation Helpers
// ============================================================

// IsValidEventType validates an EventType
func IsValidEventType(t EventType) bool {
	switch t {
	// Orders
	case EventOrderPlaced, EventOrderFilled, EventOrderCanceled:
		return true
	// Positions
	case EventPositionOpened, EventPositionClosed:
		return true
	// Wallets
	case EventWalletAdded, EventWalletRemoved, EventWalletChanged, EventBalanceUpdated:
		return true
	// Strategies
	case EventStrategyStarted, EventStrategyStopped, EventStrategyAlert:
		return true
	// Markets
	case EventMarketsUpdated, EventPriceAlert:
		return true
	// System
	case EventConfigReloaded, EventHealthSnapshot:
		return true
	default:
		return false
	}
}

// IsValidCommandType validates a CommandType
func IsValidCommandType(t CommandType) bool {
	switch t {
	// Orders
	case CommandPlaceOrder, CommandCancelOrder, CommandCancelAllOrders:
		return true
	// Wallets
	case CommandAddWallet, CommandRemoveWallet, CommandToggleWallet, CommandUpdateWallet:
		return true
	// Strategies
	case CommandStartStrategy, CommandStopStrategy:
		return true
	// Config
	case CommandReloadConfig:
		return true
	default:
		return false
	}
}

// IsValidCommandStatus validates a CommandStatus
func IsValidCommandStatus(s CommandStatus) bool {
	switch s {
	case StatusPending, StatusProcessing, StatusCompleted, StatusFailed, StatusTimedOut:
		return true
	default:
		return false
	}
}
