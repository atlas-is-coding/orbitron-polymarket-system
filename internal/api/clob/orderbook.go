package clob

import "fmt"

// GetOrderBook возвращает книгу ордеров для токена по token_id.
func (c *Client) GetOrderBook(tokenID string) (*OrderBook, error) {
	resp, err := c.publicGet("/order-book/" + tokenID)
	if err != nil {
		return nil, fmt.Errorf("clob: GetOrderBook: %w", err)
	}
	return decode[OrderBook](resp)
}

// GetMidpoint возвращает среднюю цену между лучшим bid и ask.
func (c *Client) GetMidpoint(tokenID string) (*Midpoint, error) {
	resp, err := c.publicGet("/midpoint?token_id=" + tokenID)
	if err != nil {
		return nil, fmt.Errorf("clob: GetMidpoint: %w", err)
	}
	return decode[Midpoint](resp)
}

// GetPrice возвращает лучшую цену для заданной стороны.
// side: "BUY" или "SELL"
func (c *Client) GetPrice(tokenID, side string) (*Price, error) {
	path := fmt.Sprintf("/price?token_id=%s&side=%s", tokenID, side)
	resp, err := c.publicGet(path)
	if err != nil {
		return nil, fmt.Errorf("clob: GetPrice: %w", err)
	}
	return decode[Price](resp)
}

// GetSpread возвращает спред между лучшим ask и bid.
func (c *Client) GetSpread(tokenID string) (*Spread, error) {
	resp, err := c.publicGet("/spread?token_id=" + tokenID)
	if err != nil {
		return nil, fmt.Errorf("clob: GetSpread: %w", err)
	}
	return decode[Spread](resp)
}
