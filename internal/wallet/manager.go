package wallet

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"github.com/atlasdev/orbitron/internal/analytics"
	"github.com/atlasdev/orbitron/internal/api"
	"github.com/atlasdev/orbitron/internal/api/clob"
	"github.com/atlasdev/orbitron/internal/api/data"
	"github.com/atlasdev/orbitron/internal/api/ws"
	"github.com/atlasdev/orbitron/internal/auth"
	"github.com/atlasdev/orbitron/internal/config"
	"github.com/atlasdev/orbitron/internal/copytrading"
	"github.com/atlasdev/orbitron/internal/health"
	"github.com/atlasdev/orbitron/internal/monitor"
	"github.com/atlasdev/orbitron/internal/notify"
	"github.com/atlasdev/orbitron/internal/storage"
	"github.com/atlasdev/orbitron/internal/tui"
)

// Manager manages the set of active and inactive wallet instances.
// All public methods are thread-safe.
type Manager struct {
	mu         sync.RWMutex
	instances  []*WalletInstance
	bus        *tui.EventBus
	cfg        *config.Config
	cfgPath    string
	wsClient   *ws.Client
	dataClient *data.Client
	notifier   notify.Notifier
	db         storage.Store
	dialFn     api.DialFunc
	builderKey string
	log        zerolog.Logger
}

// NewManager creates a Manager. bus, cfg, wsClient may be nil (e.g., in tests).
func NewManager(bus *tui.EventBus, cfg *config.Config, wsClient *ws.Client) *Manager {
	return &Manager{
		bus:      bus,
		cfg:      cfg,
		wsClient: wsClient,
	}
}

// SetDialer sets the proxy dialer used for geoblock checks before order placement.
func (m *Manager) SetDialer(dial api.DialFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.dialFn = dial
}

// SetLogger sets the logger for the manager.
func (m *Manager) SetLogger(log zerolog.Logger) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.log = log
}

// SetDataClient sets the data client for the manager.
func (m *Manager) SetDataClient(client *data.Client) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.dataClient = client
}

// SetNotifier sets the notifier for the manager.
func (m *Manager) SetNotifier(n notify.Notifier) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.notifier = n
}

// SetDatabase sets the database for the manager.
func (m *Manager) SetDatabase(db storage.Store) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.db = db
}

// SetConfigPath sets the config path for the manager.
func (m *Manager) SetConfigPath(path string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cfgPath = path
}

// SetBuilderKey sets the builder key for the manager.
func (m *Manager) SetBuilderKey(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.builderKey = key
}

// Activate initialises a wallet (derives L2 credentials, creates CLOB client, starts subsystems).
// If the wallet is already active, it returns the existing instance.
func (m *Manager) Activate(ctx context.Context, wCfg config.WalletConfig) (*WalletInstance, error) {
	m.mu.Lock()
	// Check if already active
	for _, w := range m.instances {
		if w.Cfg.ID == wCfg.ID && w.ClobClient != nil {
			m.mu.Unlock()
			return w, nil
		}
	}
	m.mu.Unlock()

	if !wCfg.Enabled {
		return nil, fmt.Errorf("wallet %q is disabled", wCfg.ID)
	}

	// 1. Setup L1 Signer
	l1, err := auth.NewL1Signer(wCfg.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("l1 signer: %w", err)
	}
	addr := l1.Address()

	// 2. Determine CLOB URL based on ChainID
	clobURL := m.cfg.API.ClobURL
	if wCfg.ChainID == 80002 {
		clobURL = "https://clob-testnet.polymarket.com"
	}

	// 3. Create HTTP client for this URL
	httpClient := api.NewClientWithDialer(clobURL, m.cfg.API.TimeoutSec, m.cfg.API.MaxRetries, m.dialFn)
	pubClobClient := clob.NewClient(httpClient, nil)

	// 4. Setup L2 Credentials
	l2 := &auth.L2Credentials{
		APIKey:     wCfg.APIKey,
		APISecret:  wCfg.APISecret,
		Passphrase: wCfg.Passphrase,
		Address:    addr,
	}

	if l2.APIKey == "" {
		m.log.Info().Str("wallet", wCfg.Label).Msg("auto-deriving L2 api_key...")
		derived, err := pubClobClient.DeriveAPIKey(l1, wCfg.ChainID)
		if err != nil {
			return nil, fmt.Errorf("derive api key: %w", err)
		}
		l2.APIKey = derived.APIKey
		l2.APISecret = derived.APISecret
		l2.Passphrase = derived.Passphrase
	}

	wClobClient := clob.NewClient(httpClient, l2)

	// 5. Subscribe WebSocket user events
	if m.wsClient != nil {
		m.wsClient.Subscribe(ws.UserSubscription(l2), func(msg *ws.Message) {
			m.log.Debug().Str("event", msg.EventType).Str("wallet", wCfg.Label).Msg("ws user event")
		})
	}

	// 6. Create instance
	instCtx, instCancel := context.WithCancel(ctx)
	inst := &WalletInstance{
		Cfg:        wCfg,
		Address:    addr,
		L2:         l2,
		ClobClient: wClobClient,
		Stats:      &WalletStats{},
		cancel:     instCancel,
	}

	// 7. Setup order executor
	orderSigner := auth.NewOrderSigner(l1, wCfg.ChainID, wCfg.NegRisk)
	inst.Executor = copytrading.NewOrderExecutor(wClobClient, orderSigner, l2.APIKey, addr, m.log)
	if m.builderKey != "" {
		inst.Executor.WithBuilderKey(m.builderKey)
	}

	// 8. Setup Analytics
	var analyticsHub *analytics.AnalyticsHub
	if m.cfg.Analytics.Enabled && l1 != nil {
		analyticsHub = analytics.NewAnalyticsHub(m.cfg.Analytics.BatchSize)
		analyticsClient := analytics.NewClient(l1, wCfg.Label, m.cfg.Analytics.Endpoint, m.log)
		go func(hub *analytics.AnalyticsHub, client *analytics.Client) {
			interval := time.Duration(m.cfg.Analytics.ReportInterval) * time.Second
			if interval == 0 {
				interval = 30 * time.Second
			}
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			for {
				select {
				case <-instCtx.Done():
					return
				case <-hub.Trigger():
				case <-ticker.C:
					trades := hub.Flush()
					if len(trades) > 0 {
						if err := client.Report(instCtx, trades); err != nil {
							m.log.Warn().Err(err).Str("wallet", wCfg.Label).Msg("failed to report analytics")
						}
					}
				}
			}
		}(analyticsHub, analyticsClient)
	}

	// 9. Setup Trades Monitor
	if m.cfg.Monitor.Trades.Enabled {
		tm := monitor.NewTradesMonitor(analyticsHub, wClobClient, m.dataClient, m.notifier, &m.cfg.Monitor.Trades, m.log, addr, m.db)
		if m.bus != nil {
			tm.SetBus(m.bus)
		}
		inst.TradesMon = tm
		go tm.Run(instCtx)
	}

	// 10. Setup Copy Trader
	if m.cfg.Copytrading.Enabled && m.db != nil && l1 != nil {
		ct := copytrading.NewCopyTrader(
			m.cfgPath,
			func() *config.CopytradingConfig { return &m.cfg.Copytrading },
			m.dataClient,
			inst.Executor,
			m.db,
			m.notifier,
			wClobClient,
			m.log,
		)
		inst.CopyTrader = ct
		go ct.Run(instCtx)
	}

	// 11. Register or update in manager
	m.mu.Lock()
	found := false
	for i, w := range m.instances {
		if w.Cfg.ID == wCfg.ID {
			m.instances[i] = inst
			found = true
			break
		}
	}
	if !found {
		m.instances = append(m.instances, inst)
	}
	m.mu.Unlock()

	if m.bus != nil {
		m.bus.Send(tui.WalletAddedMsg{
			ID:      inst.Cfg.ID,
			Address: inst.Address,
			Label:   inst.Cfg.Label,
			Enabled: inst.Cfg.Enabled,
			Primary: inst.Cfg.Primary,
		})
	}

	return inst, nil
}

// AvailableWallets returns IDs of all enabled wallets that have an active executor.
func (m *Manager) AvailableWallets() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var ids []string
	for _, w := range m.instances {
		if w.Cfg.Enabled && w.Executor != nil {
			ids = append(ids, w.Cfg.ID)
		}
	}
	return ids
}

// AddInactive adds a wallet without starting any subsystems.
// Used for disabled wallets and in tests.
func (m *Manager) AddInactive(cfg config.WalletConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()

	addr := ""
	if cfg.PrivateKey != "" {
		if l1, err := auth.NewL1Signer(cfg.PrivateKey); err == nil {
			addr = l1.Address()
		}
	}

	m.instances = append(m.instances, &WalletInstance{
		Cfg:     cfg,
		Address: addr,
		Stats:   &WalletStats{},
	})
	if m.bus != nil {
		m.bus.Send(tui.WalletAddedMsg{
			ID:      cfg.ID,
			Address: addr,
			Label:   cfg.Label,
			Enabled: cfg.Enabled,
			Primary: cfg.Primary,
		})
	}
}

// AddActive adds a fully initialised wallet instance and broadcasts WalletAddedMsg.
func (m *Manager) AddActive(inst *WalletInstance) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.instances = append(m.instances, inst)
	if m.bus != nil {
		m.bus.Send(tui.WalletAddedMsg{
			ID:      inst.Cfg.ID,
			Address: inst.Address,
			Label:   inst.Cfg.Label,
			Enabled: inst.Cfg.Enabled,
			Primary: inst.Cfg.Primary,
		})
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
		m.bus.Send(tui.WalletChangedMsg{ID: id, Enabled: true, Primary: true})
	}
	return nil
}

// PlaceOrder places a limit order for the wallet identified by walletID.
// Requires the wallet to be active (have ClobClient and L2 credentials configured).
// negRisk is passed through to the order signer (neg-risk markets).
func (m *Manager) PlaceOrder(walletID, tokenID, side, orderType string, price, sizeUSD float64, negRisk bool) (string, error) {
	geo, geoErr := health.CheckGeoblock(m.dialFn)
	if geoErr == nil && geo.Blocked {
		return "", fmt.Errorf("trading blocked in %s (IP: %s) — configure [proxy] in config.toml", geo.Country, geo.IP)
	}

	m.mu.RLock()
	var inst *WalletInstance
	for _, w := range m.instances {
		if w.Cfg.ID == walletID {
			inst = w
			break
		}
	}
	m.mu.RUnlock()

	if inst == nil {
		return "", fmt.Errorf("wallet %q not found", walletID)
	}
	if inst.ClobClient == nil || inst.L2 == nil {
		return "", fmt.Errorf("wallet %q is not active (no CLOB client)", walletID)
	}
	if inst.Cfg.PrivateKey == "" {
		return "", fmt.Errorf("wallet %q has no private key", walletID)
	}

	l1, err := auth.NewL1Signer(inst.Cfg.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("wallet %q: derive L1: %w", walletID, err)
	}
	signer := auth.NewOrderSigner(l1, inst.Cfg.ChainID, negRisk)

	exec := copytrading.NewOrderExecutor(
		inst.ClobClient,
		signer,
		inst.L2.APIKey,
		inst.Address,
		zerolog.Nop(),
	)
	return exec.PlaceLimit(tokenID, side, orderType, price, sizeUSD)
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
