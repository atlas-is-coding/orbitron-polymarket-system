package clob

import "fmt"

// GetMarkets возвращает список рынков с пагинацией.
// cursor — курсор пагинации (пустая строка для первой страницы).
func (c *Client) GetMarkets(cursor string) (*MarketsResponse, error) {
	path := "/markets"
	if cursor != "" {
		path += "?next_cursor=" + cursor
	}
	resp, err := c.publicGet(path)
	if err != nil {
		return nil, fmt.Errorf("clob: GetMarkets: %w", err)
	}
	return decode[MarketsResponse](resp)
}

// GetMarket возвращает рынок по condition_id.
func (c *Client) GetMarket(conditionID string) (*Market, error) {
	resp, err := c.publicGet("/markets/" + conditionID)
	if err != nil {
		return nil, fmt.Errorf("clob: GetMarket: %w", err)
	}
	return decode[Market](resp)
}

// GetAllMarkets возвращает все рынки, проходя по всем страницам пагинации.
func (c *Client) GetAllMarkets() ([]Market, error) {
	var all []Market
	cursor := ""
	for {
		page, err := c.GetMarkets(cursor)
		if err != nil {
			return nil, err
		}
		all = append(all, page.Data...)
		if page.NextCursor == "" || page.NextCursor == "LTE=" {
			break
		}
		cursor = page.NextCursor
	}
	return all, nil
}
