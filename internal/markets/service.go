package markets

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/atlasdev/orbitron/internal/api/gamma"
	"github.com/atlasdev/orbitron/internal/tui"
)

// Service polls Gamma API for markets and manages price alerts.
type Service struct {
	gamma *gamma.Client
	bus   *tui.EventBus
	log   *zerolog.Logger

	mu      sync.RWMutex
	markets []gamma.Market
	tags    []gamma.Tag
	alerts  map[string]*AlertRule
}

// NewService creates a Service. gammaClient and bus may be nil (for tests).
func NewService(gammaClient *gamma.Client, bus *tui.EventBus) *Service {
	return &Service{
		gamma:  gammaClient,
		bus:    bus,
		alerts: make(map[string]*AlertRule),
	}
}

// WithLogger attaches a logger.
func (s *Service) WithLogger(log *zerolog.Logger) *Service {
	s.log = log
	return s
}

// Run starts polling every 30 seconds. Blocks until ctx is cancelled.
func (s *Service) Run(ctx context.Context) error {
	if err := s.poll(); err != nil && s.log != nil {
		s.log.Warn().Err(err).Msg("markets: initial poll failed")
	}
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := s.poll(); err != nil && s.log != nil {
				s.log.Warn().Err(err).Msg("markets: poll failed")
			}
		}
	}
}

func (s *Service) poll() error {
	if s.gamma == nil {
		return nil
	}
	t := true
	events, err := s.gamma.GetEvents(gamma.EventsParams{Active: &t, Limit: 200})
	if err != nil {
		return fmt.Errorf("GetEvents: %w", err)
	}

	var markets []gamma.Market
	tagSet := map[string]gamma.Tag{}
	for _, ev := range events {
		for _, m := range ev.Markets {
			markets = append(markets, m)
			for _, tg := range m.Tags {
				tagSet[tg.Slug] = tg
			}
		}
	}

	tags := make([]gamma.Tag, 0, len(tagSet))
	for _, tg := range tagSet {
		tags = append(tags, tg)
	}

	s.mu.Lock()
	s.markets = markets
	s.tags = tags
	s.mu.Unlock()

	s.checkAlerts(markets)

	if s.bus != nil {
		s.bus.Send(tui.MarketsUpdatedMsg{Markets: markets, Tags: tags})
	}
	return nil
}

func (s *Service) checkAlerts(markets []gamma.Market) {
	// Build price map outside any lock (operates on passed slice, not s.markets)
	priceMap := map[string]float64{}
	for _, m := range markets {
		if len(m.OutcomePrices) > 0 {
			if p, err := strconv.ParseFloat(string(m.OutcomePrices[0]), 64); err == nil {
				priceMap[m.ConditionID] = p
			}
		}
	}

	// Collect fired alerts under lock, send outside lock
	s.mu.Lock()
	var fired []tui.MarketAlertMsg
	for _, a := range s.alerts {
		if a.Triggered {
			continue
		}
		price, ok := priceMap[a.ConditionID]
		if !ok {
			continue
		}
		f := (a.Direction == "above" && price >= a.Threshold) ||
			(a.Direction == "below" && price <= a.Threshold)
		if f {
			a.Triggered = true
			fired = append(fired, tui.MarketAlertMsg{
				ConditionID:  a.ConditionID,
				Threshold:    a.Threshold,
				Direction:    a.Direction,
				CurrentPrice: price,
			})
		}
	}
	s.mu.Unlock()

	if s.bus != nil {
		for _, msg := range fired {
			s.bus.Send(msg)
		}
	}
}

// GetByTag returns markets matching the given tag slug. Empty slug returns all.
func (s *Service) GetByTag(slug string) []gamma.Market {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if slug == "" {
		out := make([]gamma.Market, len(s.markets))
		copy(out, s.markets)
		return out
	}
	var out []gamma.Market
	for _, m := range s.markets {
		for _, tg := range m.Tags {
			if tg.Slug == slug {
				out = append(out, m)
				break
			}
		}
	}
	return out
}

// GetMarket returns a single market by conditionID.
func (s *Service) GetMarket(conditionID string) (gamma.Market, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, m := range s.markets {
		if m.ConditionID == conditionID {
			return m, true
		}
	}
	return gamma.Market{}, false
}

// Tags returns the current tag list.
func (s *Service) Tags() []gamma.Tag {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]gamma.Tag, len(s.tags))
	copy(out, s.tags)
	return out
}

// AddAlert adds an alert rule and returns its generated ID.
func (s *Service) AddAlert(rule AlertRule) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	rule.ID = fmt.Sprintf("alert-%d", time.Now().UnixNano())
	rule.CreatedAt = time.Now()
	s.alerts[rule.ID] = &rule
	return rule.ID
}

// RemoveAlert removes an alert by ID.
func (s *Service) RemoveAlert(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.alerts, id)
}

// Alerts returns a snapshot of all alert rules.
func (s *Service) Alerts() []AlertRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]AlertRule, 0, len(s.alerts))
	for _, a := range s.alerts {
		out = append(out, *a)
	}
	return out
}

// SetMarketsForTest injects markets directly. The tb argument ensures this method
// can only be called from test code.
func (s *Service) SetMarketsForTest(_ testing.TB, markets []gamma.Market) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.markets = markets
}
