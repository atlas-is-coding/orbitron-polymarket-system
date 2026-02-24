// Package clob реализует клиент для Polymarket CLOB API.
// Base URL: https://clob.polymarket.com
// Документация: https://docs.polymarket.com/#clob-api
package clob

// --- Рынки (Markets) ---

// Market описывает предсказательный рынок в CLOB.
type Market struct {
	// condition_id — уникальный идентификатор рынка (hex)
	ConditionID string `json:"condition_id"`
	// Адрес CTF Exchange контракта
	QuestionID string `json:"question_id"`
	// Токены YES/NO
	Tokens []Token `json:"tokens"`
	// Минимальный размер ордера в USDC
	MinimumOrderSize float64 `json:"minimum_order_size"`
	// Минимальный тик (шаг цены)
	MinimumTickSize float64 `json:"minimum_tick_size"`
	// Статус рынка: active, closed, resolved
	Active       bool   `json:"active"`
	Closed       bool   `json:"closed"`
	Description  string `json:"description"`
	// neg_risk — для рынков с взаимоисключающими исходами
	NegRisk bool `json:"neg_risk"`
}

// Token представляет исходный токен (YES или NO) рынка.
type Token struct {
	// token_id — ERC-1155 token ID (uint256 в виде строки)
	TokenID string `json:"token_id"`
	// outcome — "YES" или "NO"
	Outcome string `json:"outcome"`
	// Текущая цена (от 0 до 1, т.е. от 0¢ до 100¢)
	Price float64 `json:"price"`
	// Победивший исход (после резолюции)
	Winner bool `json:"winner"`
}

// MarketsResponse — ответ на GET /markets
type MarketsResponse struct {
	Data       []Market `json:"data"`
	NextCursor string   `json:"next_cursor"`
	Limit      int      `json:"limit"`
	Count      int      `json:"count"`
}

// --- Книга ордеров (Order Book) ---

// OrderBook представляет снимок книги ордеров для токена.
type OrderBook struct {
	// asset_id = token_id
	AssetID string       `json:"asset_id"`
	Bids    []PriceLevel `json:"bids"`
	Asks    []PriceLevel `json:"asks"`
	Hash    string       `json:"hash"`
}

// PriceLevel — уровень цены с объёмом.
type PriceLevel struct {
	Price string `json:"price"` // строка для точности (напр. "0.65")
	Size  string `json:"size"`  // объём в USDC
}

// Midpoint — средняя точка между лучшим bid и ask.
type Midpoint struct {
	Mid string `json:"mid"`
}

// Price — лучшая цена для заданной стороны.
type Price struct {
	Price string `json:"price"`
}

// Spread — спред между лучшим ask и bid.
type Spread struct {
	Spread string `json:"spread"`
}

// --- Ордера (Orders) ---

// OrderSide — сторона ордера.
type OrderSide string

const (
	SideBuy  OrderSide = "BUY"
	SideSell OrderSide = "SELL"
)

// OrderType — тип ордера.
type OrderType string

const (
	OrderTypeGTC OrderType = "GTC" // Good Till Cancel
	OrderTypeFOK OrderType = "FOK" // Fill or Kill
	OrderTypeGTD OrderType = "GTD" // Good Till Date
)

// OrderStatus — текущий статус ордера.
type OrderStatus string

const (
	StatusLive      OrderStatus = "LIVE"
	StatusMatched   OrderStatus = "MATCHED"
	StatusCanceled  OrderStatus = "CANCELED"
	StatusRetrying  OrderStatus = "RETRYING"
	StatusUnmatched OrderStatus = "UNMATCHED"
)

// Order — ордер в книге ордеров.
type Order struct {
	// ID ордера (UUID)
	ID string `json:"id"`
	// Статус
	Status OrderStatus `json:"status"`
	// Токен
	AssetID string `json:"asset_id"`
	// Сторона
	Side OrderSide `json:"side"`
	// Тип
	OrderType OrderType `json:"type"`
	// Цена (0-1)
	Price string `json:"price"`
	// Исходный размер
	OriginalSize string `json:"original_size"`
	// Оставшийся размер
	SizeRemaining string `json:"size_remaining"`
	// Исполненный размер
	SizeFilled string `json:"size_filled"`
	// Maker/Taker сборы
	MakerAmount string `json:"maker_amount"`
	TakerAmount string `json:"taker_amount"`
	// Адрес создателя
	Maker string `json:"maker"`
	// Время создания (unix ms)
	CreatedAt int64 `json:"created_at"`
	// Время экспирации (для GTD)
	ExpiresAt int64 `json:"expires_at,omitempty"`
}

// OrdersResponse — ответ на GET /orders
type OrdersResponse struct {
	Data       []Order `json:"data"`
	NextCursor string  `json:"next_cursor"`
	Count      int     `json:"count"`
}

// --- Создание ордера ---

// CreateOrderRequest — тело запроса POST /order
type CreateOrderRequest struct {
	// Подписанный ордер (EIP-712)
	Order     SignedOrder `json:"order"`
	Owner     string     `json:"owner"`
	OrderType OrderType  `json:"orderType"`
}

// SignedOrder — подписанный EIP-712 ордер для POST /order.
type SignedOrder struct {
	// Salt — случайное число для уникальности ордера
	Salt string `json:"salt"`
	// Maker — адрес создателя (proxy wallet)
	Maker string `json:"maker"`
	// Signer — адрес подписанта (обычно совпадает с Maker)
	Signer string `json:"signer"`
	// Taker — обычно нулевой адрес
	Taker string `json:"taker"`
	// TokenID — token_id токена YES/NO (ERC-1155)
	TokenID string `json:"tokenId"`
	// MakerAmount — USDC (BUY) или токены (SELL) в base units (6 decimals)
	MakerAmount string `json:"makerAmount"`
	// TakerAmount — токены (BUY) или USDC (SELL) в base units (6 decimals)
	TakerAmount string `json:"takerAmount"`
	// Expiration — unix timestamp (0 для GTC)
	Expiration string `json:"expiration"`
	// Nonce
	Nonce string `json:"nonce"`
	// FeeRateBps — сбор в базисных пунктах
	FeeRateBps string `json:"feeRateBps"`
	// Side — 0=BUY, 1=SELL
	Side int `json:"side"`
	// SignatureType — 0=EOA, 1=POLY_PROXY, 2=POLY_GNOSIS_SAFE
	SignatureType int `json:"signatureType"`
	// Signature — EIP-712 подпись
	Signature string `json:"signature"`
}

// CreateOrderResponse — ответ на POST /order
type CreateOrderResponse struct {
	Success       bool   `json:"success"`
	ErrorMsg      string `json:"errorMsg,omitempty"`
	OrderID       string `json:"orderID"`
	TransactionHash string `json:"transactionsHashes,omitempty"`
	Status        OrderStatus `json:"status"`
}

// CancelOrderResponse — ответ на DELETE /order/{id}
type CancelOrderResponse struct {
	Canceled bool   `json:"canceled"`
	OrderID  string `json:"order_id"`
}

// --- Сделки (Trades) ---

// Trade — исполненная сделка.
type Trade struct {
	ID             string    `json:"id"`
	TakerOrderID   string    `json:"taker_order_id"`
	MakerOrderID   string    `json:"maker_order_id"`
	TradeID        int64     `json:"trade_id"`
	Status         string    `json:"status"`
	Outcome        string    `json:"outcome"`
	Price          string    `json:"price"`
	Size           string    `json:"size"`
	AssetID        string    `json:"asset_id"`
	MakerAssetID   string    `json:"maker_asset_id"`
	TakerAssetID   string    `json:"taker_asset_id"`
	Side           OrderSide `json:"side"`
	Timestamp      int64     `json:"timestamp"`
	TransactionHash string   `json:"transaction_hash,omitempty"`
	FeeRateBps     string    `json:"fee_rate_bps"`
	MatchTime      string    `json:"match_time"`
	LastUpdate     string    `json:"last_update"`
}

// TradesResponse — ответ на GET /trades
type TradesResponse struct {
	Data       []Trade `json:"data"`
	NextCursor string  `json:"next_cursor"`
	Count      int     `json:"count"`
}

// --- Позиции ---

// Position — текущая открытая позиция пользователя (CLOB API, GET /positions).
type Position struct {
	AssetID      string  `json:"asset_id"`
	ConditionID  string  `json:"condition_id"`
	Outcome      string  `json:"outcome"`
	Size         float64 `json:"size"`
	AveragePrice float64 `json:"average_price"`
	InitialValue float64 `json:"initial_value"`
	CurrentValue float64 `json:"current_value"`
	PnL          float64 `json:"pnl"`
	RealizedPnL  float64 `json:"realized_pnl"`
}

// --- Фильтры запросов ---

// OrdersFilter — параметры фильтрации GET /orders (открытые ордера).
type OrdersFilter struct {
	// ID — фильтр по ID ордера
	ID string
	// Market — condition_id рынка
	Market string
	// AssetID — token_id
	AssetID string
}

// TradesFilter — параметры фильтрации GET /trades и GET /data/trades.
type TradesFilter struct {
	// ID — ID сделки
	ID string
	// Market — condition_id рынка
	Market string
	// AssetID — token_id токена
	AssetID string
	// MakerAddress — адрес maker'a
	MakerAddress string
	// After — unix timestamp, вернуть сделки после этого времени
	After int64
	// Cursor — курсор пагинации (для /data/trades)
	Cursor string
	// Limit — максимальное количество результатов
	Limit int
}

// --- Баланс и лимиты ---

// BalanceAllowance — баланс и разрешения для токена.
type BalanceAllowance struct {
	// AssetType: "COLLATERAL" (USDC) или "CONDITIONAL" (токен YES/NO)
	AssetType string `json:"asset_type"`
	// TokenID — для CONDITIONAL токена
	TokenID string `json:"token_id,omitempty"`
	// Баланс в wei (6 decimals для USDC)
	Balance string `json:"balance"`
	// Разрешение для Exchange контракта
	Allowance string `json:"allowance"`
}

// --- MakerOrder (вложен в Trade) ---

// MakerOrder — вложенный объект в Trade, описывает maker ордер.
type MakerOrder struct {
	OrderID       string `json:"order_id"`
	Owner         string `json:"owner"`
	MakerAddress  string `json:"maker_address"`
	MatchedAmount string `json:"matched_amount"`
	Price         string `json:"price"`
	FeeRateBps    string `json:"fee_rate_bps"`
	AssetID       string `json:"asset_id"`
	Outcome       string `json:"outcome"`
	Side          string `json:"side"`
}

// --- Уведомления ---

// Notification — системное уведомление от CLOB (GET /notifications).
type Notification struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Market    string `json:"market"`
	AssetID   string `json:"asset_id"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}
