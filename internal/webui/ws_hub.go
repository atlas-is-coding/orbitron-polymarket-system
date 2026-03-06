package webui

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/atlasdev/polytrade-bot/internal/tui"
)

// WsEvent is the JSON envelope sent to web clients.
type WsEvent struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type hub struct {
	mu      sync.Mutex
	clients map[string]chan<- []byte
}

func newHub() *hub {
	return &hub{clients: make(map[string]chan<- []byte)}
}

func (h *hub) register(id string, ch chan<- []byte) {
	h.mu.Lock()
	h.clients[id] = ch
	h.mu.Unlock()
}

func (h *hub) unregister(id string) {
	h.mu.Lock()
	delete(h.clients, id)
	h.mu.Unlock()
}

func (h *hub) broadcast(ev WsEvent) {
	raw, err := json.Marshal(ev)
	if err != nil {
		return
	}
	h.mu.Lock()
	for _, ch := range h.clients {
		select {
		case ch <- raw:
		default: // drop if client is slow
		}
	}
	h.mu.Unlock()
}

// consume reads from EventBus tap channel and broadcasts to WS clients,
// also updating state snapshot.
func (h *hub) consume(ctx context.Context, tap <-chan tea.Msg, state *WebState) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-tap:
			if !ok {
				return
			}
			h.handleMsg(msg, state)
		}
	}
}

func (h *hub) handleMsg(msg tea.Msg, state *WebState) {
	switch m := msg.(type) {
	case tui.BalanceMsg:
		state.SetBalance(m.USDC)
		h.broadcast(WsEvent{Type: "balance", Data: map[string]any{"usdc": m.USDC}})
	case tui.OrdersUpdateMsg:
		state.SetOrders(m.Rows)
		h.broadcast(WsEvent{Type: "orders", Data: m.Rows})
	case tui.PositionsUpdateMsg:
		state.SetPositions(m.Rows)
		h.broadcast(WsEvent{Type: "positions", Data: m.Rows})
	case tui.BotEventMsg:
		state.AddLog(m.Level, m.Message)
		h.broadcast(WsEvent{Type: "log", Data: LogEntry{Level: m.Level, Message: m.Message}})
	case tui.SubsystemStatusMsg:
		state.SetSubsystem(m.Name, m.Active)
		h.broadcast(WsEvent{Type: "subsystem", Data: map[string]any{"name": m.Name, "active": m.Active}})
	case tui.ConfigReloadedMsg:
		if m.Config != nil {
			state.SetConfig(m.Config)
		}
		h.broadcast(WsEvent{Type: "config_reloaded", Data: nil})

	case tui.WalletAddedMsg:
		e := WalletEntry{ID: m.ID, Label: m.Label, Enabled: m.Enabled}
		state.UpsertWallet(e)
		h.broadcast(WsEvent{Type: "wallet_added", Data: e})

	case tui.WalletRemovedMsg:
		state.RemoveWallet(m.ID)
		h.broadcast(WsEvent{Type: "wallet_removed", Data: map[string]string{"id": m.ID}})

	case tui.WalletChangedMsg:
		// Partial update: toggle enabled flag in existing entry
		wallets := state.Wallets()
		for _, w := range wallets {
			if w.ID == m.ID {
				w.Enabled = m.Enabled
				state.UpsertWallet(w)
				h.broadcast(WsEvent{Type: "wallet_changed", Data: map[string]any{"id": m.ID, "enabled": m.Enabled}})
				break
			}
		}

	case tui.WalletStatsMsg:
		e := WalletEntry{
			ID:          m.ID,
			Label:       m.Label,
			Enabled:     m.Enabled,
			BalanceUSD:  m.BalanceUSD,
			PnLUSD:      m.PnLUSD,
			OpenOrders:  m.OpenOrders,
			TotalTrades: m.TotalTrades,
		}
		state.UpsertWallet(e)
		h.broadcast(WsEvent{Type: "wallet_stats", Data: e})

	case tui.MarketAlertMsg:
		h.broadcast(WsEvent{Type: "market_alert", Data: map[string]any{
			"conditionId":  m.ConditionID,
			"question":     m.Question,
			"direction":    m.Direction,
			"threshold":    m.Threshold,
			"currentPrice": m.CurrentPrice,
		}})
	case tui.CopytradingTradeMsg:
		h.broadcast(WsEvent{Type: "copy_trade", Data: map[string]string{"line": m.Line}})
	}
}

// serveWS handles a WebSocket upgrade and pumps messages to the client.
func (h *hub) serveWS(w http.ResponseWriter, r *http.Request, clientID string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	ch := make(chan []byte, 64)
	h.register(clientID, ch)
	defer h.unregister(clientID)

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// read pump (detect disconnect)
	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				cancel()
				return
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case raw := <-ch:
			if err := conn.WriteMessage(websocket.TextMessage, raw); err != nil {
				return
			}
		}
	}
}
