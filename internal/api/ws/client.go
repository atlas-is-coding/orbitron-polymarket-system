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
	Markets []string     `json:"markets,omitempty"`
	Assets  []string     `json:"assets,omitempty"`
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

	mu      sync.RWMutex
	conn    *websocket.Conn
	subs    []SubscribeRequest
	handler Handler

	reconnectDelay time.Duration
}

// NewClient создаёт WebSocket клиент.
func NewClient(wsURL string, log zerolog.Logger) *Client {
	return &Client{
		url:            wsURL,
		logger:         log.With().Str("component", "ws").Logger(),
		reconnectDelay: 3 * time.Second,
	}
}

// Subscribe добавляет подписку и регистрирует обработчик сообщений.
func (c *Client) Subscribe(req SubscribeRequest, handler Handler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.subs = append(c.subs, req)
	c.handler = handler
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
	dialer := websocket.DefaultDialer
	conn, _, err := dialer.DialContext(ctx, c.url, nil)
	if err != nil {
		return fmt.Errorf("ws: dial: %w", err)
	}
	defer conn.Close()

	c.mu.Lock()
	c.conn = conn
	subs := c.subs
	c.mu.Unlock()

	c.logger.Info().Str("url", c.url).Msg("ws connected")

	// Отправляем подписки
	for _, sub := range subs {
		data, err := json.Marshal(sub)
		if err != nil {
			return fmt.Errorf("ws: marshal subscribe: %w", err)
		}
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			return fmt.Errorf("ws: write subscribe: %w", err)
		}
	}

	// Пинг-горутина для поддержания соединения
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				c.mu.RLock()
				_ = conn.WriteMessage(websocket.PingMessage, nil)
				c.mu.RUnlock()
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

		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			c.logger.Warn().Err(err).Str("raw", string(data)).Msg("ws: decode message")
			continue
		}

		c.mu.RLock()
		handler := c.handler
		c.mu.RUnlock()

		if handler != nil {
			handler(&msg)
		}
	}
}

// Close закрывает WebSocket соединение.
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}
}
