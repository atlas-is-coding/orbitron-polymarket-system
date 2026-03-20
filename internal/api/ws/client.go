// Package ws реализует WebSocket клиент для Polymarket CLOB стримов.
// Base URL: wss://ws-subscriptions-clob.polymarket.com/ws/
//
// Каналы подписки:
//   - market: обновления книги ордеров (bids/asks)
//   - user:   события пользователя (ордера, трейды)
//   - asset:  цены токенов
package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

// ChannelType — тип WebSocket канала.
type ChannelType string

const (
	ChannelMarket ChannelType = "market"
	ChannelUser   ChannelType = "user"
	ChannelAsset  ChannelType = "asset"
)

// Message — входящее WebSocket сообщение.
type Message struct {
	EventType string          `json:"event_type"`
	Channel   ChannelType     `json:"channel"`
	AssetID   string          `json:"asset_id,omitempty"`
	Market    string          `json:"market,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
}

// SubscribeRequest — запрос на подписку.
type SubscribeRequest struct {
	Auth    *AuthPayload `json:"auth,omitempty"`
	Type    ChannelType  `json:"type"`
	// Markets/Assets use pointer-to-slice so that an empty slice serializes as []
	// (not omitted). Polymarket WS rejects subscriptions without the markets field.
	Markets *[]string `json:"markets,omitempty"`
	Assets  *[]string `json:"assets,omitempty"`
}

// AuthPayload — данные аутентификации для user channel.
type AuthPayload struct {
	APIKey     string `json:"apiKey"`
	Secret     string `json:"secret"`
	Passphrase string `json:"passphrase"`
}

// Handler — функция обработки входящих сообщений.
type Handler func(msg *Message)

// Client — WebSocket клиент с авто-переподключением.
type Client struct {
	url    string
	logger zerolog.Logger

	mu       sync.RWMutex
	conn     *websocket.Conn
	nextID   int
	subs     map[int]SubscribeRequest
	handlers map[int]Handler

	reconnectDelay time.Duration
	netDial        func(network, addr string) (net.Conn, error)
}

// NewClient создаёт WebSocket клиент.
func NewClient(wsURL string, log zerolog.Logger) *Client {
	return &Client{
		url:            wsURL,
		logger:         log.With().Str("component", "ws").Logger(),
		reconnectDelay: 3 * time.Second,
		subs:           make(map[int]SubscribeRequest),
		handlers:       make(map[int]Handler),
	}
}

// WithDialer sets a custom net dialer for WebSocket connections (e.g. SOCKS5/HTTP proxy).
func (c *Client) WithDialer(dial func(network, addr string) (net.Conn, error)) {
	c.netDial = dial
}

// Subscribe добавляет подписку и регистрирует обработчик сообщений.
// Все зарегистрированные обработчики вызываются для каждого входящего сообщения.
// Returns a subscription ID that can be passed to Unsubscribe().
func (c *Client) Subscribe(req SubscribeRequest, handler Handler) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	id := c.nextID
	c.nextID++
	c.subs[id] = req
	c.handlers[id] = handler
	return id
}

// Unsubscribe removes a subscription and its handler by ID.
// Safe to call with an ID that was already removed (no-op).
func (c *Client) Unsubscribe(id int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.subs, id)
	delete(c.handlers, id)
}

// Run запускает WebSocket клиент с авто-переподключением.
// Блокирует горутину до отмены ctx.
func (c *Client) Run(ctx context.Context) error {
	for {
		if err := c.connect(ctx); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			c.logger.Warn().Err(err).Dur("retry_in", c.reconnectDelay).Msg("ws disconnected, reconnecting")
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(c.reconnectDelay):
			}
		}
	}
}

func (c *Client) connect(ctx context.Context) error {
	c.mu.RLock()
	subs := make([]SubscribeRequest, 0, len(c.subs))
	for _, sub := range c.subs {
		subs = append(subs, sub)
	}
	c.mu.RUnlock()

	// Each Polymarket WS channel has its own URL path (/ws/market, /ws/user).
	// Derive it from the first subscription's type; fall back to the base URL.
	url := strings.TrimRight(c.url, "/")
	if len(subs) > 0 {
		url = url + "/" + string(subs[0].Type)
	}

	dialer := &websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
	if c.netDial != nil {
		dialer.NetDial = c.netDial
	}
	conn, _, err := dialer.DialContext(ctx, url, nil)
	if err != nil {
		return fmt.Errorf("ws: dial: %w", err)
	}
	defer conn.Close()

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	c.logger.Info().Str("url", url).Msg("ws connected")

	// Send subscriptions.
	for _, sub := range subs {
		data, err := json.Marshal(sub)
		if err != nil {
			return fmt.Errorf("ws: marshal subscribe: %w", err)
		}
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			return fmt.Errorf("ws: write subscribe: %w", err)
		}
	}

	// Heartbeat: server expects text "PING" every 10 seconds, replies "PONG".
	pingCtx, pingCancel := context.WithCancel(ctx)
	defer pingCancel()

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-pingCtx.Done():
				return
			case <-ticker.C:
				c.mu.RLock()
				conn := c.conn
				c.mu.RUnlock()
				if conn != nil {
					if err := conn.WriteMessage(websocket.TextMessage, []byte("PING")); err != nil {
						c.logger.Debug().Err(err).Msg("ws: ping failed")
					}
				}
			}
		}
	}()

	// Читаем сообщения
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("ws: read: %w", err)
		}

		// Skip non-JSON control messages (PONG, INVALID OPERATION, etc.)
		if len(data) == 0 || data[0] != '{' {
			c.logger.Debug().Str("raw", string(data)).Msg("ws: control message")
			continue
		}

		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			c.logger.Warn().Err(err).Str("raw", string(data)).Msg("ws: decode message")
			continue
		}

		c.mu.RLock()
		handlers := make([]Handler, 0, len(c.handlers))
		for _, h := range c.handlers {
			handlers = append(handlers, h)
		}
		c.mu.RUnlock()

		for _, h := range handlers {
			if h != nil {
				h(&msg)
			}
		}
	}
}

// Close закрывает WebSocket соединение.
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			c.logger.Debug().Err(err).Msg("ws: close error")
		}
		c.conn = nil
	}
}
