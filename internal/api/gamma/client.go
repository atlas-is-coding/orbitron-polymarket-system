package gamma

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/atlasdev/orbitron/internal/api"
)

// Client — Gamma API клиент (только публичные эндпоинты, аутентификация не нужна).
type Client struct {
	http *api.Client
}

// NewClient создаёт Gamma API клиент.
func NewClient(httpClient *api.Client) *Client {
	return &Client{http: httpClient}
}

// GetMarkets возвращает рынки с фильтрацией.
func (c *Client) GetMarkets(params MarketsParams) ([]Market, error) {
	path := "/markets" + buildMarketsQuery(params)
	resp, err := c.http.Get(path, nil)
	if err != nil {
		return nil, fmt.Errorf("gamma: GetMarkets: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("gamma: GetMarkets HTTP %d: %s", resp.StatusCode, resp.Body)
	}
	var markets []Market
	if err := json.Unmarshal(resp.Body, &markets); err != nil {
		return nil, fmt.Errorf("gamma: GetMarkets: decode: %w", err)
	}
	return markets, nil
}

// GetMarket возвращает рынок по condition_id или slug.
func (c *Client) GetMarket(id string) (*Market, error) {
	resp, err := c.http.Get("/markets/"+id, nil)
	if err != nil {
		return nil, fmt.Errorf("gamma: GetMarket: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("gamma: GetMarket HTTP %d: %s", resp.StatusCode, resp.Body)
	}
	var market Market
	if err := json.Unmarshal(resp.Body, &market); err != nil {
		return nil, fmt.Errorf("gamma: GetMarket: decode: %w", err)
	}
	return &market, nil
}

// GetEvents возвращает события с фильтрацией.
func (c *Client) GetEvents(params EventsParams) ([]Event, error) {
	path := "/events" + buildEventsQuery(params)
	resp, err := c.http.Get(path, nil)
	if err != nil {
		return nil, fmt.Errorf("gamma: GetEvents: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("gamma: GetEvents HTTP %d: %s", resp.StatusCode, resp.Body)
	}
	var events []Event
	if err := json.Unmarshal(resp.Body, &events); err != nil {
		return nil, fmt.Errorf("gamma: GetEvents: decode: %w", err)
	}
	return events, nil
}

// GetEvent возвращает событие по ID.
func (c *Client) GetEvent(id string) (*Event, error) {
	resp, err := c.http.Get("/events/"+id, nil)
	if err != nil {
		return nil, fmt.Errorf("gamma: GetEvent: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("gamma: GetEvent HTTP %d: %s", resp.StatusCode, resp.Body)
	}
	var event Event
	if err := json.Unmarshal(resp.Body, &event); err != nil {
		return nil, fmt.Errorf("gamma: GetEvent: decode: %w", err)
	}
	return &event, nil
}

func buildMarketsQuery(p MarketsParams) string {
	q := "?"
	if p.Active != nil {
		q += "active=" + strconv.FormatBool(*p.Active) + "&"
	}
	if p.Category != "" {
		q += "category=" + p.Category + "&"
	}
	if p.Limit > 0 {
		q += "limit=" + strconv.Itoa(p.Limit) + "&"
	}
	if p.Offset > 0 {
		q += "offset=" + strconv.Itoa(p.Offset) + "&"
	}
	if p.Closed != nil {
		q += "closed=" + strconv.FormatBool(*p.Closed) + "&"
	}
	if p.Order != "" {
		q += "order=" + p.Order + "&"
		q += "ascending=" + strconv.FormatBool(p.Ascending) + "&"
	}
	if q == "?" {
		return ""
	}
	return q[:len(q)-1]
}

func buildEventsQuery(p EventsParams) string {
	q := "?"
	if p.Active != nil {
		q += "active=" + strconv.FormatBool(*p.Active) + "&"
	}
	if p.Category != "" {
		q += "category=" + p.Category + "&"
	}
	if p.Limit > 0 {
		q += "limit=" + strconv.Itoa(p.Limit) + "&"
	}
	if p.Offset > 0 {
		q += "offset=" + strconv.Itoa(p.Offset) + "&"
	}
	if p.Order != "" {
		q += "order=" + p.Order + "&"
		q += "ascending=" + strconv.FormatBool(p.Ascending) + "&"
	}
	if q == "?" {
		return ""
	}
	return q[:len(q)-1]
}
