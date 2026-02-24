// Package trading реализует торговый движок и интерфейс стратегий.
package trading

import "context"

// Strategy — интерфейс торговой стратегии.
// Каждая стратегия работает в своей горутине и управляется Engine.
type Strategy interface {
	// Name возвращает уникальное имя стратегии.
	Name() string
	// Start запускает стратегию. Блокирует горутину до отмены ctx или ошибки.
	Start(ctx context.Context) error
	// Stop плавно останавливает стратегию.
	Stop() error
}
