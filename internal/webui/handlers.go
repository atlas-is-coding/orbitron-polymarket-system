package webui

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/atlasdev/polytrade-bot/internal/config"
	"github.com/atlasdev/polytrade-bot/internal/i18n"
	"github.com/atlasdev/polytrade-bot/internal/tui"
)

// OrderCanceler wraps TradesMonitor cancel operations.
type OrderCanceler interface {
	CancelOrder(id string) error
	CancelAllOrders() error
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

func (s *Server) handleOverview(w http.ResponseWriter, r *http.Request) {
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
	traders := make([]config.TraderConfig, len(s.cfg.Copytrading.Traders))
	copy(traders, s.cfg.Copytrading.Traders)
	s.cfgMu.RUnlock()
	writeJSON(w, http.StatusOK, map[string]any{
		"enabled": s.cfg.Copytrading.Enabled,
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
