package wallet

import (
	"fmt"
	"sync"

	"github.com/atlasdev/polytrade-bot/internal/config"
	"github.com/atlasdev/polytrade-bot/internal/tui"
)

// Manager manages the set of active and inactive wallet instances.
// All public methods are thread-safe.
type Manager struct {
	mu        sync.RWMutex
	instances []*WalletInstance
	bus       *tui.EventBus
}

// NewManager creates a Manager. bus may be nil (e.g., in tests or headless mode).
func NewManager(bus *tui.EventBus) *Manager {
	return &Manager{bus: bus}
}

// AddInactive adds a wallet without starting any subsystems.
// Used for disabled wallets and in tests.
func (m *Manager) AddInactive(cfg config.WalletConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.instances = append(m.instances, &WalletInstance{
		Cfg:   cfg,
		Stats: &WalletStats{},
	})
}

// AddActive adds a fully initialised wallet instance and broadcasts WalletAddedMsg.
func (m *Manager) AddActive(inst *WalletInstance) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.instances = append(m.instances, inst)
	if m.bus != nil {
		m.bus.Send(tui.WalletAddedMsg{ID: inst.Cfg.ID, Label: inst.Cfg.Label, Enabled: inst.Cfg.Enabled})
	}
}

// WalletIDs returns a snapshot of all wallet IDs (both active and inactive).
// Implements tui.WalletProvider.
func (m *Manager) WalletIDs() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ids := make([]string, len(m.instances))
	for i, w := range m.instances {
		ids[i] = w.Cfg.ID
	}
	return ids
}

// Wallets returns a snapshot slice of all wallet instances (both active and inactive).
func (m *Manager) Wallets() []*WalletInstance {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*WalletInstance, len(m.instances))
	copy(out, m.instances)
	return out
}

// Get returns the wallet instance with the given ID, or (nil, false) if not found.
func (m *Manager) Get(id string) (*WalletInstance, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, w := range m.instances {
		if w.Cfg.ID == id {
			return w, true
		}
	}
	return nil, false
}

// UpdateLabel updates the display label of a wallet in-memory.
// Callers are responsible for persisting the change via config.Save.
func (m *Manager) UpdateLabel(id, label string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, w := range m.instances {
		if w.Cfg.ID == id {
			w.Cfg.Label = label
			return nil
		}
	}
	return fmt.Errorf("wallet %q not found", id)
}

// Remove stops a wallet (graceful drain) and removes it from the manager.
// Callers are responsible for persisting the change via config.Save.
func (m *Manager) Remove(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, w := range m.instances {
		if w.Cfg.ID == id {
			w.Stop()
			m.instances = append(m.instances[:i], m.instances[i+1:]...)
			if m.bus != nil {
				m.bus.Send(tui.WalletRemovedMsg{ID: id})
			}
			return nil
		}
	}
	return fmt.Errorf("wallet %q not found", id)
}

// WalletLabel returns the display label of the wallet with the given ID.
// Returns an empty string if not found.
// Implements tui.WalletProvider.
func (m *Manager) WalletLabel(id string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, w := range m.instances {
		if w.Cfg.ID == id {
			return w.Cfg.Label
		}
	}
	return ""
}

// WalletAddress returns the derived Ethereum address of the wallet with the given ID.
// Returns an empty string if not found or if the wallet is watch-only.
// Implements tui.WalletProvider.
func (m *Manager) WalletAddress(id string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, w := range m.instances {
		if w.Cfg.ID == id {
			return w.Address
		}
	}
	return ""
}

// WalletEnabled reports whether the wallet with the given ID is enabled.
// Returns false if not found.
// Implements tui.WalletProvider.
func (m *Manager) WalletEnabled(id string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, w := range m.instances {
		if w.Cfg.ID == id {
			return w.Cfg.Enabled
		}
	}
	return false
}

// WalletStats returns a cached statistics snapshot for the wallet with the given ID.
// Returns zero values if not found.
// Implements tui.WalletProvider.
func (m *Manager) WalletStats(id string) (balanceUSD, pnlUSD float64, openOrders, totalTrades int) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, w := range m.instances {
		if w.Cfg.ID == id {
			return w.Stats.Get()
		}
	}
	return 0, 0, 0, 0
}

// Primary returns the wallet marked as primary, or the first enabled wallet if none is marked.
func (m *Manager) Primary() *WalletInstance {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var first *WalletInstance
	for _, w := range m.instances {
		if !w.Cfg.Enabled {
			continue
		}
		if first == nil {
			first = w
		}
		if w.Cfg.Primary {
			return w
		}
	}
	return first
}

// SetPrimary marks walletID as primary and clears Primary on all others.
// Callers are responsible for persisting the change via config.Save.
func (m *Manager) SetPrimary(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Pre-check existence
	found := false
	for _, w := range m.instances {
		if w.Cfg.ID == id {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("wallet %q not found", id)
	}
	// Now safe to mutate
	for _, w := range m.instances {
		w.Cfg.Primary = w.Cfg.ID == id
	}
	if m.bus != nil {
		m.bus.Send(tui.WalletChangedMsg{ID: id, Enabled: true})
	}
	return nil
}

// Toggle enables or disables a wallet. Disabling triggers graceful drain (Stop).
// Callers are responsible for persisting the change via config.Save.
func (m *Manager) Toggle(id string, enabled bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, w := range m.instances {
		if w.Cfg.ID == id {
			w.Cfg.Enabled = enabled
			if !enabled {
				w.Stop()
			}
			if m.bus != nil {
				m.bus.Send(tui.WalletChangedMsg{ID: id, Enabled: enabled})
			}
			return nil
		}
	}
	return fmt.Errorf("wallet %q not found", id)
}
