package webui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/atlasdev/polytrade-bot/internal/api/gamma"
	"github.com/atlasdev/polytrade-bot/internal/auth"
	"github.com/atlasdev/polytrade-bot/internal/config"
	"github.com/atlasdev/polytrade-bot/internal/i18n"
	"github.com/atlasdev/polytrade-bot/internal/markets"
	"github.com/atlasdev/polytrade-bot/internal/tui"
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
}

// MarketsProvider exposes markets data to the Web UI.
type MarketsProvider interface {
	GetByTag(slug string) []gamma.Market
	GetMarket(conditionID string) (gamma.Market, bool)
	Tags() []gamma.Tag
	AddAlert(rule markets.AlertRule) string
}

// Server is the Web UI HTTP server.
// NOTE: Additional fields (log, embed fs wiring) added in server.go (Task 6).
type Server struct {
	cfg      *config.Config
	cfgMu    sync.RWMutex
	cfgPath  string
	password string
	bus      *tui.EventBus
	canceler OrderCanceler
	wallets  WalletMutator // may be nil
	adder    WalletAdder   // may be nil
	mkts     MarketsProvider // may be nil
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
	writeJSON(w, http.StatusOK, map[string]any{
		"balance":    s.state.Balance(),
		"subsystems": s.state.Subsystems(),
		"orders":     len(s.state.Orders()),
		"positions":  len(s.state.Positions()),
	})
}

func (s *Server) handleOrders(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.state.Orders())
}

func (s *Server) handlePositions(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.state.Positions())
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
	// Mask secrets
	cfg.Auth.PrivateKey = maskSecret(cfg.Auth.PrivateKey)
	cfg.Auth.APIKey = maskSecret(cfg.Auth.APIKey)
	cfg.Auth.APISecret = maskSecret(cfg.Auth.APISecret)
	cfg.Auth.Passphrase = maskSecret(cfg.Auth.Passphrase)
	cfg.Telegram.BotToken = maskSecret(cfg.Telegram.BotToken)
	cfg.WebUI.JWTSecret = maskSecret(cfg.WebUI.JWTSecret)
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

// handleGetWallets returns the cached wallet list from WebState.
func (s *Server) handleGetWallets(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.state.Wallets())
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

	if s.adder != nil {
		s.adder.AddInactive(wCfg)
	}
	if s.bus != nil {
		s.bus.Send(tui.WalletAddedMsg{ID: id, Label: wCfg.Label, Enabled: true})
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"id":          id,
		"address":     addr,
		"label":       wCfg.Label,
		"enabled":     true,
		"balance_usd": 0,
		"pnl_usd":     0,
		"open_orders": 0,
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
	writeJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

// ── Markets handlers ──────────────────────────────────────────────────────────

// handleMarketsList handles GET /api/v1/markets?tag=crypto&limit=50&offset=0
func (s *Server) handleMarketsList(w http.ResponseWriter, r *http.Request) {
	tag := r.URL.Query().Get("tag")
	var result []gamma.Market
	if s.mkts != nil {
		result = s.mkts.GetByTag(tag)
	}
	if result == nil {
		result = []gamma.Market{}
	}
	writeJSON(w, http.StatusOK, result)
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

// handleMarketDetail handles GET /api/v1/markets/{conditionID}
func (s *Server) handleMarketDetail(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/markets/")
	if id == "" || id == "tags" {
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
