package tui

import (
	"sync"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/atlasdev/orbitron/internal/health"
)

// Nexus is a thread-safe, centralized state manager for the Polytrade Bot.
// It maintains the "Source of Truth" for all UI components (TUI and WebUI).
type Nexus struct {
	mu         sync.RWMutex
	wallets    map[string]WalletStatsMsg
	strategies []StrategyRow
	subsystems map[string]bool
	orders     []OrderRow
	positions  []PositionRow
	health     health.HealthSnapshot
}

// NewNexus creates a new Nexus instance.
func NewNexus() *Nexus {
	return &Nexus{
		wallets:    make(map[string]WalletStatsMsg),
		subsystems: make(map[string]bool),
		strategies: make([]StrategyRow, 0),
		orders:     make([]OrderRow, 0),
		positions:  make([]PositionRow, 0),
	}
}

// Handle updates the internal state based on the message type.
func (n *Nexus) Handle(msg tea.Msg) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// fmt.Printf("[NEXUS] Handling %T\n", msg) // Manual debug if needed

	switch m := msg.(type) {
	case WalletAddedMsg:
		if _, ok := n.wallets[m.ID]; !ok {
			n.wallets[m.ID] = WalletStatsMsg{
				ID:      m.ID,
				Address: m.Address,
				Label:   m.Label,
				Enabled: m.Enabled,
				Primary: m.Primary,
			}
		}
	case WalletStatsMsg:
		n.wallets[m.ID] = m
	case StrategiesUpdateMsg:
		n.strategies = m.Rows
	case SubsystemStatusMsg:
		n.subsystems[m.Name] = m.Active
	case OrdersUpdateMsg:
		n.orders = m.Rows
	case PositionsUpdateMsg:
		n.positions = m.Rows
	case HealthSnapshotMsg:
		n.health = m.Snapshot
	case WalletRemovedMsg:
		delete(n.wallets, m.ID)
	case WalletChangedMsg:
		if w, ok := n.wallets[m.ID]; ok {
			w.Enabled = m.Enabled
			w.Primary = m.Primary
			n.wallets[m.ID] = w
		}
		if m.Primary {
			for id, w := range n.wallets {
				if id != m.ID && w.Primary {
					w.Primary = false
					n.wallets[id] = w
				}
			}
		}
	}
}

// GetTotals returns the aggregated BalanceUSD and PnLUSD across all wallets.
func (n *Nexus) GetTotals() (float64, float64) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	var totalBalance, totalPnL float64
	for _, w := range n.wallets {
		totalBalance += w.BalanceUSD
		totalPnL += w.PnLUSD
	}
	return totalBalance, totalPnL
}

// Snapshot returns a deep copy of the current state for UI initialization.
func (n *Nexus) Snapshot() map[string]any {
	n.mu.RLock()
	defer n.mu.RUnlock()

	// Deep copy wallets
	walletsCopy := make([]WalletStatsMsg, 0, len(n.wallets))
	for _, w := range n.wallets {
		walletsCopy = append(walletsCopy, w)
	}

	// Deep copy strategies
	strategiesCopy := make([]StrategyRow, len(n.strategies))
	copy(strategiesCopy, n.strategies)

	// Deep copy subsystems
	subsystemsCopy := make(map[string]bool, len(n.subsystems))
	for k, v := range n.subsystems {
		subsystemsCopy[k] = v
	}

	// Deep copy orders
	ordersCopy := make([]OrderRow, len(n.orders))
	copy(ordersCopy, n.orders)

	// Deep copy positions
	positionsCopy := make([]PositionRow, len(n.positions))
	copy(positionsCopy, n.positions)

	totalBalance, totalPnL := n.GetTotals()

	return map[string]any{
		"balance":    totalBalance,
		"pnl":        totalPnL,
		"wallets":    walletsCopy,
		"strategies": strategiesCopy,
		"subsystems": subsystemsCopy,
		"orders":     ordersCopy,
		"positions":  positionsCopy,
		"health":     n.health,
	}
}
