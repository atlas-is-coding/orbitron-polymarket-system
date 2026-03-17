package markets

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/atlasdev/orbitron/internal/api/gamma"
	"github.com/atlasdev/orbitron/internal/storage"
	"github.com/atlasdev/orbitron/internal/tui"
)

// GammaClient defines the methods we use from gamma.Client.
type GammaClient interface {
	GetMarkets(params gamma.MarketsParams) ([]gamma.Market, error)
	GetMarket(id string) (*gamma.Market, error)
}

// Service polls Gamma API for markets and manages price alerts.
type Service struct {
	gamma  GammaClient
	bus    *tui.EventBus
	cache  storage.MarketCacheStore // nil = no persistence
	log    *zerolog.Logger

	mu      sync.RWMutex
	markets []gamma.Market
	tags    []gamma.Tag
	total   int // total cached market count
	alerts  map[string]*AlertRule

	syncMu sync.Mutex // guards against concurrent full-sync runs
}

// NewService creates a Service. Any argument may be nil (for tests or optional features).
func NewService(gammaClient GammaClient, bus *tui.EventBus, cache storage.MarketCacheStore) *Service {
	return &Service{
		gamma:  gammaClient,
		bus:    bus,
		cache:  cache,
		alerts: make(map[string]*AlertRule),
	}
}

// WithLogger attaches a logger.
func (s *Service) WithLogger(log *zerolog.Logger) *Service {
	s.log = log
	return s
}

// Run starts the markets service. Blocks until ctx is cancelled.
func (s *Service) Run(ctx context.Context) error {
	// 1. Load from cache immediately (may be empty on first run)
	if s.cache != nil {
		if err := s.loadFromCache(ctx); err != nil && s.log != nil {
			s.log.Warn().Err(err).Msg("markets: cache load failed, continuing without cache")
		}
	}

	// 2. Initial load: top-500 markets (5 pages x 100)
	s.initialLoad(ctx)

	// 3. Background full sync (pages 500+)
	go s.runFullSync(ctx, 500, false)

	// 4. Main polling loop
	trendTicker := time.NewTicker(30 * time.Second)
	syncTicker := time.NewTicker(30 * time.Minute)
	defer trendTicker.Stop()
	defer syncTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-trendTicker.C:
			if err := s.pollTrending(ctx); err != nil && s.log != nil {
				s.log.Warn().Err(err).Msg("markets: trending poll failed")
			}
		case <-syncTicker.C:
			go s.runFullSync(ctx, 0, true)
		}
	}
}

// loadFromCache populates in-memory markets from SQLite.
func (s *Service) loadFromCache(ctx context.Context) error {
	records, err := s.cache.GetCachedMarkets(ctx)
	if err != nil {
		return err
	}
	markets := make([]gamma.Market, 0, len(records))
	tagSet := map[string]gamma.Tag{}
	for _, r := range records {
		var m gamma.Market
		if err := json.Unmarshal([]byte(r.Data), &m); err != nil {
			continue
		}
		markets = append(markets, m)
		for _, tg := range m.Tags {
			tagSet[tg.Slug] = tg
		}
	}
	tags := make([]gamma.Tag, 0, len(tagSet))
	for _, tg := range tagSet {
		tags = append(tags, tg)
	}
	s.mu.Lock()
	s.markets = markets
	s.tags = tags
	s.total = len(markets)
	s.mu.Unlock()
	s.notifyBus()
	return nil
}

// initialLoad fetches the top-500 markets (5 pages x 100) and emits progress.
// Emits MarketsReadyMsg when done or on any unrecoverable error.
func (s *Service) initialLoad(ctx context.Context) {
	if s.gamma == nil {
		s.emitReady()
		return
	}
	const pages = 5
	const pageSize = 100
	t := true
	f := false

	for page := 0; page < pages; page++ {
		select {
		case <-ctx.Done():
			s.emitReady()
			return
		default:
		}
		params := gamma.MarketsParams{
			Active:    &t,
			Closed:    &f,
			Order:     "volume_24hr",
			Ascending: false,
			Limit:     pageSize,
			Offset:    page * pageSize,
		}
		mks, err := s.gamma.GetMarkets(params)
		if err != nil {
			if s.log != nil {
				s.log.Warn().Err(err).Int("page", page).Msg("markets: initial load page failed")
			}
			s.emitReady()
			return
		}
		s.mergePage(ctx, mks)
		loaded := (page + 1) * pageSize
		if loaded > pages*pageSize {
			loaded = pages * pageSize
		}
		if s.bus != nil {
			s.bus.Send(tui.MarketsLoadingMsg{Loaded: loaded, Total: pages * pageSize})
		}
		if len(mks) < pageSize {
			break // reached end of results
		}
	}
	s.emitReady()
}

// pollTrending refreshes the trending top-100 in memory and cache.
func (s *Service) pollTrending(ctx context.Context) error {
	if s.gamma == nil {
		return nil
	}
	t := true
	f := false
	mks, err := s.gamma.GetMarkets(gamma.MarketsParams{
		Active: &t, Closed: &f,
		Order: "volume_24hr", Ascending: false,
		Limit: 100,
	})
	if err != nil {
		return fmt.Errorf("pollTrending: %w", err)
	}
	s.mergePage(ctx, mks)
	return nil
}

// runFullSync paginates all markets from startOffset, updating cache.
// When detectNew=true, records markets whose first_seen is after syncStart.
func (s *Service) runFullSync(ctx context.Context, startOffset int, detectNew bool) {
	if s.gamma == nil {
		return
	}
	if !s.syncMu.TryLock() {
		return // another sync is already running
	}
	defer s.syncMu.Unlock()

	syncStart := time.Now()
	const pageSize = 100
	t := true
	f := false
	offset := startOffset

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		mks, err := s.gamma.GetMarkets(gamma.MarketsParams{
			Active: &t, Closed: &f,
			Order: "volume_24hr", Ascending: false,
			Limit: pageSize, Offset: offset,
		})
		if err != nil {
			if s.log != nil {
				s.log.Warn().Err(err).Int("offset", offset).Msg("markets: full sync page failed")
			}
			return
		}
		if len(mks) == 0 {
			break
		}
		s.mergePage(ctx, mks)
		offset += len(mks)
		if len(mks) < pageSize {
			break
		}
	}

	if detectNew && s.cache != nil && s.log != nil {
		newMkts, err := s.cache.GetNewMarkets(ctx, syncStart)
		if err == nil && len(newMkts) > 0 {
			s.log.Info().Int("count", len(newMkts)).Msg("markets: new markets detected")
		}
	}
}

// mergePage adds/updates markets in memory and cache.
func (s *Service) mergePage(ctx context.Context, mks []gamma.Market) {
	s.mu.Lock()
	existing := make(map[string]int, len(s.markets))
	for i, m := range s.markets {
		existing[m.ConditionID] = i
	}
	tagSet := map[string]gamma.Tag{}
	for _, tg := range s.tags {
		tagSet[tg.Slug] = tg
	}
	for _, m := range mks {
		if idx, ok := existing[m.ConditionID]; ok {
			s.markets[idx] = m
		} else {
			s.markets = append(s.markets, m)
			existing[m.ConditionID] = len(s.markets) - 1
		}
		for _, tg := range m.Tags {
			tagSet[tg.Slug] = tg
		}
	}
	tags := make([]gamma.Tag, 0, len(tagSet))
	for _, tg := range tagSet {
		tags = append(tags, tg)
	}
	s.tags = tags
	s.total = len(s.markets)
	s.mu.Unlock()

	s.notifyBus()

	if s.cache != nil {
		now := time.Now()
		records := make([]storage.MarketCacheRecord, 0, len(mks))
		for _, m := range mks {
			data, err := json.Marshal(m)
			if err != nil {
				continue
			}
			records = append(records, storage.MarketCacheRecord{
				ConditionID: m.ConditionID,
				Data:        string(data),
				UpdatedAt:   now,
				FirstSeen:   now,
			})
		}
		if err := s.cache.UpsertMarkets(ctx, records); err != nil && s.log != nil {
			s.log.Warn().Err(err).Msg("markets: cache upsert failed")
		}
	}

	s.checkAlerts(mks)
}

// notifyBus sends a MarketsUpdatedMsg to the event bus (if configured).
func (s *Service) notifyBus() {
	if s.bus == nil {
		return
	}
	s.mu.RLock()
	markets := make([]gamma.Market, len(s.markets))
	copy(markets, s.markets)
	tags := make([]gamma.Tag, len(s.tags))
	copy(tags, s.tags)
	s.mu.RUnlock()
	s.bus.Send(tui.MarketsUpdatedMsg{Markets: markets, Tags: tags})
}

// emitReady sends MarketsReadyMsg to the event bus.
func (s *Service) emitReady() {
	if s.bus != nil {
		s.bus.Send(tui.MarketsReadyMsg{})
	}
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

// GetTrending returns markets sorted by volume descending, capped at limit.
// limit <= 0 means return all.
func (s *Service) GetTrending(limit int) []gamma.Market {
	s.mu.RLock()
	cp := make([]gamma.Market, len(s.markets))
	copy(cp, s.markets)
	s.mu.RUnlock()

	sort.Slice(cp, func(i, j int) bool {
		return cp[i].Volume > cp[j].Volume
	})
	if limit > 0 && len(cp) > limit {
		return cp[:limit]
	}
	return cp
}

// TotalCount returns the total number of cached markets.
func (s *Service) TotalCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.total
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
