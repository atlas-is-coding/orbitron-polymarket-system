package monitor

import (
	"context"
	"testing"
	"time"

	"github.com/atlasdev/orbitron/internal/api/clob"
	"github.com/atlasdev/orbitron/internal/notification"
	"github.com/atlasdev/orbitron/internal/storage"
	"github.com/atlasdev/orbitron/internal/storage/sqlite"
	"github.com/rs/zerolog"
)

// MockCLOBClient для тестирования
type mockCLOBClient struct {
	orders []clob.Order
}

func (m *mockCLOBClient) GetDataOrders(filter clob.OrdersFilter) (*clob.OrdersResponse, error) {
	return &clob.OrdersResponse{Data: m.orders}, nil
}

func (m *mockCLOBClient) GetTrades(filter clob.TradesFilter) (*clob.TradesResponse, error) {
	return &clob.TradesResponse{Data: []clob.Trade{}}, nil
}

func (m *mockCLOBClient) CancelOrder(id string) (*clob.CancelOrderResponse, error) {
	for i := range m.orders {
		if m.orders[i].ID == id {
			m.orders[i].Status = clob.StatusCanceled
		}
	}
	return &clob.CancelOrderResponse{Canceled: true}, nil
}

// MockNotifier для тестирования
type mockNotifier struct {
	sent []string
}

func (m *mockNotifier) Send(ctx context.Context, msg string) error {
	m.sent = append(m.sent, msg)
	return nil
}

func TestOrdersSystemIntegration(t *testing.T) {
	// Setup: создай in-memory БД
	db, err := sqlite.Open(":memory:")
	if err != nil {
		t.Fatalf("create db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	notifier := &mockNotifier{}

	// Test case 1: User places an order (it's received from API)
	order1 := clob.Order{
		ID:            "order-1",
		Status:        clob.StatusLive,
		Side:          clob.SideBuy,
		Price:         "0.65",
		AssetID:       "token-123",
		OriginalSize:  "100.0",
		SizeFilled:    "0",
		SizeRemaining: "100.0",
		CreatedAt:     time.Now().UnixMilli(),
	}

	// Persist to database
	err = db.InsertOrder(ctx, &storage.Order{
		ID:            order1.ID,
		WalletAddress: "0xtest",
		AssetID:       order1.AssetID,
		Side:          string(order1.Side),
		OrderType:     string(clob.OrderTypeGTC),
		Price:         0.65,
		Size:          100.0,
		Status:        string(order1.Status),
		CreatedAt:     time.UnixMilli(order1.CreatedAt),
		UpdatedAt:     time.UnixMilli(order1.CreatedAt),
	})
	if err != nil {
		t.Fatalf("insert order: %v", err)
	}

	// Test case 2: Verify order was persisted
	retrieved, err := db.GetOrder(ctx, order1.ID)
	if err != nil {
		t.Fatalf("get order: %v", err)
	}
	if retrieved == nil {
		t.Fatal("expected order to be retrieved from database")
	}
	if retrieved.Status != string(order1.Status) {
		t.Errorf("expected status %v, got %v", order1.Status, retrieved.Status)
	}

	// Test case 3: Update order status (simulating cancellation)
	canceledOrder := order1
	canceledOrder.Status = clob.StatusCanceled

	err = db.UpdateOrder(ctx, &storage.Order{
		ID:            canceledOrder.ID,
		WalletAddress: "0xtest",
		AssetID:       canceledOrder.AssetID,
		Side:          string(canceledOrder.Side),
		OrderType:     string(clob.OrderTypeGTC),
		Price:         0.65,
		Size:          100.0,
		Status:        string(canceledOrder.Status),
		CreatedAt:     time.UnixMilli(canceledOrder.CreatedAt),
		UpdatedAt:     time.Now(),
	})
	if err != nil {
		t.Fatalf("update order: %v", err)
	}

	// Verify database state
	retrieved, err = db.GetOrder(ctx, order1.ID)
	if err != nil {
		t.Fatalf("get order: %v", err)
	}
	if retrieved == nil || retrieved.Status != string(clob.StatusCanceled) {
		t.Errorf("expected CANCELED, got %v", retrieved)
	}

	// Test case 4: Queue notification
	notifQueue := notification.NewQueue(notifier, db, zerolog.Nop())
	walletAddress := "0xtest"
	err = notifQueue.Enqueue(ctx, &storage.Notification{
		WalletAddress: walletAddress,
		EventType:     "order_canceled",
		Payload:       "Order canceled: order-1",
	})
	if err != nil {
		t.Fatalf("enqueue notification: %v", err)
	}

	// Test case 5: Verify notification was persisted to database
	pending, err := db.GetPendingNotifications(ctx, walletAddress)
	if err != nil {
		t.Fatalf("get pending notifications: %v", err)
	}
	if len(pending) != 1 {
		t.Errorf("expected 1 pending notification, got %d", len(pending))
	}
	if pending[0].EventType != "order_canceled" {
		t.Errorf("expected event type 'order_canceled', got %s", pending[0].EventType)
	}

	// Test case 6: Verify notification can be processed and sent
	for _, notif := range pending {
		err := notifier.Send(ctx, notif.Payload)
		if err != nil {
			t.Fatalf("send notification: %v", err)
		}
		// Mark as delivered
		err = db.UpdateNotificationStatus(ctx, notif.ID, "DELIVERED", 1, nil)
		if err != nil {
			t.Fatalf("update notification status: %v", err)
		}
	}

	// Verify notification was delivered
	if len(notifier.sent) != 1 {
		t.Errorf("expected 1 notification sent, got %d", len(notifier.sent))
	}
	if notifier.sent[0] != "Order canceled: order-1" {
		t.Errorf("expected message 'Order canceled: order-1', got %s", notifier.sent[0])
	}

	t.Logf("Integration test passed: order persistence, status update, and notification queue all working")
}
