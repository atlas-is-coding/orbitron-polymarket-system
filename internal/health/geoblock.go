package health

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/atlasdev/orbitron/internal/api"
)

const geoblockURL = "https://polymarket.com/api/geoblock"

// CheckGeoblock queries the Polymarket geoblock endpoint.
// dial may be nil for direct connection (no proxy).
func CheckGeoblock(dial api.DialFunc) (*GeoStatus, error) {
	httpClient := buildHTTPClient(dial)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, geoblockURL, nil)
	if err != nil {
		return nil, fmt.Errorf("geoblock: build request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("geoblock: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("geoblock: read response: %w", err)
	}

	var gs GeoStatus
	if err := json.Unmarshal(body, &gs); err != nil {
		return nil, fmt.Errorf("geoblock: parse response (status %d): %w", resp.StatusCode, err)
	}
	return &gs, nil
}

// buildHTTPClient creates an http.Client that routes through the given dialer.
// When dial is nil a default client with 10s timeout is used.
func buildHTTPClient(dial api.DialFunc) *http.Client {
	if dial == nil {
		return &http.Client{Timeout: 10 * time.Second}
	}
	transport := &http.Transport{
		DialContext: func(_ context.Context, network, addr string) (net.Conn, error) {
			return dial(addr)
		},
	}
	return &http.Client{Transport: transport, Timeout: 10 * time.Second}
}
