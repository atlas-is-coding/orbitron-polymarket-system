package notification

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/atlasdev/orbitron/internal/storage"
)

// mockNotifier is a test double for the Notifier interface
type mockNotifier struct {
	sent  []string
	fails int
}

func (m *mockNotifier) Send(ctx context.Context, msg string) error {
	if m.fails > 0 {
		m.fails--
		return errors.New("send failed")
	}
	m.sent = append(m.sent, msg)
	return nil
}

// mockStore is a test double for NotificationQueueStore
type mockStore struct {
	notifications map[string]*storage.Notification
}

func newMockStore() *mockStore {
	return &mockStore{
		notifications: make(map[string]*storage.Notification),
	}
}

func (m *mockStore) EnqueueNotification(ctx context.Context, notif *storage.Notification) error {
	m.notifications[notif.ID] = notif
	return nil
}

func (m *mockStore) GetPendingNotifications(ctx context.Context, walletAddress string) ([]*storage.Notification, error) {
	var pending []*storage.Notification
	for _, n := range m.notifications {
		// If walletAddress is empty, return all pending; otherwise filter by wallet
		if (walletAddress == "" || n.WalletAddress == walletAddress) && n.Status == "PENDING" {
			pending = append(pending, n)
		}
	}
	return pending, nil
}

func (m *mockStore) UpdateNotificationStatus(ctx context.Context, id, status string, retryCount int, nextRetryAt *time.Time) error {
	if n, ok := m.notifications[id]; ok {
		n.Status = status
		n.RetryCount = retryCount
		n.NextRetryAt = nextRetryAt
		n.UpdatedAt = time.Now()
	}
	return nil
}

func (m *mockStore) DeleteNotification(ctx context.Context, id string) error {
	delete(m.notifications, id)
	return nil
}

// mockLogger is a test double for logger
type mockLogger struct {
	logs []string
}

func newMockLogger() *mockLogger {
	return &mockLogger{}
}

func (m *mockLogger) Info(msg string, fields ...interface{}) {
	m.logs = append(m.logs, fmt.Sprintf("INFO: %s", msg))
}

func (m *mockLogger) Error(msg string, fields ...interface{}) {
	m.logs = append(m.logs, fmt.Sprintf("ERROR: %s", msg))
}

func TestNotificationQueue_DeliveryRetry(t *testing.T) {
	ctx := context.Background()

	notifier := &mockNotifier{fails: 2} // Fail first 2 times
	store := newMockStore()
	logger := newMockLogger()

	queue := NewQueue(notifier, store, logger)

	// Create and enqueue a notification
	notif := &storage.Notification{
		ID:            "test-notif-1",
		WalletAddress: "0xtest",
		EventType:     "ORDER_PLACED",
		Payload:       `{"orderId":"123"}`,
		Status:        "PENDING",
		RetryCount:    0,
		MaxRetries:    5,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := queue.Enqueue(ctx, notif)
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	// First attempt - should fail
	queue.ProcessPending(ctx)
	storedNotif := store.notifications["test-notif-1"]
	if storedNotif.RetryCount != 1 {
		t.Errorf("After first attempt, RetryCount should be 1, got %d", storedNotif.RetryCount)
	}
	if storedNotif.Status != "PENDING" {
		t.Errorf("After first failure, Status should be PENDING, got %s", storedNotif.Status)
	}

	// Second attempt - should fail
	// Reset NextRetryAt to allow immediate retry
	storedNotif.NextRetryAt = nil
	queue.ProcessPending(ctx)
	storedNotif = store.notifications["test-notif-1"]
	if storedNotif.RetryCount != 2 {
		t.Errorf("After second attempt, RetryCount should be 2, got %d", storedNotif.RetryCount)
	}

	// Third attempt - should succeed
	// Reset NextRetryAt to allow immediate retry
	storedNotif.NextRetryAt = nil
	queue.ProcessPending(ctx)
	storedNotif = store.notifications["test-notif-1"]
	if storedNotif.RetryCount != 3 {
		t.Errorf("After third attempt, RetryCount should be 3, got %d", storedNotif.RetryCount)
	}
	if storedNotif.Status != "DELIVERED" {
		t.Errorf("After successful delivery, Status should be DELIVERED, got %s", storedNotif.Status)
	}

	// Check that the message was sent exactly once
	if len(notifier.sent) != 1 {
		t.Errorf("Expected 1 message sent, got %d", len(notifier.sent))
	}
}
