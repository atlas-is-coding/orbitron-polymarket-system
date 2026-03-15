package data

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/atlasdev/orbitron/internal/api"
)

// Client — Data API клиент (https://data-api.polymarket.com).
// Публичный API, аутентификация не требуется.
type Client struct {
	http *api.Client
}

// NewClient создаёт Data API клиент.
func NewClient(httpClient *api.Client) *Client {
	return &Client{http: httpClient}
}

// GetPositions возвращает открытые позиции пользователя по адресу кошелька.
func (c *Client) GetPositions(params PositionsParams) ([]Position, error) {
	path := "/positions" + buildPositionsQuery(params)
	resp, err := c.http.Get(path, nil)
	if err != nil {
		return nil, fmt.Errorf("data: GetPositions: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("data: GetPositions HTTP %d: %s", resp.StatusCode, resp.Body)
	}
	var positions []Position
	if err := json.Unmarshal(resp.Body, &positions); err != nil {
		return nil, fmt.Errorf("data: GetPositions: decode: %w", err)
	}
	return positions, nil
}

// GetClosedPositions возвращает закрытые позиции пользователя.
func (c *Client) GetClosedPositions(user string, limit, offset int) ([]ClosedPosition, error) {
	path := fmt.Sprintf("/closed-positions?user=%s", user)
	if limit > 0 {
		path += "&limit=" + strconv.Itoa(limit)
	}
	if offset > 0 {
		path += "&offset=" + strconv.Itoa(offset)
	}
	resp, err := c.http.Get(path, nil)
	if err != nil {
		return nil, fmt.Errorf("data: GetClosedPositions: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("data: GetClosedPositions HTTP %d: %s", resp.StatusCode, resp.Body)
	}
	var positions []ClosedPosition
	if err := json.Unmarshal(resp.Body, &positions); err != nil {
		return nil, fmt.Errorf("data: GetClosedPositions: decode: %w", err)
	}
	return positions, nil
}

// GetTrades возвращает историю сделок пользователя из Data API.
func (c *Client) GetTrades(params TradesParams) ([]Trade, error) {
	path := "/trades" + buildTradesQuery(params)
	resp, err := c.http.Get(path, nil)
	if err != nil {
		return nil, fmt.Errorf("data: GetTrades: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("data: GetTrades HTTP %d: %s", resp.StatusCode, resp.Body)
	}
	var trades []Trade
	if err := json.Unmarshal(resp.Body, &trades); err != nil {
		return nil, fmt.Errorf("data: GetTrades: decode: %w", err)
	}
	return trades, nil
}

func buildPositionsQuery(p PositionsParams) string {
	q := "?"
	if p.User != "" {
		q += "user=" + p.User + "&"
	}
	if p.SortBy != "" {
		q += "sortBy=" + p.SortBy + "&"
	}
	if p.SortOrder != "" {
		q += "sortDirection=" + p.SortOrder + "&"
	}
	if p.SizeThreshold > 0 {
		q += "sizeThreshold=" + strconv.FormatFloat(p.SizeThreshold, 'f', -1, 64) + "&"
	}
	if p.Limit > 0 {
		q += "limit=" + strconv.Itoa(p.Limit) + "&"
	}
	if p.Offset > 0 {
		q += "offset=" + strconv.Itoa(p.Offset) + "&"
	}
	if q == "?" {
		return ""
	}
	return q[:len(q)-1]
}

func buildTradesQuery(p TradesParams) string {
	q := "?"
	if p.User != "" {
		q += "user=" + p.User + "&"
	}
	if p.Market != "" {
		q += "market=" + p.Market + "&"
	}
	if p.AssetID != "" {
		q += "asset_id=" + p.AssetID + "&"
	}
	if p.Limit > 0 {
		q += "limit=" + strconv.Itoa(p.Limit) + "&"
	}
	if p.Offset > 0 {
		q += "offset=" + strconv.Itoa(p.Offset) + "&"
	}
	if q == "?" {
		return ""
	}
	return q[:len(q)-1]
}
