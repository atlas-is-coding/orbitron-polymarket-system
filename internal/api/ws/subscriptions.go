package ws

import "github.com/atlasdev/orbitron/internal/auth"

// MarketSubscription создаёт запрос подписки на обновления книги ордеров рынков.
// markets — список condition_id рынков.
func MarketSubscription(markets []string) SubscribeRequest {
	return SubscribeRequest{
		Type:    ChannelMarket,
		Markets: &markets,
	}
}

// AssetSubscription создаёт запрос подписки на обновления цен токенов.
// assets — список token_id.
func AssetSubscription(assets []string) SubscribeRequest {
	return SubscribeRequest{
		Type:   ChannelAsset,
		Assets: &assets,
	}
}

// UserSubscription создаёт запрос подписки на события пользователя (ордера, трейды).
// Требует L2 credentials.
func UserSubscription(creds *auth.L2Credentials) SubscribeRequest {
	return SubscribeRequest{
		Type: ChannelUser,
		Auth: &AuthPayload{
			APIKey:     creds.APIKey,
			Secret:     creds.APISecret,
			Passphrase: creds.Passphrase,
		},
	}
}

// --- Типы событий из WebSocket ---

// OrderBookUpdate — обновление книги ордеров (channel: market/asset).
type OrderBookUpdate struct {
	AssetID string       `json:"asset_id"`
	Buys    []PriceLevel `json:"buys"`
	Sells   []PriceLevel `json:"sells"`
}

// PriceLevel — уровень цены.
type PriceLevel struct {
	Price string `json:"price"`
	Size  string `json:"size"`
}

// UserOrderEvent — событие ордера пользователя (channel: user).
type UserOrderEvent struct {
	OrderID      string `json:"order_id"`
	Status       string `json:"status"`
	AssetID      string `json:"asset_id"`
	Side         string `json:"side"`
	Price        string `json:"price"`
	OriginalSize string `json:"original_size"`
	SizeMatched  string `json:"size_matched"`
}

// UserTradeEvent — событие сделки пользователя (channel: user).
type UserTradeEvent struct {
	TradeID   string `json:"trade_id"`
	OrderID   string `json:"taker_order_id"`
	AssetID   string `json:"asset_id"`
	Side      string `json:"side"`
	Price     string `json:"price"`
	Size      string `json:"size"`
	Timestamp int64  `json:"timestamp"`
}
