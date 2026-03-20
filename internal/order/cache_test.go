package order

import (
	"context"
	"testing"

	"github.com/atlasdev/orbitron/internal/api/clob"
)

// TestOptimisticCache_ApplyOptimistic проверяет что MarkCanceled работает
// и IsOptimisticallyCanceled возвращает true
func TestOptimisticCache_ApplyOptimistic(t *testing.T) {
	cache := NewOptimisticCache()
	ctx := context.Background()

	orderID := "order-123"

	// Изначально ордер не помечен как отменённый
	isCanceled := cache.IsOptimisticallyCanceled(ctx, orderID)
	if isCanceled {
		t.Errorf("Expected order to not be optimistically canceled, got true")
	}

	// Применяем оптимистичную отмену
	cache.MarkCanceled(ctx, orderID)

	// Теперь ордер должен быть помечен как отменённый
	isCanceled = cache.IsOptimisticallyCanceled(ctx, orderID)
	if !isCanceled {
		t.Errorf("Expected order to be optimistically canceled, got false")
	}
}

// TestOptimisticCache_ConflictResolution проверяет что при конфликте
// (пользователь отменил, но API вернул MATCHED), API побеждает
func TestOptimisticCache_ConflictResolution(t *testing.T) {
	cache := NewOptimisticCache()
	ctx := context.Background()

	orderID := "order-456"

	// Применяем оптимистичную отмену
	cache.MarkCanceled(ctx, orderID)

	// Проверяем что ордер помечен как отменённый
	isCanceled := cache.IsOptimisticallyCanceled(ctx, orderID)
	if !isCanceled {
		t.Errorf("Expected order to be optimistically canceled after MarkCanceled")
	}

	// Теперь API возвращает, что ордер MATCHED (был исполнен)
	apiOrder := &clob.Order{
		ID:     orderID,
		Status: clob.StatusMatched,
		Side:   clob.SideBuy,
		Price:  "0.5",
	}

	// Reconcile с API ордером (API побеждает, оптимистичное состояние очищается)
	cache.Reconcile(ctx, orderID, apiOrder)

	// Проверяем что оптимистичное состояние очищено (API源 истины)
	isCanceled = cache.IsOptimisticallyCanceled(ctx, orderID)
	if isCanceled {
		t.Errorf("Expected order to not be optimistically canceled after Reconcile with MATCHED status")
	}

	// Проверяем что Get возвращает API ордер
	retrieved := cache.Get(ctx, orderID)
	if retrieved == nil {
		t.Errorf("Expected to retrieve order after Reconcile")
	}
	if retrieved.Status != clob.StatusMatched {
		t.Errorf("Expected order status to be MATCHED, got %v", retrieved.Status)
	}
}

// TestOptimisticCache_GetAll проверяет получение всех ордеров
func TestOptimisticCache_GetAll(t *testing.T) {
	cache := NewOptimisticCache()
	ctx := context.Background()

	// Добавляем несколько ордеров
	order1 := &clob.Order{
		ID:     "order-1",
		Status: clob.StatusLive,
		Side:   clob.SideBuy,
		Price:  "0.5",
	}
	order2 := &clob.Order{
		ID:     "order-2",
		Status: clob.StatusLive,
		Side:   clob.SideSell,
		Price:  "0.6",
	}

	cache.Reconcile(ctx, order1.ID, order1)
	cache.Reconcile(ctx, order2.ID, order2)

	// Получаем все ордера
	all := cache.GetAll(ctx)
	if len(all) != 2 {
		t.Errorf("Expected 2 orders, got %d", len(all))
	}

	// Проверяем что оба ордера в результате
	found := make(map[string]bool)
	for _, o := range all {
		found[o.ID] = true
	}

	if !found["order-1"] || !found["order-2"] {
		t.Errorf("Expected both orders to be in GetAll result")
	}
}

// TestOptimisticCache_Clear проверяет удаление ордера из кэша
func TestOptimisticCache_Clear(t *testing.T) {
	cache := NewOptimisticCache()
	ctx := context.Background()

	orderID := "order-789"
	apiOrder := &clob.Order{
		ID:     orderID,
		Status: clob.StatusLive,
		Side:   clob.SideBuy,
		Price:  "0.5",
	}

	cache.Reconcile(ctx, orderID, apiOrder)

	// Проверяем что ордер в кэше
	retrieved := cache.Get(ctx, orderID)
	if retrieved == nil {
		t.Errorf("Expected order to be in cache after Reconcile")
	}

	// Удаляем ордер из кэша
	cache.Clear(ctx, orderID)

	// Проверяем что ордера больше нет в кэше
	retrieved = cache.Get(ctx, orderID)
	if retrieved != nil {
		t.Errorf("Expected order to be removed from cache after Clear")
	}
}
