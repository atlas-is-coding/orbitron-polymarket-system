package clob

import (
	"encoding/json"
	"fmt"
)

// CreateOrder размещает подписанный ордер через POST /order.
func (c *Client) CreateOrder(req *CreateOrderRequest) (*CreateOrderResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("clob: CreateOrder: marshal: %w", err)
	}
	resp, err := c.privatePost("/order", body)
	if err != nil {
		return nil, fmt.Errorf("clob: CreateOrder: %w", err)
	}
	return decode[CreateOrderResponse](resp)
}

// CancelOrder отменяет ордер по ID.
func (c *Client) CancelOrder(orderID string) (*CancelOrderResponse, error) {
	resp, err := c.privateDelete("/order/"+orderID, nil)
	if err != nil {
		return nil, fmt.Errorf("clob: CancelOrder: %w", err)
	}
	return decode[CancelOrderResponse](resp)
}

// CancelOrders отменяет несколько ордеров по ID.
func (c *Client) CancelOrders(orderIDs []string) (*CancelOrderResponse, error) {
	body, err := json.Marshal(map[string][]string{"orderIDs": orderIDs})
	if err != nil {
		return nil, fmt.Errorf("clob: CancelOrders: marshal: %w", err)
	}
	resp, err := c.privateDelete("/orders", body)
	if err != nil {
		return nil, fmt.Errorf("clob: CancelOrders: %w", err)
	}
	return decode[CancelOrderResponse](resp)
}

// CancelAllOrders отменяет все открытые ордера.
func (c *Client) CancelAllOrders() (*CancelOrderResponse, error) {
	resp, err := c.privateDelete("/cancel-all", nil)
	if err != nil {
		return nil, fmt.Errorf("clob: CancelAllOrders: %w", err)
	}
	return decode[CancelOrderResponse](resp)
}

// GetOrders возвращает список открытых ордеров пользователя с опциональным фильтром.
func (c *Client) GetOrders(filter ...OrdersFilter) (*OrdersResponse, error) {
	path := "/orders"
	if len(filter) > 0 {
		path += buildOrdersQuery(filter[0])
	}
	resp, err := c.privateGet(path)
	if err != nil {
		return nil, fmt.Errorf("clob: GetOrders: %w", err)
	}
	return decode[OrdersResponse](resp)
}

// GetOrder возвращает ордер по ID.
func (c *Client) GetOrder(orderID string) (*Order, error) {
	resp, err := c.privateGet("/order?id=" + orderID)
	if err != nil {
		return nil, fmt.Errorf("clob: GetOrder: %w", err)
	}
	return decode[Order](resp)
}

// GetTrades возвращает историю сделок пользователя с опциональным фильтром.
func (c *Client) GetTrades(filter ...TradesFilter) (*TradesResponse, error) {
	path := "/trades"
	if len(filter) > 0 {
		path += buildTradesQuery(filter[0])
	}
	resp, err := c.privateGet(path)
	if err != nil {
		return nil, fmt.Errorf("clob: GetTrades: %w", err)
	}
	return decode[TradesResponse](resp)
}
