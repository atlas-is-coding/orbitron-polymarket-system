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
