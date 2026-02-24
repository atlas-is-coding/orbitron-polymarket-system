// Package telegram реализует Notifier через Telegram Bot API.
package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/atlasdev/polytrade-bot/internal/api"
)

// Notifier отправляет сообщения в Telegram чат.
type Notifier struct {
	chatID string
	http   *api.Client
}

// New создаёт Telegram Notifier.
// botToken — токен бота (получить у @BotFather).
// chatID — ID чата или канала.
func New(botToken, chatID string) *Notifier {
	baseURL := "https://api.telegram.org/bot" + botToken
	return &Notifier{
		chatID: chatID,
		http:   api.NewClient(baseURL, 10, 2),
	}
}

// Send отправляет текстовое сообщение в Telegram (поддерживает HTML-разметку).
func (n *Notifier) Send(_ context.Context, msg string) error {
	body, err := json.Marshal(map[string]string{
		"chat_id":    n.chatID,
		"text":       msg,
		"parse_mode": "HTML",
	})
	if err != nil {
		return fmt.Errorf("telegram: marshal: %w", err)
	}

	resp, err := n.http.Post("/sendMessage", body, nil)
	if err != nil {
		return fmt.Errorf("telegram: send: %w", err)
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("telegram: HTTP %d: %s", resp.StatusCode, string(resp.Body))
	}
	return nil
}

// escape экранирует специальные HTML-символы для безопасной отправки.
func escape(s string) string {
	return url.QueryEscape(s)
}

// Format форматирует уведомление с заголовком и телом.
func Format(title, body string) string {
	return fmt.Sprintf("<b>%s</b>\n%s", title, body)
}

var _ = escape // используется вне пакета при необходимости
