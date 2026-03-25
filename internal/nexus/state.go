package nexus

import (
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// ============================================================
// State Structures
// ============================================================

// WalletState represents the state of a wallet
type WalletState struct {
	ID          string
	Address     string
	Label       string
	Enabled     bool
	Primary     bool
	BalanceUSD  float64
	PnLUSD      float64
	OpenOrders  int
	TotalTrades int
	UpdatedAt   time.Time
}

// OrderState represents the state of an order
type OrderState struct {
	ID         string
	WalletID   string
	TokenID    string
	Side       string
	Price      float64
	SizeUSD    float64
	Status     string
	FilledSize float64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// PositionState represents the state of a position
type PositionState struct {
	ID       string
	WalletID string
	TokenID  string
	Market   string
	Outcome  string
	Size     float64
	Value    float64
	UpdatedAt time.Time
}

// StrategyState represents the state of a strategy
type StrategyState struct {
	Name       string
	Status     string
	WalletID   string
	LastSignal time.Time
	LastAction string
	UpdatedAt  time.Time
}

// MarketState represents the state of a market
type MarketState struct {
	ConditionID string
	Question    string
	YesPrice    float64
	NoPrice     float64
	Volume24h   float64
	UpdatedAt   time.Time
}

// HealthState represents the health status of a component
type HealthState struct {
	Name      string
	Status    string
	Message   string
	UpdatedAt time.Time
}

// ============================================================
// StateStore
// ============================================================

// StateStore maintains in-memory cache of all system state
type StateStore struct {
	mu         sync.RWMutex
	wallets    map[string]*WalletState
	orders     map[string]*OrderState
	positions  map[string]*PositionState
	strategies map[string]*StrategyState
	markets    map[string]*MarketState
	health     map[string]*HealthState
	log        zerolog.Logger
}

// NewStateStore creates a new StateStore instance
func NewStateStore(log zerolog.Logger) *StateStore {
	return &StateStore{
		wallets:    make(map[string]*WalletState),
		orders:     make(map[string]*OrderState),
		positions:  make(map[string]*PositionState),
		strategies: make(map[string]*StrategyState),
		markets:    make(map[string]*MarketState),
		health:     make(map[string]*HealthState),
		log:        log,
	}
}

// ============================================================
// Wallet Operations
// ============================================================

// UpdateWallet updates or creates a wallet in the store
func (s *StateStore) UpdateWallet(wallet *WalletState) {
	if wallet == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	wallet.UpdatedAt = time.Now()
	s.wallets[wallet.ID] = wallet
}

// GetWallet retrieves a wallet by ID
func (s *StateStore) GetWallet(id string) *WalletState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.wallets[id]
}

// GetAllWallets returns all wallets
func (s *StateStore) GetAllWallets() []*WalletState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	wallets := make([]*WalletState, 0, len(s.wallets))
	for _, w := range s.wallets {
		wallets = append(wallets, w)
	}
	return wallets
}

// RemoveWallet removes a wallet by ID
func (s *StateStore) RemoveWallet(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.wallets, id)
}

// ============================================================
// Order Operations
// ============================================================

// UpdateOrder updates or creates an order in the store
func (s *StateStore) UpdateOrder(order *OrderState) {
	if order == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.orders[order.ID]; !exists {
		// First creation: set CreatedAt
		order.CreatedAt = time.Now()
	}
	order.UpdatedAt = time.Now()
	s.orders[order.ID] = order
}

// GetOrder retrieves an order by ID
func (s *StateStore) GetOrder(id string) *OrderState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.orders[id]
}

// GetAllOrders returns all orders
func (s *StateStore) GetAllOrders() []*OrderState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	orders := make([]*OrderState, 0, len(s.orders))
	for _, o := range s.orders {
		orders = append(orders, o)
	}
	return orders
}

// GetOrdersByWallet returns all orders for a specific wallet
func (s *StateStore) GetOrdersByWallet(walletID string) []*OrderState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	orders := make([]*OrderState, 0)
	for _, o := range s.orders {
		if o.WalletID == walletID {
			orders = append(orders, o)
		}
	}
	return orders
}

// RemoveOrder removes an order by ID
func (s *StateStore) RemoveOrder(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.orders, id)
}

// ============================================================
// Position Operations
// ============================================================

// UpdatePosition updates or creates a position in the store
func (s *StateStore) UpdatePosition(pos *PositionState) {
	if pos == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	pos.UpdatedAt = time.Now()
	s.positions[pos.ID] = pos
}

// GetPosition retrieves a position by ID
func (s *StateStore) GetPosition(id string) *PositionState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.positions[id]
}

// GetAllPositions returns all positions
func (s *StateStore) GetAllPositions() []*PositionState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	positions := make([]*PositionState, 0, len(s.positions))
	for _, p := range s.positions {
		positions = append(positions, p)
	}
	return positions
}

// GetPositionsByWallet returns all positions for a specific wallet
func (s *StateStore) GetPositionsByWallet(walletID string) []*PositionState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	positions := make([]*PositionState, 0)
	for _, p := range s.positions {
		if p.WalletID == walletID {
			positions = append(positions, p)
		}
	}
	return positions
}

// RemovePosition removes a position by ID
func (s *StateStore) RemovePosition(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.positions, id)
}

// ============================================================
// Strategy Operations
// ============================================================

// UpdateStrategy updates or creates a strategy in the store
func (s *StateStore) UpdateStrategy(strategy *StrategyState) {
	if strategy == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	strategy.UpdatedAt = time.Now()
	s.strategies[strategy.Name] = strategy
}

// GetStrategy retrieves a strategy by name
func (s *StateStore) GetStrategy(name string) *StrategyState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.strategies[name]
}

// GetAllStrategies returns all strategies
func (s *StateStore) GetAllStrategies() []*StrategyState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	strategies := make([]*StrategyState, 0, len(s.strategies))
	for _, s := range s.strategies {
		strategies = append(strategies, s)
	}
	return strategies
}

// ============================================================
// Market Operations
// ============================================================

// UpdateMarket updates or creates a market in the store
func (s *StateStore) UpdateMarket(market *MarketState) {
	if market == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	market.UpdatedAt = time.Now()
	s.markets[market.ConditionID] = market
}

// GetMarket retrieves a market by ConditionID
func (s *StateStore) GetMarket(conditionID string) *MarketState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.markets[conditionID]
}

// GetAllMarkets returns all markets
func (s *StateStore) GetAllMarkets() []*MarketState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	markets := make([]*MarketState, 0, len(s.markets))
	for _, m := range s.markets {
		markets = append(markets, m)
	}
	return markets
}

// ============================================================
// Health Operations
// ============================================================

// UpdateHealth updates or creates a health status in the store
func (s *StateStore) UpdateHealth(health *HealthState) {
	if health == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	health.UpdatedAt = time.Now()
	s.health[health.Name] = health
}

// GetHealth retrieves a health status by name
func (s *StateStore) GetHealth(name string) *HealthState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.health[name]
}

// GetAllHealth returns all health statuses
func (s *StateStore) GetAllHealth() []*HealthState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	health := make([]*HealthState, 0, len(s.health))
	for _, h := range s.health {
		health = append(health, h)
	}
	return health
}

// ============================================================
// Snapshot
// ============================================================

// Snapshot returns a snapshot of all counts
func (s *StateStore) Snapshot() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"wallets":    len(s.wallets),
		"orders":     len(s.orders),
		"positions":  len(s.positions),
		"strategies": len(s.strategies),
		"markets":    len(s.markets),
		"health":     len(s.health),
	}
}
