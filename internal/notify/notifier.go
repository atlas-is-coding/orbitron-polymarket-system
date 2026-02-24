// Package notify определяет интерфейс уведомлений.
// Добавление нового канала (Slack, Discord) — реализовать Notifier и зарегистрировать.
package notify

import "context"

// Notifier — интерфейс отправки уведомлений.
type Notifier interface {
	Send(ctx context.Context, msg string) error
}

// NoopNotifier — пустой notifier (используется когда уведомления отключены).
type NoopNotifier struct{}

func (n *NoopNotifier) Send(_ context.Context, _ string) error { return nil }
