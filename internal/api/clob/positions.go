package clob

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// GetPositions возвращает текущие открытые позиции пользователя.
// Требует L2 авторизации (GET /positions).
func (c *Client) GetPositions() ([]Position, error) {
	resp, err := c.privateGet("/positions")
	if err != nil {
		return nil, fmt.Errorf("clob: GetPositions: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("clob: GetPositions HTTP %d: %s", resp.StatusCode, resp.Body)
	}
	var positions []Position
	if err := json.Unmarshal(resp.Body, &positions); err != nil {
		return nil, fmt.Errorf("clob: GetPositions: decode: %w", err)
	}
	return positions, nil
}

// GetBalanceAllowance возвращает баланс и разрешения для токена.
// assetType: "COLLATERAL" для USDC или "CONDITIONAL" для YES/NO токена.
// tokenID нужен только для "CONDITIONAL".
func (c *Client) GetBalanceAllowance(assetType, tokenID string) (*BalanceAllowance, error) {
	path := "/balance-allowance?asset_type=" + assetType
	if tokenID != "" {
		path += "&token_id=" + tokenID
	}
	resp, err := c.privateGet(path)
	if err != nil {
		return nil, fmt.Errorf("clob: GetBalanceAllowance: %w", err)
	}
	return decode[BalanceAllowance](resp)
}

// GetDataTrades возвращает историю сделок с поддержкой пагинации (GET /data/trades).
// Требует L2 авторизации. Поддерживает cursor-based пагинацию.
func (c *Client) GetDataTrades(filter TradesFilter) (*TradesResponse, error) {
	path := "/data/trades" + buildTradesQuery(filter)
	resp, err := c.privateGet(path)
	if err != nil {
		return nil, fmt.Errorf("clob: GetDataTrades: %w", err)
	}
	return decode[TradesResponse](resp)
}

// GetDataOrders возвращает историю всех ордеров пользователя (GET /data/orders).
// Включает завершённые и отменённые ордера (не только открытые).
func (c *Client) GetDataOrders(filter OrdersFilter) (*OrdersResponse, error) {
	path := "/data/orders" + buildOrdersQuery(filter)
	resp, err := c.privateGet(path)
	if err != nil {
		return nil, fmt.Errorf("clob: GetDataOrders: %w", err)
	}
	return decode[OrdersResponse](resp)
}

// CancelMarketOrders отменяет все открытые ордера для указанного рынка или токена.
// Передайте marketID (condition_id) или assetID (token_id), или оба.
func (c *Client) CancelMarketOrders(marketID, assetID string) (*CancelOrderResponse, error) {
	body := map[string]string{}
	if marketID != "" {
		body["market"] = marketID
	}
	if assetID != "" {
		body["asset_id"] = assetID
	}

	raw, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("clob: CancelMarketOrders: marshal: %w", err)
	}
	resp, err := c.privateDelete("/cancel-market-orders", raw)
	if err != nil {
		return nil, fmt.Errorf("clob: CancelMarketOrders: %w", err)
	}
	return decode[CancelOrderResponse](resp)
}

// GetMarketTrades возвращает публичную историю сделок для токена (GET /trades?token_id=...).
func (c *Client) GetMarketTrades(tokenID string, limit int) (*TradesResponse, error) {
	path := "/trades?token_id=" + tokenID
	if limit > 0 {
		path += "&limit=" + strconv.Itoa(limit)
	}
	resp, err := c.privateGet(path)
	if err != nil {
		return nil, fmt.Errorf("clob: GetMarketTrades: %w", err)
	}
	return decode[TradesResponse](resp)
}

// GetNotifications возвращает системные уведомления пользователя.
func (c *Client) GetNotifications() ([]Notification, error) {
	resp, err := c.privateGet("/notifications")
	if err != nil {
		return nil, fmt.Errorf("clob: GetNotifications: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("clob: GetNotifications HTTP %d: %s", resp.StatusCode, resp.Body)
	}
	var notifs []Notification
	if err := json.Unmarshal(resp.Body, &notifs); err != nil {
		return nil, fmt.Errorf("clob: GetNotifications: decode: %w", err)
	}
	return notifs, nil
}

func buildTradesQuery(f TradesFilter) string {
	q := "?"
	if f.ID != "" {
		q += "id=" + f.ID + "&"
	}
	if f.Market != "" {
		q += "market=" + f.Market + "&"
	}
	if f.AssetID != "" {
		q += "asset_id=" + f.AssetID + "&"
	}
	if f.MakerAddress != "" {
		q += "maker_address=" + f.MakerAddress + "&"
	}
	if f.After > 0 {
		q += "after=" + strconv.FormatInt(f.After, 10) + "&"
	}
	if f.Cursor != "" {
		q += "next_cursor=" + f.Cursor + "&"
	}
	if f.Limit > 0 {
		q += "limit=" + strconv.Itoa(f.Limit) + "&"
	}
	if q == "?" {
		return ""
	}
	return q[:len(q)-1]
}

func buildOrdersQuery(f OrdersFilter) string {
	q := "?"
	if f.ID != "" {
		q += "id=" + f.ID + "&"
	}
	if f.Market != "" {
		q += "market=" + f.Market + "&"
	}
	if f.AssetID != "" {
		q += "asset_id=" + f.AssetID + "&"
	}
	if q == "?" {
		return ""
	}
	return q[:len(q)-1]
}
