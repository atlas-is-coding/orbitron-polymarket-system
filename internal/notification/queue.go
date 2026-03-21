package notification

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/atlasdev/orbitron/internal/storage"
	"github.com/rs/zerolog"
)

const (
	maxRetries        = 5
	initialBackoffMs  = 1000
	maxBackoffMs      = 300000
	backoffMultiplier = 2.0
)

// Notifier defines the interface for sending notifications
type Notifier interface {
	Send(ctx context.Context, msg string) error
}

// Logger defines the interface for logging
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
}

// zerologAdapter adapts zerolog.Logger to Logger interface
type zerologAdapter struct {
	logger zerolog.Logger
}

func (z *zerologAdapter) Info(msg string, fields ...interface{}) {
	z.logger.Info().Msg(msg)
}

func (z *zerologAdapter) Error(msg string, fields ...interface{}) {
	z.logger.Error().Msg(msg)
}

// Queue manages queuing and retry logic for notifications
type Queue struct {
	notifier   Notifier
	store      storage.NotificationQueueStore
	logger     Logger
	mu         sync.Mutex
	processing bool
}

// NewQueue creates a new notification queue
func NewQueue(notif Notifier, store storage.NotificationQueueStore, logger interface{}) *Queue {
	var l Logger
	switch v := logger.(type) {
	case Logger:
		l = v
	case zerolog.Logger:
		l = &zerologAdapter{logger: v}
	default:
		l = &zerologAdapter{logger: zerolog.Nop()}
	}
	return &Queue{
		notifier: notif,
		store:    store,
		logger:   l,
	}
}

// Enqueue adds a notification to the queue
func (q *Queue) Enqueue(ctx context.Context, notif *storage.Notification) error {
	if notif.MaxRetries == 0 {
		notif.MaxRetries = maxRetries
	}
	if notif.Status == "" {
		notif.Status = "PENDING"
	}
	if notif.CreatedAt.IsZero() {
		notif.CreatedAt = time.Now()
	}
	if notif.UpdatedAt.IsZero() {
		notif.UpdatedAt = time.Now()
	}

	return q.store.EnqueueNotification(ctx, notif)
}

// ProcessPending processes all pending notifications with exponential backoff retry
func (q *Queue) ProcessPending(ctx context.Context) error {
	q.mu.Lock()
	if q.processing {
		q.mu.Unlock()
		return nil
	}
	q.processing = true
	q.mu.Unlock()

	defer func() {
		q.mu.Lock()
		q.processing = false
		q.mu.Unlock()
	}()

	// Get all pending notifications for all wallets
	// Note: We fetch for empty wallet address, assuming store implementation handles this
	notifications, err := q.store.GetPendingNotifications(ctx, "")
	if err != nil {
		q.logger.Error("failed to get pending notifications")
		return err
	}

	for _, notif := range notifications {
		if err := q.processNotification(ctx, notif); err != nil {
			q.logger.Error("error processing notification")
		}
	}

	return nil
}

// processNotification attempts to deliver a single notification
func (q *Queue) processNotification(ctx context.Context, notif *storage.Notification) error {
	// Check if we should retry
	if !q.shouldRetry(notif) {
		q.logger.Error("notification exceeded max retries")
		return q.store.UpdateNotificationStatus(ctx, notif.ID, "FAILED", notif.RetryCount, nil)
	}

	// Check if it's time to retry (for backoff)
	if notif.NextRetryAt != nil && notif.NextRetryAt.After(time.Now()) {
		return nil // Not yet time to retry
	}

	// Attempt to send
	err := q.notifier.Send(ctx, notif.Payload)
	notif.RetryCount++

	if err == nil {
		// Success
		q.logger.Info("notification delivered successfully")
		return q.store.UpdateNotificationStatus(ctx, notif.ID, "DELIVERED", notif.RetryCount, nil)
	}

	// Failure - calculate backoff and schedule retry
	q.logger.Error("notification send failed, scheduling retry")

	backoffMs := q.calcBackoff(notif.RetryCount)
	nextRetryAt := time.Now().Add(time.Duration(backoffMs) * time.Millisecond)

	return q.store.UpdateNotificationStatus(ctx, notif.ID, "PENDING", notif.RetryCount, &nextRetryAt)
}

// shouldRetry checks if the notification should be retried
func (q *Queue) shouldRetry(notif *storage.Notification) bool {
	return notif.RetryCount < notif.MaxRetries
}

// calcBackoff calculates exponential backoff delay in milliseconds
func (q *Queue) calcBackoff(attemptCount int) int {
	// backoff = initialBackoff * (multiplier ^ attemptCount)
	backoffMs := float64(initialBackoffMs) * math.Pow(backoffMultiplier, float64(attemptCount))

	// Cap at maxBackoffMs
	if backoffMs > float64(maxBackoffMs) {
		backoffMs = float64(maxBackoffMs)
	}

	return int(backoffMs)
}
