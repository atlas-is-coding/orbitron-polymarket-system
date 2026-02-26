package telegrambot

import (
	"sync"

	"github.com/atlasdev/polytrade-bot/internal/tui"
)

// SubsystemStatus holds name + active state.
type SubsystemStatus struct {
	Name   string
	Active bool
}

// BotState is a thread-safe cache of the latest bot data,
// updated by the EventBus consumer goroutine.
type BotState struct {
	mu         sync.RWMutex
	balance    float64
	orders     []tui.OrderRow
	positions  []tui.PositionRow
	traders    []tui.TraderRow
	logs       []string
	subsystems map[string]bool
}

// NewBotState creates an empty BotState.
func NewBotState() *BotState {
	return &BotState{subsystems: make(map[string]bool)}
}

func (s *BotState) SetBalance(v float64) {
	s.mu.Lock()
	s.balance = v
	s.mu.Unlock()
}

func (s *BotState) Balance() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.balance
}

func (s *BotState) SetOrders(rows []tui.OrderRow) {
	s.mu.Lock()
	s.orders = rows
	s.mu.Unlock()
}

func (s *BotState) Orders() []tui.OrderRow {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp := make([]tui.OrderRow, len(s.orders))
	copy(cp, s.orders)
	return cp
}

func (s *BotState) SetPositions(rows []tui.PositionRow) {
	s.mu.Lock()
	s.positions = rows
	s.mu.Unlock()
}

func (s *BotState) Positions() []tui.PositionRow {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp := make([]tui.PositionRow, len(s.positions))
	copy(cp, s.positions)
	return cp
}

func (s *BotState) SetTraders(rows []tui.TraderRow) {
	s.mu.Lock()
	s.traders = rows
	s.mu.Unlock()
}

func (s *BotState) Traders() []tui.TraderRow {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp := make([]tui.TraderRow, len(s.traders))
	copy(cp, s.traders)
	return cp
}

// AddLog appends a log line and caps the buffer at 50 lines.
func (s *BotState) AddLog(line string) {
	s.mu.Lock()
	s.logs = append(s.logs, line)
	if len(s.logs) > 50 {
		s.logs = s.logs[len(s.logs)-50:]
	}
	s.mu.Unlock()
}

func (s *BotState) Logs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp := make([]string, len(s.logs))
	copy(cp, s.logs)
	return cp
}

func (s *BotState) SetSubsystem(name string, active bool) {
	s.mu.Lock()
	s.subsystems[name] = active
	s.mu.Unlock()
}

func (s *BotState) Subsystems() []SubsystemStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]SubsystemStatus, 0, len(s.subsystems))
	for name, active := range s.subsystems {
		result = append(result, SubsystemStatus{Name: name, Active: active})
	}
	return result
}
