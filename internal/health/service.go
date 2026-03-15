package health

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"

	"github.com/atlasdev/orbitron/internal/api"
)

const (
	pollInterval  = 60 * time.Second
	checkTimeout  = 10 * time.Second
	degradedCLOB  = 500 * time.Millisecond
	degradedGamma = time.Second
	degradedData  = time.Second
	degradedWS    = 2 * time.Second
)

// Publisher receives HealthSnapshot values from the service.
// Implemented by main.go's healthPublisher adapter that wraps tui.EventBus.
type Publisher interface {
	Send(snap HealthSnapshot)
}

// Endpoints holds the base URLs to check.
type Endpoints struct {
	ClobURL  string
	GammaURL string
	DataURL  string
	WSURL    string
}

// Service polls Polymarket API health every 60s and publishes snapshots.
type Service struct {
	endpoints Endpoints
	dial      api.DialFunc
	pub       Publisher
	log       zerolog.Logger

	mu       sync.RWMutex
	snapshot HealthSnapshot
}

// New creates a health Service.
func New(ep Endpoints, dial api.DialFunc, pub Publisher, log zerolog.Logger) *Service {
	return &Service{
		endpoints: ep,
		dial:      dial,
		pub:       pub,
		log:       log.With().Str("component", "health").Logger(),
	}
}

// Snapshot returns the latest cached health snapshot (thread-safe).
func (s *Service) Snapshot() HealthSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.snapshot
}

// Start runs the health poller until ctx is cancelled.
// Performs first check immediately (no initial delay).
func (s *Service) Start(ctx context.Context) {
	s.check()
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.check()
		}
	}
}

func (s *Service) check() {
	type result struct {
		svc ServiceHealth
		geo *GeoStatus
	}
	ch := make(chan result, 5)

	checks := []struct {
		fn func() (ServiceHealth, *GeoStatus)
	}{
		{func() (ServiceHealth, *GeoStatus) {
			return s.httpCheck("CLOB", s.endpoints.ClobURL+"/time", degradedCLOB), nil
		}},
		{func() (ServiceHealth, *GeoStatus) {
			return s.httpCheck("Gamma", s.endpoints.GammaURL+"/markets?limit=1", degradedGamma), nil
		}},
		{func() (ServiceHealth, *GeoStatus) {
			return s.httpCheck("Data", s.endpoints.DataURL+"/markets?limit=1", degradedData), nil
		}},
		{func() (ServiceHealth, *GeoStatus) {
			return s.wsCheck(), nil
		}},
		{func() (ServiceHealth, *GeoStatus) {
			start := time.Now()
			geo, err := CheckGeoblock(s.dial)
			lat := time.Since(start).Milliseconds()
			if err != nil {
				return ServiceHealth{Name: "Geoblock", Status: StatusDown, LatencyMs: lat, Error: err.Error()}, nil
			}
			return ServiceHealth{Name: "Geoblock", Status: StatusOK, LatencyMs: lat}, geo
		}},
	}

	for _, c := range checks {
		c := c
		go func() {
			svc, geo := c.fn()
			ch <- result{svc, geo}
		}()
	}

	snap := HealthSnapshot{UpdatedAt: time.Now()}
	for range checks {
		r := <-ch
		snap.Services = append(snap.Services, r.svc)
		if r.geo != nil {
			snap.Geo = r.geo
		}
	}

	s.mu.Lock()
	s.snapshot = snap
	s.mu.Unlock()

	if s.pub != nil {
		s.pub.Send(snap)
	}
	s.log.Debug().Int("services", len(snap.Services)).Msg("health check complete")
}

func (s *Service) httpCheck(name, url string, degradedThreshold time.Duration) ServiceHealth {
	client := &http.Client{Timeout: checkTimeout}
	start := time.Now()
	resp, err := client.Get(url) //nolint:noctx
	latency := time.Since(start)
	if err != nil {
		return ServiceHealth{Name: name, Status: StatusDown, LatencyMs: latency.Milliseconds(), Error: err.Error()}
	}
	resp.Body.Close()
	status := StatusOK
	if latency > degradedThreshold {
		status = StatusDegraded
	}
	return ServiceHealth{Name: name, Status: status, LatencyMs: latency.Milliseconds()}
}

func (s *Service) wsCheck() ServiceHealth {
	wsURL := strings.TrimRight(s.endpoints.WSURL, "/") + "/market"
	start := time.Now()
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	latency := time.Since(start)
	if err != nil {
		return ServiceHealth{Name: "WebSocket", Status: StatusDown, LatencyMs: latency.Milliseconds(), Error: err.Error()}
	}
	conn.Close()
	status := StatusOK
	if latency > degradedWS {
		status = StatusDegraded
	}
	return ServiceHealth{Name: "WebSocket", Status: status, LatencyMs: latency.Milliseconds()}
}
