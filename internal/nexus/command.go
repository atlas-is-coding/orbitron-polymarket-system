package nexus

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// CommandHandler is a function that executes a command and returns a result or error.
type CommandHandler func(ctx context.Context, cmd *Command) (interface{}, error)

// AuditLog interface defines audit logging methods for commands and events.
type AuditLog interface {
	SaveCommand(ctx context.Context, cmd *Command) error
	SaveEvent(ctx context.Context, event *Event) error
	GetCommandHistory(ctx context.Context, limit int) ([]*Command, error)
	GetEventHistory(ctx context.Context, limit int) ([]*Event, error)
}

// CommandProcessor manages synchronous and asynchronous command execution
// with timeout handling, status tracking, and audit logging.
type CommandProcessor struct {
	mu              sync.RWMutex
	inFlight        map[string]*Command                  // tracking active commands
	handlers        map[CommandType]CommandHandler
	auditLog        AuditLog
	asyncQueue      chan Command
	asyncWorkers    int
	defaultTimeout  time.Duration
	commandTimeout  map[CommandType]time.Duration
	log             zerolog.Logger
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	processingCount atomic.Int32
	timeoutMutex    sync.RWMutex
}

// NewCommandProcessor creates a new CommandProcessor with N async worker goroutines.
// asyncWorkers specifies the number of concurrent async command handlers.
// If asyncWorkers <= 0, defaults to 4.
func NewCommandProcessor(ctx context.Context, auditLog AuditLog, asyncWorkers int, log zerolog.Logger) *CommandProcessor {
	if asyncWorkers <= 0 {
		asyncWorkers = 4
	}

	procCtx, cancel := context.WithCancel(ctx)

	cp := &CommandProcessor{
		inFlight:       make(map[string]*Command),
		handlers:       make(map[CommandType]CommandHandler),
		auditLog:       auditLog,
		asyncQueue:     make(chan Command, 1000),
		asyncWorkers:   asyncWorkers,
		defaultTimeout: 30 * time.Second,
		commandTimeout: make(map[CommandType]time.Duration),
		log:            log,
		ctx:            procCtx,
		cancel:         cancel,
	}

	// Start async worker goroutines
	for i := 0; i < asyncWorkers; i++ {
		cp.wg.Add(1)
		go cp.asyncWorker(i)
	}

	return cp
}

// RegisterHandler registers a command handler for the given command type.
func (cp *CommandProcessor) RegisterHandler(cmdType CommandType, handler CommandHandler) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.handlers[cmdType] = handler
}

// SetCommandTimeout sets a custom timeout for a specific command type.
// If the command's timeout is 0, this timeout will be used instead of the default.
func (cp *CommandProcessor) SetCommandTimeout(cmdType CommandType, timeout time.Duration) {
	cp.timeoutMutex.Lock()
	defer cp.timeoutMutex.Unlock()
	cp.commandTimeout[cmdType] = timeout
}

// getTimeoutForCommand returns the timeout to use for a command.
// Priority: command.Timeout > commandType custom timeout > default timeout
func (cp *CommandProcessor) getTimeoutForCommand(cmd *Command) time.Duration {
	if cmd.Timeout > 0 {
		return cmd.Timeout
	}

	cp.timeoutMutex.RLock()
	defer cp.timeoutMutex.RUnlock()

	if timeout, ok := cp.commandTimeout[cmd.Type]; ok {
		return timeout
	}

	return cp.defaultTimeout
}

// Execute synchronously executes a command and blocks until completion.
// Updates command status from Pending → Processing → Completed/Failed/TimedOut.
// Returns the updated command and any error from the handler.
func (cp *CommandProcessor) Execute(ctx context.Context, cmd *Command) (*Command, error) {
	// Generate UUID if not set
	if cmd.ID == "" {
		cmd.ID = uuid.New().String()
	}

	// Set initial status and timestamp
	if cmd.Timestamp.IsZero() {
		cmd.Timestamp = time.Now()
	}
	cmd.Status = StatusPending

	// Get timeout for this command
	timeout := cp.getTimeoutForCommand(cmd)
	if cmd.DeadlineAt.IsZero() {
		cmd.DeadlineAt = time.Now().Add(timeout)
	}

	// Add to inFlight map
	cp.mu.Lock()
	cp.inFlight[cmd.ID] = cmd
	cp.mu.Unlock()

	defer func() {
		// Remove from inFlight after processing
		cp.mu.Lock()
		delete(cp.inFlight, cmd.ID)
		cp.mu.Unlock()
	}()

	// Look up handler
	cp.mu.RLock()
	handler, ok := cp.handlers[cmd.Type]
	cp.mu.RUnlock()

	if !ok {
		cmd.Status = StatusFailed
		cmd.Error = fmt.Sprintf("handler not found for command type: %s", cmd.Type)
		cp.auditLog.SaveCommand(ctx, cmd)
		return cmd, fmt.Errorf("handler not found for command type: %s", cmd.Type)
	}

	// Create context with deadline
	ctxWithDeadline, cancel := context.WithDeadline(ctx, cmd.DeadlineAt)
	defer cancel()

	// Update status to processing
	cmd.Status = StatusProcessing
	cp.mu.Lock()
	cp.inFlight[cmd.ID] = cmd
	cp.mu.Unlock()

	// Execute handler
	result, err := handler(ctxWithDeadline, cmd)

	// Update command based on result
	if err != nil {
		cmd.Error = err.Error()
		// Check if context deadline exceeded
		if ctxWithDeadline.Err() == context.DeadlineExceeded {
			cmd.Status = StatusTimedOut
		} else {
			cmd.Status = StatusFailed
		}
	} else {
		cmd.Status = StatusCompleted
		cmd.Result = result
	}

	// Save to audit log
	_ = cp.auditLog.SaveCommand(ctx, cmd)

	return cmd, err
}

// ExecuteAsync asynchronously executes a command and returns the command ID immediately.
// The command is queued for async processing; status can be checked with GetStatus.
// Returns error if the async queue is full or context is cancelled.
func (cp *CommandProcessor) ExecuteAsync(ctx context.Context, cmd *Command) (string, error) {
	// Generate UUID if not set
	if cmd.ID == "" {
		cmd.ID = uuid.New().String()
	}

	// Set initial status and timestamp
	if cmd.Timestamp.IsZero() {
		cmd.Timestamp = time.Now()
	}
	if cmd.Status == "" {
		cmd.Status = StatusPending
	}

	// Set deadline
	if cmd.Timeout > 0 && cmd.DeadlineAt.IsZero() {
		cmd.DeadlineAt = time.Now().Add(cmd.Timeout)
	} else if cmd.Timeout == 0 && cmd.DeadlineAt.IsZero() {
		timeout := cp.getTimeoutForCommand(cmd)
		cmd.DeadlineAt = time.Now().Add(timeout)
	}

	// Add to inFlight map
	cp.mu.Lock()
	cp.inFlight[cmd.ID] = cmd
	cp.mu.Unlock()

	cmdID := cmd.ID

	// Try to queue the command (non-blocking select)
	select {
	case cp.asyncQueue <- *cmd:
		return cmdID, nil
	case <-cp.ctx.Done():
		cp.mu.Lock()
		delete(cp.inFlight, cmdID)
		cp.mu.Unlock()
		return "", fmt.Errorf("processor context cancelled")
	default:
		cp.mu.Lock()
		delete(cp.inFlight, cmdID)
		cp.mu.Unlock()
		return "", fmt.Errorf("async queue full")
	}
}

// GetStatus returns the current status of a command by ID.
// Returns error if the command is not found or has already completed and been removed.
func (cp *CommandProcessor) GetStatus(cmdID string) (*Command, error) {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	cmd, ok := cp.inFlight[cmdID]
	if !ok {
		return nil, fmt.Errorf("command not found: %s", cmdID)
	}

	return cmd, nil
}

// asyncWorker processes commands from the async queue.
// Each worker runs in its own goroutine and handles timeout and execution.
func (cp *CommandProcessor) asyncWorker(workerID int) {
	defer cp.wg.Done()

	for {
		select {
		case <-cp.ctx.Done():
			cp.log.Debug().Int("worker_id", workerID).Msg("async worker shutting down")
			return
		case cmd := <-cp.asyncQueue:
			cp.processAsyncCommand(&cmd)
		}
	}
}

// processAsyncCommand executes a single async command.
func (cp *CommandProcessor) processAsyncCommand(cmd *Command) {
	cp.processingCount.Add(1)
	defer cp.processingCount.Add(-1)

	// Check if deadline already passed
	if time.Now().After(cmd.DeadlineAt) {
		cp.mu.Lock()
		if c, ok := cp.inFlight[cmd.ID]; ok {
			c.Status = StatusTimedOut
			c.Error = "deadline exceeded before execution"
		}
		cp.mu.Unlock()
		cp.auditLog.SaveCommand(cp.ctx, cmd)
		return
	}

	// Look up handler
	cp.mu.RLock()
	handler, ok := cp.handlers[cmd.Type]
	cp.mu.RUnlock()

	if !ok {
		cp.mu.Lock()
		if c, ok := cp.inFlight[cmd.ID]; ok {
			c.Status = StatusFailed
			c.Error = fmt.Sprintf("handler not found for command type: %s", cmd.Type)
		}
		cp.mu.Unlock()
		cp.auditLog.SaveCommand(cp.ctx, cmd)
		return
	}

	// Update status to processing
	cp.mu.Lock()
	if c, ok := cp.inFlight[cmd.ID]; ok {
		c.Status = StatusProcessing
	}
	cp.mu.Unlock()

	// Create context with deadline
	ctxWithDeadline, cancel := context.WithDeadline(cp.ctx, cmd.DeadlineAt)
	defer cancel()

	// Execute handler
	result, err := handler(ctxWithDeadline, cmd)

	// Update command based on result
	cp.mu.Lock()
	if c, ok := cp.inFlight[cmd.ID]; ok {
		if err != nil {
			c.Error = err.Error()
			if ctxWithDeadline.Err() == context.DeadlineExceeded {
				c.Status = StatusTimedOut
			} else {
				c.Status = StatusFailed
			}
		} else {
			c.Status = StatusCompleted
			c.Result = result
		}
	}
	cp.mu.Unlock()

	// Save to audit log
	cp.mu.RLock()
	cmdToLog := cp.inFlight[cmd.ID]
	cp.mu.RUnlock()
	if cmdToLog != nil {
		_ = cp.auditLog.SaveCommand(cp.ctx, cmdToLog)
	}
}

// Stats returns statistics about the command processor.
// Returns a map with keys: "in_flight", "queue_len", "workers", "processing"
func (cp *CommandProcessor) Stats() map[string]interface{} {
	cp.mu.RLock()
	inFlightCount := len(cp.inFlight)
	cp.mu.RUnlock()

	return map[string]interface{}{
		"in_flight":  inFlightCount,
		"queue_len":  len(cp.asyncQueue),
		"workers":    cp.asyncWorkers,
		"processing": cp.processingCount.Load(),
	}
}

// Close gracefully shuts down the command processor.
// Closes the async queue, cancels the context, and waits for workers to finish.
func (cp *CommandProcessor) Close() error {
	cp.log.Debug().Msg("closing command processor")

	// Signal context cancellation
	cp.cancel()

	// Close async queue
	close(cp.asyncQueue)

	// Wait for all workers to finish
	cp.wg.Wait()

	cp.log.Debug().Msg("command processor closed")
	return nil
}
