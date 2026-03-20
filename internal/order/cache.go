package order

import (
	"context"
	"sync"
	"time"

	"github.com/atlasdev/orbitron/internal/api/clob"
)

// OptimisticOrder хранит состояние ордера с оптимистичными обновлениями.
// APIOrder - источник истины из API.
// Optimistic - локальные оптимистичные изменения (ключи: "canceled", и т.д.)
type OptimisticOrder struct {
	APIOrder   *clob.Order
	Optimistic map[string]bool // ключи: "canceled"
	LastUpdate int64           // время последнего обновления в миллисекундах
}

// OptimisticCache хранит кэш ордеров с поддержкой оптимистичных обновлений.
// Потокобезопасен благодаря sync.RWMutex.
type OptimisticCache struct {
	mu    sync.RWMutex
	cache map[string]*OptimisticOrder
}

// NewOptimisticCache создаёт новый оптимистичный кэш.
func NewOptimisticCache() *OptimisticCache {
	return &OptimisticCache{
		cache: make(map[string]*OptimisticOrder),
	}
}

// MarkCanceled применяет оптимистичную отмену к ордеру.
// Это показывает локально, что ордер был отменён, даже до подтверждения API.
func (oc *OptimisticCache) MarkCanceled(ctx context.Context, orderID string) {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	// Если ордера нет в кэше, создаём новую запись с оптимистичным состоянием
	if _, exists := oc.cache[orderID]; !exists {
		oc.cache[orderID] = &OptimisticOrder{
			APIOrder:   nil,
			Optimistic: make(map[string]bool),
			LastUpdate: time.Now().UnixMilli(),
		}
	}

	// Помечаем как "canceled"
	oc.cache[orderID].Optimistic["canceled"] = true
	oc.cache[orderID].LastUpdate = time.Now().UnixMilli()
}

// IsOptimisticallyCanceled проверяет, помечен ли ордер как оптимистично отменённый.
func (oc *OptimisticCache) IsOptimisticallyCanceled(ctx context.Context, orderID string) bool {
	oc.mu.RLock()
	defer oc.mu.RUnlock()

	if order, exists := oc.cache[orderID]; exists {
		return order.Optimistic["canceled"]
	}
	return false
}

// Reconcile обновляет кэш с данными из API.
// API является источником истины, поэтому оптимистичное состояние очищается.
func (oc *OptimisticCache) Reconcile(ctx context.Context, orderID string, apiOrder *clob.Order) {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	// Заменяем или создаём запись в кэше
	oc.cache[orderID] = &OptimisticOrder{
		APIOrder:   apiOrder,
		Optimistic: make(map[string]bool), // Очищаем оптимистичное состояние
		LastUpdate: time.Now().UnixMilli(),
	}
}

// Get возвращает текущее состояние ордера (с учётом оптимистичных обновлений).
// Возвращает nil если ордера нет в кэше.
func (oc *OptimisticCache) Get(ctx context.Context, orderID string) *clob.Order {
	oc.mu.RLock()
	defer oc.mu.RUnlock()

	if order, exists := oc.cache[orderID]; exists {
		return order.APIOrder
	}
	return nil
}

// GetAll возвращает все ордера из кэша.
func (oc *OptimisticCache) GetAll(ctx context.Context) []*clob.Order {
	oc.mu.RLock()
	defer oc.mu.RUnlock()

	var orders []*clob.Order
	for _, opt := range oc.cache {
		if opt.APIOrder != nil {
			orders = append(orders, opt.APIOrder)
		}
	}
	return orders
}

// Clear удаляет ордер из кэша.
func (oc *OptimisticCache) Clear(ctx context.Context, orderID string) {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	delete(oc.cache, orderID)
}
