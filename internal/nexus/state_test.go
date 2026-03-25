package nexus

import (
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestStateStoreWalletOperations(t *testing.T) {
	log := zerolog.New(nil)
	store := NewStateStore(log)

	// Test UpdateWallet
	wallet := &WalletState{
		ID:         "wallet1",
		Address:    "0x123",
		Label:      "Main",
		Enabled:    true,
		Primary:    true,
		BalanceUSD: 1000.0,
		PnLUSD:     50.0,
	}
	store.UpdateWallet(wallet)

	// Test GetWallet
	retrieved := store.GetWallet("wallet1")
	if retrieved == nil {
		t.Fatal("GetWallet returned nil")
	}
	if retrieved.ID != "wallet1" {
		t.Errorf("expected ID wallet1, got %s", retrieved.ID)
	}
	if retrieved.Address != "0x123" {
		t.Errorf("expected address 0x123, got %s", retrieved.Address)
	}
	if retrieved.BalanceUSD != 1000.0 {
		t.Errorf("expected balance 1000.0, got %f", retrieved.BalanceUSD)
	}

	// Test GetAllWallets
	wallet2 := &WalletState{
		ID:      "wallet2",
		Address: "0x456",
		Label:   "Secondary",
		Enabled: true,
		Primary: false,
	}
	store.UpdateWallet(wallet2)

	allWallets := store.GetAllWallets()
	if len(allWallets) != 2 {
		t.Errorf("expected 2 wallets, got %d", len(allWallets))
	}

	// Test RemoveWallet
	store.RemoveWallet("wallet1")
	retrieved = store.GetWallet("wallet1")
	if retrieved != nil {
		t.Fatal("RemoveWallet failed: wallet still exists")
	}

	allWallets = store.GetAllWallets()
	if len(allWallets) != 1 {
		t.Errorf("expected 1 wallet after removal, got %d", len(allWallets))
	}

	// Test UpdatedAt is set
	if wallet2.UpdatedAt.IsZero() {
		t.Error("UpdatedAt not set on wallet")
	}
}

func TestStateStoreOrderOperations(t *testing.T) {
	log := zerolog.New(nil)
	store := NewStateStore(log)

	// Test UpdateOrder
	order := &OrderState{
		ID:       "order1",
		WalletID: "wallet1",
		TokenID:  "token1",
		Side:     "BUY",
		Price:    0.65,
		SizeUSD:  100.0,
		Status:   "OPEN",
	}
	store.UpdateOrder(order)

	// Test GetOrder
	retrieved := store.GetOrder("order1")
	if retrieved == nil {
		t.Fatal("GetOrder returned nil")
	}
	if retrieved.ID != "order1" {
		t.Errorf("expected ID order1, got %s", retrieved.ID)
	}
	if retrieved.WalletID != "wallet1" {
		t.Errorf("expected wallet1, got %s", retrieved.WalletID)
	}

	// Test GetAllOrders
	order2 := &OrderState{
		ID:       "order2",
		WalletID: "wallet2",
		TokenID:  "token1",
		Side:     "SELL",
		Price:    0.35,
		SizeUSD:  50.0,
		Status:   "OPEN",
	}
	store.UpdateOrder(order2)

	allOrders := store.GetAllOrders()
	if len(allOrders) != 2 {
		t.Errorf("expected 2 orders, got %d", len(allOrders))
	}

	// Test GetOrdersByWallet
	wallet1Orders := store.GetOrdersByWallet("wallet1")
	if len(wallet1Orders) != 1 {
		t.Errorf("expected 1 order for wallet1, got %d", len(wallet1Orders))
	}
	if wallet1Orders[0].ID != "order1" {
		t.Errorf("expected order1, got %s", wallet1Orders[0].ID)
	}

	wallet2Orders := store.GetOrdersByWallet("wallet2")
	if len(wallet2Orders) != 1 {
		t.Errorf("expected 1 order for wallet2, got %d", len(wallet2Orders))
	}

	// Test RemoveOrder
	store.RemoveOrder("order1")
	retrieved = store.GetOrder("order1")
	if retrieved != nil {
		t.Fatal("RemoveOrder failed: order still exists")
	}

	allOrders = store.GetAllOrders()
	if len(allOrders) != 1 {
		t.Errorf("expected 1 order after removal, got %d", len(allOrders))
	}

	// Test CreatedAt is set
	if order.CreatedAt.IsZero() {
		t.Error("CreatedAt not set on order")
	}
}

func TestStateStorePositionOperations(t *testing.T) {
	log := zerolog.New(nil)
	store := NewStateStore(log)

	// Test UpdatePosition
	pos := &PositionState{
		ID:       "pos1",
		WalletID: "wallet1",
		TokenID:  "token1",
		Market:   "market1",
		Outcome:  "YES",
		Size:     100.0,
		Value:    65.0,
	}
	store.UpdatePosition(pos)

	// Test GetPosition
	retrieved := store.GetPosition("pos1")
	if retrieved == nil {
		t.Fatal("GetPosition returned nil")
	}
	if retrieved.ID != "pos1" {
		t.Errorf("expected ID pos1, got %s", retrieved.ID)
	}

	// Test GetAllPositions
	pos2 := &PositionState{
		ID:       "pos2",
		WalletID: "wallet2",
		TokenID:  "token2",
		Market:   "market2",
		Outcome:  "NO",
		Size:     50.0,
		Value:    17.5,
	}
	store.UpdatePosition(pos2)

	allPos := store.GetAllPositions()
	if len(allPos) != 2 {
		t.Errorf("expected 2 positions, got %d", len(allPos))
	}

	// Test GetPositionsByWallet
	wallet1Pos := store.GetPositionsByWallet("wallet1")
	if len(wallet1Pos) != 1 {
		t.Errorf("expected 1 position for wallet1, got %d", len(wallet1Pos))
	}
	if wallet1Pos[0].ID != "pos1" {
		t.Errorf("expected pos1, got %s", wallet1Pos[0].ID)
	}

	// Test RemovePosition
	store.RemovePosition("pos1")
	retrieved = store.GetPosition("pos1")
	if retrieved != nil {
		t.Fatal("RemovePosition failed: position still exists")
	}

	allPos = store.GetAllPositions()
	if len(allPos) != 1 {
		t.Errorf("expected 1 position after removal, got %d", len(allPos))
	}

	// Test UpdatedAt is set
	if pos.UpdatedAt.IsZero() {
		t.Error("UpdatedAt not set on position")
	}
}

func TestStateStoreStrategyOperations(t *testing.T) {
	log := zerolog.New(nil)
	store := NewStateStore(log)

	// Test UpdateStrategy
	strategy := &StrategyState{
		Name:        "strategy1",
		Status:      "RUNNING",
		WalletID:    "wallet1",
		LastSignal:  time.Now(),
		LastAction:  "PLACE_ORDER",
	}
	store.UpdateStrategy(strategy)

	// Test GetStrategy
	retrieved := store.GetStrategy("strategy1")
	if retrieved == nil {
		t.Fatal("GetStrategy returned nil")
	}
	if retrieved.Name != "strategy1" {
		t.Errorf("expected name strategy1, got %s", retrieved.Name)
	}
	if retrieved.Status != "RUNNING" {
		t.Errorf("expected status RUNNING, got %s", retrieved.Status)
	}

	// Test GetAllStrategies
	strategy2 := &StrategyState{
		Name:     "strategy2",
		Status:   "STOPPED",
		WalletID: "wallet2",
	}
	store.UpdateStrategy(strategy2)

	allStrategies := store.GetAllStrategies()
	if len(allStrategies) != 2 {
		t.Errorf("expected 2 strategies, got %d", len(allStrategies))
	}

	// Test UpdatedAt is set
	if strategy.UpdatedAt.IsZero() {
		t.Error("UpdatedAt not set on strategy")
	}
}

func TestStateStoreMarketOperations(t *testing.T) {
	log := zerolog.New(nil)
	store := NewStateStore(log)

	// Test UpdateMarket
	market := &MarketState{
		ConditionID: "cond1",
		Question:    "Will BTC reach $100k?",
		YesPrice:    0.75,
		NoPrice:     0.25,
		Volume24h:   1000000.0,
	}
	store.UpdateMarket(market)

	// Test GetMarket
	retrieved := store.GetMarket("cond1")
	if retrieved == nil {
		t.Fatal("GetMarket returned nil")
	}
	if retrieved.ConditionID != "cond1" {
		t.Errorf("expected condition cond1, got %s", retrieved.ConditionID)
	}
	if retrieved.YesPrice != 0.75 {
		t.Errorf("expected price 0.75, got %f", retrieved.YesPrice)
	}

	// Test GetAllMarkets
	market2 := &MarketState{
		ConditionID: "cond2",
		Question:    "Will ETH reach $5k?",
		YesPrice:    0.60,
		NoPrice:     0.40,
		Volume24h:   500000.0,
	}
	store.UpdateMarket(market2)

	allMarkets := store.GetAllMarkets()
	if len(allMarkets) != 2 {
		t.Errorf("expected 2 markets, got %d", len(allMarkets))
	}

	// Test UpdatedAt is set
	if market.UpdatedAt.IsZero() {
		t.Error("UpdatedAt not set on market")
	}
}

func TestStateStoreHealthOperations(t *testing.T) {
	log := zerolog.New(nil)
	store := NewStateStore(log)

	// Test UpdateHealth
	health := &HealthState{
		Name:    "api_health",
		Status:  "OK",
		Message: "All systems operational",
	}
	store.UpdateHealth(health)

	// Test GetHealth
	retrieved := store.GetHealth("api_health")
	if retrieved == nil {
		t.Fatal("GetHealth returned nil")
	}
	if retrieved.Name != "api_health" {
		t.Errorf("expected name api_health, got %s", retrieved.Name)
	}
	if retrieved.Status != "OK" {
		t.Errorf("expected status OK, got %s", retrieved.Status)
	}

	// Test GetAllHealth
	health2 := &HealthState{
		Name:    "db_health",
		Status:  "OK",
		Message: "Database connected",
	}
	store.UpdateHealth(health2)

	allHealth := store.GetAllHealth()
	if len(allHealth) != 2 {
		t.Errorf("expected 2 health items, got %d", len(allHealth))
	}

	// Test UpdatedAt is set
	if health.UpdatedAt.IsZero() {
		t.Error("UpdatedAt not set on health")
	}
}

func TestStateStoreThreadSafety(t *testing.T) {
	log := zerolog.New(nil)
	store := NewStateStore(log)

	var wg sync.WaitGroup
	errChan := make(chan error, 200)

	// 100 concurrent writes
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			wallet := &WalletState{
				ID:         "wallet" + string(rune(idx)),
				Address:    "0x" + string(rune(idx)),
				Label:      "wallet" + string(rune(idx)),
				Enabled:    true,
				Primary:    idx == 0,
				BalanceUSD: float64(idx * 100),
			}
			store.UpdateWallet(wallet)
		}(i)
	}

	// 100 concurrent reads
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			wallets := store.GetAllWallets()
			if wallets == nil {
				errChan <- nil // valid case during setup
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	// Verify no panics and data integrity
	allWallets := store.GetAllWallets()
	if len(allWallets) < 100 {
		t.Errorf("expected at least 100 wallets, got %d", len(allWallets))
	}
}

func TestStateStoreSnapshot(t *testing.T) {
	log := zerolog.New(nil)
	store := NewStateStore(log)

	// Add various items
	for i := 0; i < 5; i++ {
		wallet := &WalletState{
			ID:      "wallet" + string(rune(i)),
			Address: "0x" + string(rune(i)),
			Label:   "wallet" + string(rune(i)),
		}
		store.UpdateWallet(wallet)
	}

	for i := 0; i < 10; i++ {
		order := &OrderState{
			ID:       "order" + string(rune(i)),
			WalletID: "wallet" + string(rune(i%5)),
			TokenID:  "token" + string(rune(i)),
		}
		store.UpdateOrder(order)
	}

	for i := 0; i < 3; i++ {
		pos := &PositionState{
			ID:       "pos" + string(rune(i)),
			WalletID: "wallet" + string(rune(i)),
			TokenID:  "token" + string(rune(i)),
		}
		store.UpdatePosition(pos)
	}

	for i := 0; i < 2; i++ {
		strategy := &StrategyState{
			Name:     "strategy" + string(rune(i)),
			WalletID: "wallet" + string(rune(i)),
		}
		store.UpdateStrategy(strategy)
	}

	for i := 0; i < 4; i++ {
		market := &MarketState{
			ConditionID: "cond" + string(rune(i)),
			Question:    "Question " + string(rune(i)),
		}
		store.UpdateMarket(market)
	}

	for i := 0; i < 3; i++ {
		health := &HealthState{
			Name:   "health" + string(rune(i)),
			Status: "OK",
		}
		store.UpdateHealth(health)
	}

	// Get snapshot
	snapshot := store.Snapshot()

	// Verify counts
	if walletCount, ok := snapshot["wallets"].(int); !ok || walletCount != 5 {
		t.Errorf("expected 5 wallets in snapshot, got %v", snapshot["wallets"])
	}

	if orderCount, ok := snapshot["orders"].(int); !ok || orderCount != 10 {
		t.Errorf("expected 10 orders in snapshot, got %v", snapshot["orders"])
	}

	if posCount, ok := snapshot["positions"].(int); !ok || posCount != 3 {
		t.Errorf("expected 3 positions in snapshot, got %v", snapshot["positions"])
	}

	if stratCount, ok := snapshot["strategies"].(int); !ok || stratCount != 2 {
		t.Errorf("expected 2 strategies in snapshot, got %v", snapshot["strategies"])
	}

	if marketCount, ok := snapshot["markets"].(int); !ok || marketCount != 4 {
		t.Errorf("expected 4 markets in snapshot, got %v", snapshot["markets"])
	}

	if healthCount, ok := snapshot["health"].(int); !ok || healthCount != 3 {
		t.Errorf("expected 3 health items in snapshot, got %v", snapshot["health"])
	}
}

func TestStateStoreConcurrentWalletAndOrderOps(t *testing.T) {
	log := zerolog.New(nil)
	store := NewStateStore(log)

	var wg sync.WaitGroup

	// Concurrent wallet and order operations
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			wallet := &WalletState{
				ID:      "wallet" + string(rune(idx)),
				Address: "0x" + string(rune(idx)),
				Label:   "wallet" + string(rune(idx)),
			}
			store.UpdateWallet(wallet)

			order := &OrderState{
				ID:       "order" + string(rune(idx)),
				WalletID: "wallet" + string(rune(idx)),
				TokenID:  "token" + string(rune(idx)),
			}
			store.UpdateOrder(order)

			// Read while writing
			_ = store.GetWallet("wallet" + string(rune(idx)))
			_ = store.GetOrder("order" + string(rune(idx)))
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = store.GetAllWallets()
			_ = store.GetAllOrders()
			_ = store.GetAllPositions()
		}()
	}

	wg.Wait()

	// Verify integrity
	allWallets := store.GetAllWallets()
	allOrders := store.GetAllOrders()

	if len(allWallets) < 50 {
		t.Errorf("expected at least 50 wallets, got %d", len(allWallets))
	}
	if len(allOrders) < 50 {
		t.Errorf("expected at least 50 orders, got %d", len(allOrders))
	}
}
