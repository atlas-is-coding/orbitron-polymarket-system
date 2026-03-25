package nexus

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

// MockAuditLog implements a no-op AuditLog for testing
type MockAuditLog struct {
	savedCommands []*Command
	savedEvents   []*Event
	mu            sync.Mutex
}

func (m *MockAuditLog) SaveCommand(ctx context.Context, cmd *Command) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.savedCommands = append(m.savedCommands, cmd)
	return nil
}

func (m *MockAuditLog) SaveEvent(ctx context.Context, event *Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.savedEvents = append(m.savedEvents, event)
	return nil
}

func (m *MockAuditLog) GetCommandHistory(ctx context.Context, limit int) ([]*Command, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if limit <= 0 || limit > len(m.savedCommands) {
		limit = len(m.savedCommands)
	}
	result := make([]*Command, limit)
	copy(result, m.savedCommands[:limit])
	return result, nil
}

func (m *MockAuditLog) GetEventHistory(ctx context.Context, limit int) ([]*Event, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if limit <= 0 || limit > len(m.savedEvents) {
		limit = len(m.savedEvents)
	}
	result := make([]*Event, limit)
	copy(result, m.savedEvents[:limit])
	return result, nil
}

func newTestLogger() zerolog.Logger {
	return zerolog.New(zerolog.NewTestWriter(&testing.T{})).With().Timestamp().Logger()
}

func TestCommandProcessorSyncExecute(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	auditLog := &MockAuditLog{}
	cp := NewCommandProcessor(ctx, auditLog, 1, newTestLogger())
	defer cp.Close()

	handlerCalled := atomic.Bool{}
	cp.RegisterHandler(CommandPlaceOrder, func(ctx context.Context, cmd *Command) (interface{}, error) {
		handlerCalled.Store(true)
		return map[string]string{"order_id": "test-123"}, nil
	})

	cmd := &Command{
		Type:      CommandPlaceOrder,
		Timestamp: time.Now(),
		Timeout:   5 * time.Second,
		Payload:   PlaceOrderPayload{},
	}

	result, err := cp.Execute(context.Background(), cmd)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected command result, got nil")
	}

	if result.Status != StatusCompleted {
		t.Fatalf("Expected status %s, got %s", StatusCompleted, result.Status)
	}

	if !handlerCalled.Load() {
		t.Fatal("Handler was not called")
	}

	if result.ID == "" {
		t.Fatal("Expected command ID to be set")
	}
}

func TestCommandProcessorAsyncExecute(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	auditLog := &MockAuditLog{}
	cp := NewCommandProcessor(ctx, auditLog, 2, newTestLogger())
	defer cp.Close()

	handlerCalled := atomic.Bool{}
	cp.RegisterHandler(CommandCancelOrder, func(ctx context.Context, cmd *Command) (interface{}, error) {
		handlerCalled.Store(true)
		time.Sleep(100 * time.Millisecond) // simulate work
		return "canceled", nil
	})

	cmd := &Command{
		Type:      CommandCancelOrder,
		Timestamp: time.Now(),
		Timeout:   5 * time.Second,
		Payload:   CancelOrderPayload{},
	}

	cmdID, err := cp.ExecuteAsync(context.Background(), cmd)
	if err != nil {
		t.Fatalf("ExecuteAsync failed: %v", err)
	}

	if cmdID == "" {
		t.Fatal("Expected command ID, got empty string")
	}

	// Poll for completion
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	var finalCmd *Command
	for {
		select {
		case <-timeout:
			t.Fatal("Timeout waiting for command completion")
		case <-ticker.C:
			c, err := cp.GetStatus(cmdID)
			if err == nil && c != nil && c.Status == StatusCompleted {
				finalCmd = c
				goto done
			}
		}
	}

done:
	if finalCmd == nil {
		t.Fatal("Command never reached completed status")
	}

	if finalCmd.Status != StatusCompleted {
		t.Fatalf("Expected status %s, got %s", StatusCompleted, finalCmd.Status)
	}

	if !handlerCalled.Load() {
		t.Fatal("Handler was not called")
	}
}

func TestCommandProcessorTimeout(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	auditLog := &MockAuditLog{}
	cp := NewCommandProcessor(ctx, auditLog, 1, newTestLogger())
	defer cp.Close()

	cp.RegisterHandler(CommandStartStrategy, func(ctx context.Context, cmd *Command) (interface{}, error) {
		// Block forever unless context is cancelled
		<-ctx.Done()
		return nil, ctx.Err()
	})

	cmd := &Command{
		Type:      CommandStartStrategy,
		Timestamp: time.Now(),
		Timeout:   50 * time.Millisecond,
		Payload:   StartStrategyPayload{},
	}

	result, _ := cp.Execute(context.Background(), cmd)
	if result == nil {
		t.Fatal("Expected command result, got nil")
	}

	if result.Status != StatusTimedOut && result.Status != StatusFailed {
		t.Fatalf("Expected status TimedOut or Failed, got %s", result.Status)
	}

	if result.Error == "" {
		t.Fatal("Expected error message in command")
	}
}

func TestCommandProcessorHandlerError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	auditLog := &MockAuditLog{}
	cp := NewCommandProcessor(ctx, auditLog, 1, newTestLogger())
	defer cp.Close()

	expectedErr := "wallet not found"
	cp.RegisterHandler(CommandAddWallet, func(ctx context.Context, cmd *Command) (interface{}, error) {
		return nil, errors.New(expectedErr)
	})

	cmd := &Command{
		Type:      CommandAddWallet,
		Timestamp: time.Now(),
		Timeout:   5 * time.Second,
		Payload:   AddWalletPayload{},
	}

	result, err := cp.Execute(context.Background(), cmd)
	if result == nil {
		t.Fatal("Expected command result, got nil")
	}

	if result.Status != StatusFailed {
		t.Fatalf("Expected status %s, got %s", StatusFailed, result.Status)
	}

	if result.Error != expectedErr {
		t.Fatalf("Expected error %q, got %q", expectedErr, result.Error)
	}

	if err == nil {
		t.Fatal("Expected Execute to return error when handler fails")
	}
}

func TestCommandProcessorConcurrentCommands(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	auditLog := &MockAuditLog{}
	cp := NewCommandProcessor(ctx, auditLog, 4, newTestLogger())
	defer cp.Close()

	handlerCount := atomic.Int32{}
	cp.RegisterHandler(CommandRemoveWallet, func(ctx context.Context, cmd *Command) (interface{}, error) {
		handlerCount.Add(1)
		time.Sleep(50 * time.Millisecond)
		return nil, nil
	})

	var wg sync.WaitGroup
	numCommands := 10

	for i := 0; i < numCommands; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			cmd := &Command{
				Type:      CommandRemoveWallet,
				Timestamp: time.Now(),
				Timeout:   5 * time.Second,
				Payload:   RemoveWalletPayload{},
			}
			_, err := cp.ExecuteAsync(context.Background(), cmd)
			if err != nil {
				t.Errorf("ExecuteAsync failed: %v", err)
			}
		}(i)
	}

	wg.Wait()

	// Poll for all commands to complete
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatalf("Timeout waiting for commands, only %d completed", handlerCount.Load())
		case <-ticker.C:
			if handlerCount.Load() >= int32(numCommands) {
				goto done
			}
		}
	}

done:
	if handlerCount.Load() != int32(numCommands) {
		t.Fatalf("Expected %d handlers called, got %d", numCommands, handlerCount.Load())
	}

	stats := cp.Stats()
	if stats == nil {
		t.Fatal("Stats returned nil")
	}
}

func TestCommandProcessorInFlightTracking(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	auditLog := &MockAuditLog{}
	cp := NewCommandProcessor(ctx, auditLog, 1, newTestLogger())
	defer cp.Close()

	handlerStarted := make(chan struct{})
	handlerDone := make(chan struct{})

	cp.RegisterHandler(CommandToggleWallet, func(ctx context.Context, cmd *Command) (interface{}, error) {
		close(handlerStarted)
		<-handlerDone
		return nil, nil
	})

	cmd := &Command{
		Type:      CommandToggleWallet,
		Timestamp: time.Now(),
		Timeout:   5 * time.Second,
		Payload:   ToggleWalletPayload{},
	}

	cmdID, err := cp.ExecuteAsync(context.Background(), cmd)
	if err != nil {
		t.Fatalf("ExecuteAsync failed: %v", err)
	}

	// Wait for handler to start
	<-handlerStarted

	// While handler is running, GetStatus should work
	c, err := cp.GetStatus(cmdID)
	if err != nil {
		t.Fatalf("GetStatus failed while command in flight: %v", err)
	}

	if c == nil {
		t.Fatal("Expected command from GetStatus, got nil")
	}

	if c.Status != StatusProcessing && c.Status != StatusPending {
		t.Fatalf("Expected status Processing or Pending, got %s", c.Status)
	}

	// Signal handler to complete
	close(handlerDone)

	// Poll until completion
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatal("Timeout waiting for command completion")
		case <-ticker.C:
			c, err := cp.GetStatus(cmdID)
			if err == nil && c != nil && c.Status == StatusCompleted {
				goto done
			}
		}
	}

done:
	// After completion, GetStatus should still work briefly then fail
	// or return command if still in cache
	finalCmd, _ := cp.GetStatus(cmdID)
	if finalCmd != nil && finalCmd.Status != StatusCompleted {
		t.Fatalf("Expected status Completed after handler finishes, got %s", finalCmd.Status)
	}
}

func TestCommandProcessorHandlerNotFound(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	auditLog := &MockAuditLog{}
	cp := NewCommandProcessor(ctx, auditLog, 1, newTestLogger())
	defer cp.Close()

	cmd := &Command{
		Type:      CommandType("unknown.command"),
		Timestamp: time.Now(),
		Timeout:   5 * time.Second,
	}

	result, err := cp.Execute(context.Background(), cmd)
	if result == nil {
		t.Fatal("Expected command result, got nil")
	}

	if result.Status != StatusFailed {
		t.Fatalf("Expected status %s, got %s", StatusFailed, result.Status)
	}

	if err == nil {
		t.Fatal("Expected Execute to return error when handler not found")
	}
}

func TestCommandProcessorSetTimeout(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	auditLog := &MockAuditLog{}
	cp := NewCommandProcessor(ctx, auditLog, 1, newTestLogger())
	defer cp.Close()

	// Set custom timeout
	cp.SetCommandTimeout(CommandStopStrategy, 100*time.Millisecond)

	cp.RegisterHandler(CommandStopStrategy, func(ctx context.Context, cmd *Command) (interface{}, error) {
		// Block forever
		<-ctx.Done()
		return nil, ctx.Err()
	})

	cmd := &Command{
		Type:      CommandStopStrategy,
		Timestamp: time.Now(),
		Payload:   StopStrategyPayload{},
	}

	result, _ := cp.Execute(context.Background(), cmd)
	if result == nil {
		t.Fatal("Expected command result, got nil")
	}

	// Should timeout due to custom timeout setting
	if result.Status != StatusTimedOut && result.Status != StatusFailed {
		t.Fatalf("Expected status TimedOut or Failed, got %s", result.Status)
	}
}
