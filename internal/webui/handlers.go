package webui

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/atlasdev/orbitron/internal/api/gamma"
	"github.com/atlasdev/orbitron/internal/auth"
	"github.com/atlasdev/orbitron/internal/config"
	"github.com/atlasdev/orbitron/internal/i18n"
	"github.com/atlasdev/orbitron/internal/markets"
	"github.com/atlasdev/orbitron/internal/nexus"
	"github.com/atlasdev/orbitron/internal/storage"
	"github.com/atlasdev/orbitron/internal/tui"
	"github.com/atlasdev/orbitron/internal/wallet"
)

// OrderCanceler wraps TradesMonitor cancel operations.
type OrderCanceler interface {
	CancelOrder(id string) error
	CancelAllOrders() error
}

// WalletMutator allows the Web UI to mutate wallet runtime state.
// Implemented by *wallet.Manager.
type WalletMutator interface {
	UpdateLabel(id, label string) error
	Toggle(id string, enabled bool) error
	Remove(id string) error
}

// WalletAdder allows the Web UI to register a new wallet in the manager.
type WalletAdder interface {
	AddInactive(cfg config.WalletConfig)
	Activate(ctx context.Context, cfg config.WalletConfig) (*wallet.WalletInstance, error)
}

// MarketsProvider exposes markets data to the Web UI.
type MarketsProvider interface {
	GetByTag(slug string) []gamma.Market
	GetMarket(conditionID string) (gamma.Market, bool)
	Tags() []gamma.Tag
	AddAlert(rule markets.AlertRule) string
	GetTrending() []gamma.Market
	TotalCount() int
}

// OrderPlacer places a limit order for a given wallet.
// Implemented by *wallet.Manager.
type OrderPlacer interface {
        PlaceOrder(walletID, tokenID, side, orderType string, price, sizeUSD float64, negRisk bool) (string, error)
}

// TradingProvider allows starting/stopping strategies.
type TradingProvider interface {
        StartStrategy(name string) error
        StopStrategy(name string) error
        SetStrategyWallets(name string, walletIDs []string) error
}

// Server is the Web UI HTTP server.
type Server struct {
	cfg      *config.Config
	cfgMu    sync.RWMutex
	cfgPath  string
	password string
	bus      *tui.EventBus
	nx       *tui.Nexus
	nexus    *nexus.Nexus     // may be nil
	canceler OrderCanceler
	wallets  WalletMutator    // may be nil
	adder    WalletAdder      // may be nil
	mkts     MarketsProvider  // may be nil
	placer   OrderPlacer      // may be nil
	trading  TradingProvider  // may be nil
	store    storage.Store    // may be nil
	state    *WebState
	hub      *hub
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// jwtMiddleware validates Bearer token from Authorization header or ?token= query param.
func (s *Server) jwtMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		token := strings.TrimPrefix(auth, "Bearer ")
		if token == "" {
			token = r.URL.Query().Get("token")
		}
		if _, err := verifyJWT(token, s.cfg.WebUI.JWTSecret); err != nil {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next(w, r)
	}
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad request")
		return
	}
	if req.Password != s.password {
		writeError(w, http.StatusUnauthorized, "invalid password")
		return
	}
	token, err := signJWT("admin", 24*time.Hour, s.cfg.WebUI.JWTSecret)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (s *Server) handleOverview(w http.ResponseWriter, _ *http.Request) {
	snap := s.nx.Snapshot()

	subsMap := snap["subsystems"].(map[string]bool)
	type subsystemEntry struct {
		Name   string `json:"name"`
		Active bool   `json:"active"`
	}
	subsArr := make([]subsystemEntry, 0, len(subsMap))
	for name, active := range subsMap {
		subsArr = append(subsArr, subsystemEntry{Name: name, Active: active})
	}
	sort.Slice(subsArr, func(i, j int) bool { return subsArr[i].Name < subsArr[j].Name })

	wallets := snap["wallets"].([]tui.WalletStatsMsg)
	var primaryAddr, primaryID string
	for _, wl := range wallets {
		if wl.Primary {
			primaryAddr = wl.Address
			primaryID = wl.ID
			break
		}
	}
	if primaryAddr == "" && len(wallets) > 0 {
		primaryAddr = wallets[0].Address
		primaryID = wallets[0].ID
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"balance":        snap["balance"],
		"pnl":            snap["pnl"],
		"wallet":         primaryID,
		"wallet_address": primaryAddr,
		"subsystems":     subsArr,
		"orders":         snap["orders"],
		"positions":      snap["positions"],
		"strategies":     snap["strategies"],
		"wallets":        snap["wallets"],
	})
}

func (s *Server) handleStrategies(w http.ResponseWriter, _ *http.Request) {
	snap := s.nx.Snapshot()
	writeJSON(w, http.StatusOK, snap["strategies"])
}

func (s *Server) handleStartStrategy(w http.ResponseWriter, r *http.Request) {
        name := strings.TrimPrefix(r.URL.Path, "/api/v1/strategies/")
        name = strings.TrimSuffix(name, "/start")
        var req struct {
                WalletIDs []string `json:"wallet_ids"`
        }
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
                writeError(w, http.StatusBadRequest, "bad request")
                return
        }
        if s.trading == nil {
                writeError(w, http.StatusServiceUnavailable, "trading engine unavailable")
                return
        }
        if err := s.trading.SetStrategyWallets(name, req.WalletIDs); err != nil {
                writeError(w, http.StatusInternalServerError, err.Error())
                return
        }
        if err := s.trading.StartStrategy(name); err != nil {
                writeError(w, http.StatusInternalServerError, err.Error())
                return
        }
        writeJSON(w, http.StatusOK, map[string]string{"status": "started"})
}

func (s *Server) handleStopStrategy(w http.ResponseWriter, r *http.Request) {
        name := strings.TrimPrefix(r.URL.Path, "/api/v1/strategies/")
        name = strings.TrimSuffix(name, "/stop")
        if s.trading == nil {
                writeError(w, http.StatusServiceUnavailable, "trading engine unavailable")
                return
        }
        if err := s.trading.StopStrategy(name); err != nil {
                writeError(w, http.StatusInternalServerError, err.Error())
                return
        }
        writeJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
}

func (s *Server) handleOrders(w http.ResponseWriter, _ *http.Request) {
	snap := s.nx.Snapshot()
	writeJSON(w, http.StatusOK, snap["orders"])
}

func (s *Server) handlePositions(w http.ResponseWriter, _ *http.Request) {
	snap := s.nx.Snapshot()
	writeJSON(w, http.StatusOK, snap["positions"])
}

func (s *Server) handleLogs(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.state.Logs())
}

func (s *Server) handleCopytrading(w http.ResponseWriter, _ *http.Request) {
	s.cfgMu.RLock()
	enabled := s.cfg.Copytrading.Enabled
	traders := make([]config.TraderConfig, len(s.cfg.Copytrading.Traders))
	copy(traders, s.cfg.Copytrading.Traders)
	s.cfgMu.RUnlock()
	writeJSON(w, http.StatusOK, map[string]any{
		"enabled": enabled,
		"traders": traders,
	})
}

func (s *Server) handleGetSettings(w http.ResponseWriter, _ *http.Request) {
	s.cfgMu.RLock()
	cfg := *s.cfg
	s.cfgMu.RUnlock()
	// Mask secrets in auth (deprecated section)
	cfg.Auth.PrivateKey = maskSecret(cfg.Auth.PrivateKey)
	cfg.Auth.APIKey = maskSecret(cfg.Auth.APIKey)
	cfg.Auth.APISecret = maskSecret(cfg.Auth.APISecret)
	cfg.Auth.Passphrase = maskSecret(cfg.Auth.Passphrase)
	cfg.Telegram.BotToken = maskSecret(cfg.Telegram.BotToken)
	cfg.WebUI.JWTSecret = maskSecret(cfg.WebUI.JWTSecret)
	// Mask wallet secrets (deep copy to avoid mutating the live config)
	maskedWallets := make([]config.WalletConfig, len(cfg.Wallets))
	copy(maskedWallets, cfg.Wallets)
	for i := range maskedWallets {
		maskedWallets[i].PrivateKey = maskSecret(maskedWallets[i].PrivateKey)
		maskedWallets[i].APIKey = maskSecret(maskedWallets[i].APIKey)
		maskedWallets[i].APISecret = maskSecret(maskedWallets[i].APISecret)
		maskedWallets[i].Passphrase = maskSecret(maskedWallets[i].Passphrase)
	}
	cfg.Wallets = maskedWallets
	writeJSON(w, http.StatusOK, cfg)
}

func maskSecret(s string) string {
	if s == "" {
		return ""
	}
	return "***"
}

func (s *Server) handlePostSettings(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad request")
		return
	}
	s.cfgMu.Lock()
	cfgCopy := *s.cfg
	if err := applyConfigKey(&cfgCopy, req.Key, req.Value); err != nil {
		s.cfgMu.Unlock()
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := config.Save(s.cfgPath, &cfgCopy); err != nil {
		s.cfgMu.Unlock()
		writeError(w, http.StatusInternalServerError, "save failed")
		return
	}
	*s.cfg = cfgCopy
	s.cfgMu.Unlock()
	if s.bus != nil {
		s.bus.Send(tui.ConfigReloadedMsg{Config: s.cfg})
	}
	// Side effect: language change
	if req.Key == "ui.language" {
		i18n.SetLanguage(req.Value)
		if s.bus != nil {
			s.bus.Send(tui.LanguageChangedMsg{})
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleCancelOrder(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/orders/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing order id")
		return
	}
	if s.canceler == nil {
		writeError(w, http.StatusServiceUnavailable, "TradesMonitor not enabled")
		return
	}
	if err := s.canceler.CancelOrder(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

func (s *Server) handleCancelAll(w http.ResponseWriter, _ *http.Request) {
	if s.canceler == nil {
		writeError(w, http.StatusServiceUnavailable, "TradesMonitor not enabled")
		return
	}
	if err := s.canceler.CancelAllOrders(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "all cancelled"})
}

func (s *Server) handlePlaceOrder(w http.ResponseWriter, r *http.Request) {
        var req struct {
                TokenID   string  `json:"token_id"`
                Side      string  `json:"side"`
                OrderType string  `json:"order_type"`
                Price     float64 `json:"price"`
                SizeUSD   float64 `json:"size_usd"`
                WalletID  string  `json:"wallet_id"`
                NegRisk   bool    `json:"neg_risk"`
        }
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
                writeError(w, http.StatusBadRequest, "bad request")
                return
        }
        if req.Price <= 0 || req.Price >= 1 {
                writeError(w, http.StatusBadRequest, "price must be between 0.001 and 0.999")
                return
        }
        if req.SizeUSD <= 0 {
                writeError(w, http.StatusBadRequest, "size_usd must be positive")
                return
        }
        if req.Side != "YES" && req.Side != "NO" {
                writeError(w, http.StatusBadRequest, "side must be YES or NO")
                return
        }
        if s.placer == nil {
                writeError(w, http.StatusServiceUnavailable, "order placement unavailable")
                return
        }
        orderID, err := s.placer.PlaceOrder(req.WalletID, req.TokenID, req.Side, req.OrderType, req.Price, req.SizeUSD, req.NegRisk)
        if err != nil {
                writeError(w, http.StatusInternalServerError, err.Error())
                return
        }
        writeJSON(w, http.StatusOK, map[string]string{"order_id": orderID})
}

// handleOrderHistory returns order history for a wallet.
func (s *Server) handleOrderHistory(w http.ResponseWriter, r *http.Request) {
	walletAddr := r.URL.Query().Get("wallet_address")
	if walletAddr == "" {
		writeError(w, http.StatusBadRequest, "missing wallet_address")
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	if s.store == nil {
		writeError(w, http.StatusServiceUnavailable, "order store not available")
		return
	}

	filters := storage.OrderFilters{
		WalletAddress: walletAddr,
		Limit:         limit,
	}
	orders, err := s.store.GetOrders(r.Context(), filters)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch order history")
		return
	}

	// Apply offset and limit
	if offset > len(orders) {
		offset = len(orders)
	}
	orders = orders[offset:]
	if limit > 0 && len(orders) > limit {
		orders = orders[:limit]
	}

	if orders == nil {
		orders = []*storage.Order{}
	}

	writeJSON(w, http.StatusOK, orders)
}

// handleTradeHistory returns trade history for a wallet.
func (s *Server) handleTradeHistory(w http.ResponseWriter, r *http.Request) {
	walletAddr := r.URL.Query().Get("wallet_address")
	if walletAddr == "" {
		writeError(w, http.StatusBadRequest, "missing wallet_address")
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if s.store == nil {
		writeError(w, http.StatusServiceUnavailable, "trade store not available")
		return
	}

	// Get trades for the wallet within a time range (e.g., last 90 days)
	now := time.Now()
	ninetyDaysAgo := now.AddDate(0, 0, -90)

	trades, err := s.store.GetTrades(r.Context(), walletAddr, ninetyDaysAgo, now)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch trade history")
		return
	}

	// Apply limit
	if limit > 0 && len(trades) > limit {
		trades = trades[:limit]
	}

	if trades == nil {
		trades = []*storage.Trade{}
	}

	writeJSON(w, http.StatusOK, trades)
}

// handleWalletStats returns wallet statistics.
func (s *Server) handleWalletStats(w http.ResponseWriter, r *http.Request) {
	walletAddr := r.URL.Query().Get("wallet_address")
	if walletAddr == "" {
		writeError(w, http.StatusBadRequest, "missing wallet_address")
		return
	}

	if s.store == nil {
		writeError(w, http.StatusServiceUnavailable, "stats store not available")
		return
	}

	stats, err := s.store.GetWalletStats(r.Context(), walletAddr, 1)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch wallet stats")
		return
	}

	if stats == nil || len(stats) == 0 {
		writeJSON(w, http.StatusOK, map[string]any{
			"wallet_address": walletAddr,
			"balance_usd":    0.0,
			"pnl_usd":        0.0,
		})
		return
	}

	writeJSON(w, http.StatusOK, stats[0])
}

func (s *Server) handleAddTrader(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Address  string  `json:"address"`
		Label    string  `json:"label"`
		AllocPct float64 `json:"alloc_pct"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad request")
		return
	}
	if req.AllocPct == 0 {
		req.AllocPct = 5.0
	}
	s.cfgMu.Lock()
	cfgCopy := *s.cfg
	traders := make([]config.TraderConfig, len(cfgCopy.Copytrading.Traders))
	copy(traders, cfgCopy.Copytrading.Traders)
	cfgCopy.Copytrading.Traders = traders
	for _, t := range cfgCopy.Copytrading.Traders {
		if t.Address == req.Address {
			s.cfgMu.Unlock()
			writeError(w, http.StatusConflict, "trader already exists")
			return
		}
	}
	cfgCopy.Copytrading.Traders = append(cfgCopy.Copytrading.Traders, config.TraderConfig{
		Address:        req.Address,
		Label:          req.Label,
		Enabled:        true,
		AllocationPct:  req.AllocPct,
		MaxPositionUSD: 50.0,
		SizeMode:       cfgCopy.Copytrading.SizeMode,
	})
	if err := config.Save(s.cfgPath, &cfgCopy); err != nil {
		s.cfgMu.Unlock()
		writeError(w, http.StatusInternalServerError, "save failed")
		return
	}
	*s.cfg = cfgCopy
	s.cfgMu.Unlock()
	if s.bus != nil {
		s.bus.Send(tui.ConfigReloadedMsg{Config: s.cfg})
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "added"})
}

func (s *Server) handleRemoveTrader(w http.ResponseWriter, r *http.Request) {
	// path: /api/v1/copytrading/traders/{addr}
	addr := strings.TrimPrefix(r.URL.Path, "/api/v1/copytrading/traders/")
	s.cfgMu.Lock()
	cfgCopy := *s.cfg
	found := false
	traders := make([]config.TraderConfig, 0, len(cfgCopy.Copytrading.Traders))
	for _, t := range cfgCopy.Copytrading.Traders {
		if t.Address == addr {
			found = true
			continue
		}
		traders = append(traders, t)
	}
	if !found {
		s.cfgMu.Unlock()
		writeError(w, http.StatusNotFound, "trader not found")
		return
	}
	cfgCopy.Copytrading.Traders = traders
	if err := config.Save(s.cfgPath, &cfgCopy); err != nil {
		s.cfgMu.Unlock()
		writeError(w, http.StatusInternalServerError, "save failed")
		return
	}
	*s.cfg = cfgCopy
	s.cfgMu.Unlock()
	if s.bus != nil {
		s.bus.Send(tui.ConfigReloadedMsg{Config: s.cfg})
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

func (s *Server) handleToggleTrader(w http.ResponseWriter, r *http.Request) {
	// path: /api/v1/copytrading/traders/{addr}/toggle
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/copytrading/traders/")
	addr := strings.TrimSuffix(path, "/toggle")
	s.cfgMu.Lock()
	cfgCopy := *s.cfg
	found := false
	traders := make([]config.TraderConfig, len(cfgCopy.Copytrading.Traders))
	copy(traders, cfgCopy.Copytrading.Traders)
	cfgCopy.Copytrading.Traders = traders
	for i, t := range cfgCopy.Copytrading.Traders {
		if t.Address == addr {
			cfgCopy.Copytrading.Traders[i].Enabled = !t.Enabled
			found = true
			break
		}
	}
	if !found {
		s.cfgMu.Unlock()
		writeError(w, http.StatusNotFound, "trader not found")
		return
	}
	if err := config.Save(s.cfgPath, &cfgCopy); err != nil {
		s.cfgMu.Unlock()
		writeError(w, http.StatusInternalServerError, "save failed")
		return
	}
	*s.cfg = cfgCopy
	s.cfgMu.Unlock()
	if s.bus != nil {
		s.bus.Send(tui.ConfigReloadedMsg{Config: s.cfg})
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "toggled"})
}

func (s *Server) handleEditTrader(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/copytrading/traders/")
	addr := strings.TrimSuffix(path, "/edit")
	var req struct {
		Label          string  `json:"label"`
		AllocPct       float64 `json:"alloc_pct"`
		MaxPositionUSD float64 `json:"max_position_usd"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad request")
		return
	}
	s.cfgMu.Lock()
	cfgCopy := *s.cfg
	traders := make([]config.TraderConfig, len(cfgCopy.Copytrading.Traders))
	copy(traders, cfgCopy.Copytrading.Traders)
	cfgCopy.Copytrading.Traders = traders
	found := false
	for i, t := range cfgCopy.Copytrading.Traders {
		if t.Address == addr {
			cfgCopy.Copytrading.Traders[i].Label = req.Label
			if req.AllocPct > 0 {
				cfgCopy.Copytrading.Traders[i].AllocationPct = req.AllocPct
			}
			if req.MaxPositionUSD > 0 {
				cfgCopy.Copytrading.Traders[i].MaxPositionUSD = req.MaxPositionUSD
			}
			found = true
			break
		}
	}
	if !found {
		s.cfgMu.Unlock()
		writeError(w, http.StatusNotFound, "trader not found")
		return
	}
	if err := config.Save(s.cfgPath, &cfgCopy); err != nil {
		s.cfgMu.Unlock()
		writeError(w, http.StatusInternalServerError, "save failed")
		return
	}
	*s.cfg = cfgCopy
	s.cfgMu.Unlock()
	if s.bus != nil {
		s.bus.Send(tui.ConfigReloadedMsg{Config: s.cfg})
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// ── Wallet handlers ───────────────────────────────────────────────────────────

// handleGetWallets returns the cached wallet list from Nexus.
func (s *Server) handleGetWallets(w http.ResponseWriter, _ *http.Request) {
	snap := s.nx.Snapshot()
	writeJSON(w, http.StatusOK, snap["wallets"])
}

// handleAddWallet handles POST /api/v1/wallets
// Body: {"private_key": "hex (with or without 0x prefix)"}
func (s *Server) handleAddWallet(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PrivateKey string `json:"private_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PrivateKey == "" {
		writeError(w, http.StatusBadRequest, "private_key required")
		return
	}
	l1, err := auth.NewL1Signer(req.PrivateKey)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid private key")
		return
	}
	addr := l1.Address()

	s.cfgMu.Lock()
	// Check for duplicate by address
	for _, wc := range s.cfg.Wallets {
		existingL1, err2 := auth.NewL1Signer(wc.PrivateKey)
		if err2 == nil && existingL1.Address() == addr {
			s.cfgMu.Unlock()
			writeError(w, http.StatusConflict, "wallet already exists")
			return
		}
	}
	id := fmt.Sprintf("w%d", time.Now().UnixMilli())
	chainID := int64(137)
	if len(s.cfg.Wallets) > 0 && s.cfg.Wallets[0].ChainID != 0 {
		chainID = s.cfg.Wallets[0].ChainID
	}
	wCfg := config.WalletConfig{
		ID:         id,
		Label:      addr[:8] + "…" + addr[len(addr)-4:],
		PrivateKey: strings.TrimPrefix(req.PrivateKey, "0x"),
		ChainID:    chainID,
		Enabled:    true,
	}
	cfgCopy := *s.cfg
	newWallets := make([]config.WalletConfig, len(s.cfg.Wallets)+1)
	copy(newWallets, s.cfg.Wallets)
	newWallets[len(s.cfg.Wallets)] = wCfg
	cfgCopy.Wallets = newWallets
	if err := config.Save(s.cfgPath, &cfgCopy); err != nil {
		s.cfgMu.Unlock()
		writeError(w, http.StatusInternalServerError, "save failed")
		return
	}
	*s.cfg = cfgCopy
	s.cfgMu.Unlock()

	// Check token allowances (non-blocking for UI, but we return results)
	allowances, _ := wallet.CheckAllowances(r.Context(), s.cfg.API.PolygonRPC, addr)

	// Automatically grant missing allowances
	if err := wallet.GrantMissingAllowances(r.Context(), s.cfg.API.PolygonRPC, req.PrivateKey, allowances); err == nil {
		// Re-check after granting to update the status in response
		if updated, err2 := wallet.CheckAllowances(r.Context(), s.cfg.API.PolygonRPC, addr); err2 == nil {
			allowances = updated
		}
	}

	if s.adder != nil {
		s.adder.Activate(r.Context(), wCfg)
	}
	if s.bus != nil {
		s.bus.Send(tui.WalletAddedMsg{
			ID:         id,
			Address:    addr,
			Label:      wCfg.Label,
			Enabled:    true,
			Allowances: allowances,
		})
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"id":          id,
		"address":     addr,
		"label":       wCfg.Label,
		"enabled":     true,
		"balance_usd": 0,
		"pnl_usd":     0,
		"open_orders": 0,
		"allowances":  allowances,
	})
}

// handleUpdateWallet handles PATCH /api/v1/wallets/:id
// Accepts JSON body {"label": "new name"}.
func (s *Server) handleUpdateWallet(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/wallets/")
	id = strings.TrimSuffix(id, "/toggle")
	var req struct {
		Label string `json:"label"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Label == "" {
		writeError(w, http.StatusBadRequest, "label required")
		return
	}
	if s.wallets == nil {
		writeError(w, http.StatusServiceUnavailable, "wallet manager unavailable")
		return
	}
	if err := s.wallets.UpdateLabel(id, req.Label); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	// Persist label change to config
	s.cfgMu.Lock()
	cfgCopy := *s.cfg
	newWallets := make([]config.WalletConfig, len(cfgCopy.Wallets))
	copy(newWallets, cfgCopy.Wallets)
	cfgCopy.Wallets = newWallets
	for i, wc := range cfgCopy.Wallets {
		if wc.ID == id {
			cfgCopy.Wallets[i].Label = req.Label
			break
		}
	}
	_ = config.Save(s.cfgPath, &cfgCopy)
	*s.cfg = cfgCopy
	s.cfgMu.Unlock()
	// Sync WebState synchronously and broadcast WS event
	wallets := s.state.Wallets()
	for _, we := range wallets {
		if we.ID == id {
			we.Label = req.Label
			s.state.UpsertWallet(we)
			s.hub.broadcast(WsEvent{Type: "wallet_stats", Data: we})
			if s.bus != nil {
				s.bus.Send(tui.WalletStatsMsg{
					ID: we.ID, Label: we.Label, Enabled: we.Enabled, Primary: we.Primary,
					BalanceUSD: we.BalanceUSD, PnLUSD: we.PnLUSD,
					OpenOrders: we.OpenOrders, TotalTrades: we.TotalTrades,
				})
			}
			break
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// handleToggleWallet handles POST /api/v1/wallets/:id/toggle
// Accepts JSON body {"enabled": true/false}.
func (s *Server) handleToggleWallet(w http.ResponseWriter, r *http.Request) {
	// path: /api/v1/wallets/{id}/toggle
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/wallets/")
	id := strings.TrimSuffix(path, "/toggle")
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad request")
		return
	}
	if s.wallets == nil {
		writeError(w, http.StatusServiceUnavailable, "wallet manager unavailable")
		return
	}
	if err := s.wallets.Toggle(id, req.Enabled); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	// Persist toggle to config
	s.cfgMu.Lock()
	cfgCopy := *s.cfg
	newWallets := make([]config.WalletConfig, len(cfgCopy.Wallets))
	copy(newWallets, cfgCopy.Wallets)
	cfgCopy.Wallets = newWallets
	for i, wc := range cfgCopy.Wallets {
		if wc.ID == id {
			cfgCopy.Wallets[i].Enabled = req.Enabled
			break
		}
	}
	_ = config.Save(s.cfgPath, &cfgCopy)
	*s.cfg = cfgCopy
	s.cfgMu.Unlock()
	// Sync WebState synchronously to avoid race with async EventBus
	wallets := s.state.Wallets()
	for _, we := range wallets {
		if we.ID == id {
			we.Enabled = req.Enabled
			s.state.UpsertWallet(we)
			break
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "toggled"})
}

// handleDeleteWallet handles DELETE /api/v1/wallets/:id
func (s *Server) handleDeleteWallet(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/wallets/")
	id = strings.TrimSuffix(id, "/toggle")
	if s.wallets == nil {
		writeError(w, http.StatusServiceUnavailable, "wallet manager unavailable")
		return
	}
	if err := s.wallets.Remove(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	// Persist removal to config
	s.cfgMu.Lock()
	cfgCopy := *s.cfg
	filtered := make([]config.WalletConfig, 0, len(cfgCopy.Wallets))
	for _, wc := range cfgCopy.Wallets {
		if wc.ID != id {
			filtered = append(filtered, wc)
		}
	}
	cfgCopy.Wallets = filtered
	_ = config.Save(s.cfgPath, &cfgCopy)
	*s.cfg = cfgCopy
	s.cfgMu.Unlock()
	// Sync WebState synchronously (manager.Remove already sends EventBus, but state update is async)
	s.state.RemoveWallet(id)
	writeJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

// handleGetHealth returns the latest health snapshot (public endpoint, no auth).
func (s *Server) handleGetHealth(w http.ResponseWriter, _ *http.Request) {
	snap := s.state.GetHealth()
	writeJSON(w, http.StatusOK, snap)
}

// ── Markets handlers ──────────────────────────────────────────────────────────

// handleMarketsList handles GET /api/v1/markets?tag=crypto&limit=50&offset=0
func (s *Server) handleMarketsList(w http.ResponseWriter, r *http.Request) {
	tag := r.URL.Query().Get("tag")
	limit := 20
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}

	var all []gamma.Market
	if s.mkts != nil {
		all = s.mkts.GetByTag(tag)
	}
	if all == nil {
		all = []gamma.Market{}
	}

	total := len(all)

	// Apply offset/limit in-handler
	result := all
	if offset > len(result) {
		offset = len(result)
	}
	result = result[offset:]
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"markets": result,
		"total":   total,
	})
}

// handleMarketsTags handles GET /api/v1/markets/tags
func (s *Server) handleMarketsTags(w http.ResponseWriter, r *http.Request) {
	var tags []gamma.Tag
	if s.mkts != nil {
		tags = s.mkts.Tags()
	}
	if tags == nil {
		tags = []gamma.Tag{}
	}
	writeJSON(w, http.StatusOK, tags)
}

// handleMarketsTrending handles GET /api/v1/markets/trending?limit=N&offset=M
func (s *Server) handleMarketsTrending(w http.ResponseWriter, r *http.Request) {
	if s.mkts == nil {
		writeJSON(w, http.StatusOK, map[string]any{"markets": []gamma.Market{}, "total": 0})
		return
	}
	limit := 24
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}

	all := s.mkts.GetTrending()
	total := len(all)

	result := all
	if offset > len(result) {
		offset = len(result)
	}
	result = result[offset:]
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"markets": result,
		"total":   total,
	})
}

// handleMarketsStats handles GET /api/v1/markets/stats
func (s *Server) handleMarketsStats(w http.ResponseWriter, r *http.Request) {
	total := 0
	if s.mkts != nil {
		total = s.mkts.TotalCount()
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"total":   total,
		"syncing": false,
	})
}

// handleMarketDetail handles GET /api/v1/markets/{conditionID}
func (s *Server) handleMarketDetail(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/markets/")
	if id == "" || id == "tags" || id == "trending" || id == "stats" {
		writeError(w, http.StatusBadRequest, "missing conditionID")
		return
	}
	if s.mkts == nil {
		writeError(w, http.StatusServiceUnavailable, "markets service not running")
		return
	}
	m, ok := s.mkts.GetMarket(id)
	if !ok {
		writeError(w, http.StatusNotFound, "market not found")
		return
	}
	writeJSON(w, http.StatusOK, m)
}

// handleApproveAllowance handles POST /api/v1/wallets/:id/approve
// Body: {"contract": "ctf" | "negrisk" | "all"}
func (s *Server) handleApproveAllowance(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/wallets/")
	id := strings.TrimSuffix(path, "/approve")

	var req struct {
		Contract string `json:"contract"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad request")
		return
	}

	s.cfgMu.RLock()
	var privKey, walletAddr, rpcURL string
	for _, wc := range s.cfg.Wallets {
		if wc.ID == id {
			privKey = wc.PrivateKey
			break
		}
	}
	rpcURL = s.cfg.API.PolygonRPC
	s.cfgMu.RUnlock()

	if privKey == "" {
		writeError(w, http.StatusNotFound, fmt.Sprintf("wallet %q not found", id))
		return
	}
	l1, err := auth.NewL1Signer(privKey)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "invalid private key in config")
		return
	}
	walletAddr = l1.Address()

	statuses, err := wallet.CheckAllowances(r.Context(), rpcURL, walletAddr)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "check allowances: "+err.Error())
		return
	}

	// Filter by contract type
	if req.Contract != "" && req.Contract != "all" {
		var filtered []config.AllowanceStatus
		for _, st := range statuses {
			if st.Approved {
				continue
			}
			switch req.Contract {
			case "ctf":
				if strings.Contains(strings.ToLower(st.SpenderName), "exchange") {
					filtered = append(filtered, st)
				}
			case "negrisk":
				if strings.Contains(strings.ToLower(st.SpenderName), "neg risk") {
					filtered = append(filtered, st)
				}
			}
		}
		statuses = filtered
	}

	if err := wallet.GrantMissingAllowances(r.Context(), rpcURL, privKey, statuses); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "approved"})
}

// handleTestEndpoint handles POST /api/v1/health/test
// Body: {"url": "https://..."} — pings the URL and returns latency.
func (s *Server) handleTestEndpoint(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		writeError(w, http.StatusBadRequest, "url required")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, req.URL, nil)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid url")
		return
	}
	start := time.Now()
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	resp.Body.Close()
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":         true,
		"status":     resp.StatusCode,
		"latency_ms": time.Since(start).Milliseconds(),
	})
}

// handleSaveConfig handles PUT /api/v1/settings.
// Accepts a partial config object; only non-"***" values are applied.
// Delegates to applyConfigKey for each dot-notation key.
func (s *Server) handleSaveConfig(w http.ResponseWriter, r *http.Request) {
	var raw map[string]any
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		writeError(w, http.StatusBadRequest, "bad request")
		return
	}
	// Flatten nested map to dot-notation keys
	flat := make(map[string]string)
	flattenMap(raw, "", flat)

	s.cfgMu.Lock()
	cfgCopy := *s.cfg
	changed := false
	for k, v := range flat {
		if v == "***" || v == "" {
			continue // skip masked or empty
		}
		if err := applyConfigKey(&cfgCopy, k, v); err != nil {
			continue // skip unknown keys silently
		}
		changed = true
	}
	if !changed {
		s.cfgMu.Unlock()
		writeJSON(w, http.StatusOK, map[string]string{"status": "no changes"})
		return
	}
	if err := config.Save(s.cfgPath, &cfgCopy); err != nil {
		s.cfgMu.Unlock()
		writeError(w, http.StatusInternalServerError, "save failed")
		return
	}
	*s.cfg = cfgCopy
	s.cfgMu.Unlock()
	if s.bus != nil {
		s.bus.Send(tui.ConfigReloadedMsg{Config: s.cfg})
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

// flattenMap recursively flattens a nested map into dot-notation string keys.
func flattenMap(m map[string]any, prefix string, out map[string]string) {
	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		switch val := v.(type) {
		case map[string]any:
			flattenMap(val, key, out)
		case nil:
			// skip
		default:
			out[key] = fmt.Sprintf("%v", val)
		}
	}
}

// handleStrategyConfig handles POST /api/v1/strategies/:name/config.
// Persists strategy-specific config fields to the TOML file and notifies the bus.
func (s *Server) handleStrategyConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	name := strings.TrimPrefix(r.URL.Path, "/api/v1/strategies/")
	name = strings.TrimSuffix(name, "/config")

	s.cfgMu.Lock()
	cfgCopy := *s.cfg

	var decodeErr error
	switch name {
	case "arbitrage":
		var sc config.ArbitrageConfig
		decodeErr = json.NewDecoder(r.Body).Decode(&sc)
		if decodeErr == nil {
			cfgCopy.Trading.Strategies.Arbitrage = sc
		}
	case "market_making":
		var sc config.MarketMakingConfig
		decodeErr = json.NewDecoder(r.Body).Decode(&sc)
		if decodeErr == nil {
			cfgCopy.Trading.Strategies.MarketMaking = sc
		}
	case "positive_ev":
		var sc config.PositiveEVConfig
		decodeErr = json.NewDecoder(r.Body).Decode(&sc)
		if decodeErr == nil {
			cfgCopy.Trading.Strategies.PositiveEV = sc
		}
	case "riskless_rate":
		var sc config.RisklessRateConfig
		decodeErr = json.NewDecoder(r.Body).Decode(&sc)
		if decodeErr == nil {
			cfgCopy.Trading.Strategies.RisklessRate = sc
		}
	case "fade_chaos":
		var sc config.FadeChaosConfig
		decodeErr = json.NewDecoder(r.Body).Decode(&sc)
		if decodeErr == nil {
			cfgCopy.Trading.Strategies.FadeChaos = sc
		}
	case "cross_market":
		var sc config.CrossMarketConfig
		decodeErr = json.NewDecoder(r.Body).Decode(&sc)
		if decodeErr == nil {
			cfgCopy.Trading.Strategies.CrossMarket = sc
		}
	default:
		s.cfgMu.Unlock()
		writeError(w, http.StatusNotFound, "unknown strategy: "+name)
		return
	}

	if decodeErr != nil {
		s.cfgMu.Unlock()
		writeError(w, http.StatusBadRequest, decodeErr.Error())
		return
	}
	if err := config.Save(s.cfgPath, &cfgCopy); err != nil {
		s.cfgMu.Unlock()
		writeError(w, http.StatusInternalServerError, "save failed")
		return
	}
	*s.cfg = cfgCopy
	s.cfgMu.Unlock()
	if s.bus != nil {
		s.bus.Send(tui.ConfigReloadedMsg{Config: s.cfg})
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleOrderbook proxies GET /api/v1/orderbook/{tokenID} to the CLOB public orderbook endpoint.
func (s *Server) handleOrderbook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "GET required")
		return
	}
	tokenID := strings.TrimPrefix(r.URL.Path, "/api/v1/orderbook/")
	if tokenID == "" {
		writeError(w, http.StatusBadRequest, "missing token_id in path")
		return
	}

	clobURL := s.cfg.API.ClobURL
	if clobURL == "" {
		clobURL = "https://clob.polymarket.com"
	}

	resp, err := http.Get(clobURL + "/book?token_id=" + tokenID) //nolint:noctx
	if err != nil {
		writeError(w, http.StatusBadGateway, "orderbook fetch failed: "+err.Error())
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body) //nolint:errcheck
}

// handleBatchOrders handles POST /api/v1/orders/batch.
// Accepts a JSON array of order requests and places each in sequence.
func (s *Server) handleBatchOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	if s.placer == nil {
		writeError(w, http.StatusServiceUnavailable, "order placement unavailable")
		return
	}

	var reqs []struct {
		TokenID   string  `json:"token_id"`
		Side      string  `json:"side"`
		OrderType string  `json:"order_type"`
		Price     float64 `json:"price"`
		SizeUSD   float64 `json:"size_usd"`
		WalletID  string  `json:"wallet_id"`
		NegRisk   bool    `json:"neg_risk"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqs); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(reqs) == 0 {
		writeError(w, http.StatusBadRequest, "orders array is empty")
		return
	}

	type result struct {
		OrderID string `json:"order_id,omitempty"`
		Error   string `json:"error,omitempty"`
	}
	results := make([]result, len(reqs))
	for i, req := range reqs {
		orderID, err := s.placer.PlaceOrder(req.WalletID, req.TokenID, req.Side, req.OrderType, req.Price, req.SizeUSD, req.NegRisk)
		if err != nil {
			results[i] = result{Error: err.Error()}
		} else {
			results[i] = result{OrderID: orderID}
		}
	}
	writeJSON(w, http.StatusOK, results)
}

// handleClosePosition handles POST /api/v1/positions/close.
// Closes an open position by placing a sell order at the specified price.
func (s *Server) handleClosePosition(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	if s.placer == nil {
		writeError(w, http.StatusServiceUnavailable, "order placement unavailable")
		return
	}

	var req struct {
		TokenID   string  `json:"token_id"`
		Side      string  `json:"side"`      // "YES" or "NO" — which side tokens to sell
		Price     float64 `json:"price"`     // sell price (0 < price < 1)
		SizeUSD   float64 `json:"size_usd"`  // USD value to close
		WalletID  string  `json:"wallet_id"`
		NegRisk   bool    `json:"neg_risk"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.TokenID == "" {
		writeError(w, http.StatusBadRequest, "token_id required")
		return
	}
	if req.Side != "YES" && req.Side != "NO" {
		writeError(w, http.StatusBadRequest, "side must be YES or NO")
		return
	}
	if req.Price <= 0 || req.Price >= 1 {
		writeError(w, http.StatusBadRequest, "price must be between 0.001 and 0.999")
		return
	}
	if req.SizeUSD <= 0 {
		writeError(w, http.StatusBadRequest, "size_usd must be positive")
		return
	}

	orderID, err := s.placer.PlaceOrder(req.WalletID, req.TokenID, req.Side, "GTC", req.Price, req.SizeUSD, req.NegRisk)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"order_id": orderID, "status": "closing"})
}

// handleCreateAlert handles POST /api/v1/alerts
func (s *Server) handleCreateAlert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	var req struct {
		ConditionID string  `json:"conditionId"`
		TokenID     string  `json:"tokenId"`
		Direction   string  `json:"direction"` // "above" or "below"
		Threshold   float64 `json:"threshold"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Direction != "above" && req.Direction != "below" {
		writeError(w, http.StatusBadRequest, "direction must be 'above' or 'below'")
		return
	}
	if req.Threshold < 0.01 || req.Threshold > 0.99 {
		writeError(w, http.StatusBadRequest, "threshold must be between 0.01 and 0.99")
		return
	}
	if s.mkts == nil {
		writeError(w, http.StatusServiceUnavailable, "markets service not running")
		return
	}
	id := s.mkts.AddAlert(markets.AlertRule{
		ConditionID: req.ConditionID,
		TokenID:     req.TokenID,
		Direction:   req.Direction,
		Threshold:   req.Threshold,
	})
	writeJSON(w, http.StatusOK, map[string]string{"id": id})
}
