// Package data реализует клиент для Polymarket Data API.
// Base URL: https://data-api.polymarket.com
// Публичный API — аутентификация не требуется.
// Даёт доступ к позициям пользователей, истории сделок и активности.
package data

// Position — открытая позиция пользователя по рынку.
type Position struct {
	// Адрес proxy-кошелька пользователя
	ProxyWallet string `json:"proxyWallet"`
	// Token ID позиции (ERC-1155)
	Asset string `json:"asset"`
	// Condition ID рынка
	ConditionID string `json:"conditionId"`
	// Размер позиции (количество токенов)
	Size float64 `json:"size"`
	// Средняя цена покупки
	AvgPrice float64 `json:"avgPrice"`
	// Начальная стоимость в USDC
	InitialValue float64 `json:"initialValue"`
	// Текущая стоимость в USDC
	CurrentValue float64 `json:"currentValue"`
	// P&L в USDC
	CashPnl float64 `json:"cashPnl"`
	// P&L в процентах
	PercentPnl float64 `json:"percentPnl"`
	// Суммарно куплено токенов
	TotalBought float64 `json:"totalBought"`
	// Реализованный P&L
	RealizedPnl float64 `json:"realizedPnl"`
	// Нереализованный P&L
	UnrealizedPnl float64 `json:"unrealizedPnl"`
	// Текущая цена токена
	CurPrice float64 `json:"curPrice"`
	// Название рынка
	Title string `json:"title"`
	// Slug рынка
	Slug string `json:"slug"`
	// URL иконки
	Icon string `json:"icon"`
	// Slug события
	EventSlug string `json:"eventSlug"`
	// Исход: "Yes" или "No"
	Outcome string `json:"outcome"`
	// Индекс исхода: 0 = Yes, 1 = No
	OutcomeIndex int `json:"outcomeIndex"`
	// Дата окончания рынка
	EndDate string `json:"endDate"`
}

// ClosedPosition — закрытая позиция пользователя.
type ClosedPosition struct {
	// Адрес proxy-кошелька
	ProxyWallet string `json:"proxyWallet"`
	// Token ID
	Asset string `json:"asset"`
	// Condition ID рынка
	ConditionID string `json:"conditionId"`
	// Итоговый P&L
	Pnl float64 `json:"pnl"`
	// Исход: "Yes" / "No"
	Outcome string `json:"outcome"`
	// Название рынка
	Title string `json:"title"`
	// Slug рынка
	Slug string `json:"slug"`
	// Дата резолюции
	ResolvedAt string `json:"resolvedAt"`
	// Победивший исход
	WinnerOutcome string `json:"winnerOutcome"`
}

// Trade — сделка из Data API.
type Trade struct {
	ID             string  `json:"id"`
	TakerOrderID   string  `json:"taker_order_id"`
	Market         string  `json:"market"`
	AssetID        string  `json:"asset_id"`
	Side           string  `json:"side"`
	Size           string  `json:"size"`
	FeeRateBps     string  `json:"fee_rate_bps"`
	Price          string  `json:"price"`
	Status         string  `json:"status"`
	MatchTime      string  `json:"match_time"`
	LastUpdate     string  `json:"last_update"`
	Outcome        string  `json:"outcome"`
	Owner          string  `json:"owner"`
	MakerAddress   string  `json:"maker_address"`
	TraderSide     string  `json:"trader_side"`
	TransactionHash string `json:"transaction_hash"`
	// Название рынка (обогащённые данные)
	Title          string  `json:"title"`
	Slug           string  `json:"slug"`
	// Размер в числовом формате для удобства
	SizeFloat      float64 `json:"size_float,omitempty"`
	PriceFloat     float64 `json:"price_float,omitempty"`
}

// PositionsParams — параметры запроса позиций.
type PositionsParams struct {
	// Адрес кошелька (proxy wallet)
	User string
	// Сортировка: "value", "pnl", "size"
	SortBy string
	// Порядок: "asc", "desc"
	SortOrder string
	// Только с размером > 0
	SizeThreshold float64
	// Лимит результатов
	Limit int
	// Смещение
	Offset int
}

// BuilderAnalytics — статистика Builders Program из Data API.
type BuilderAnalytics struct {
	// Адрес/ключ билдера
	BuilderAddress string  `json:"builderAddress"`
	// Суммарный объём торгов, атрибутированный билдеру (USDC)
	TotalVolume    float64 `json:"totalVolume"`
	// Сборы, заработанные билдером (USDC)
	TotalFees      float64 `json:"totalFees"`
	// Количество сделок
	TradeCount     int     `json:"tradeCount"`
	// Количество уникальных пользователей
	UniqueUsers    int     `json:"uniqueUsers"`
}

// TradesParams — параметры запроса сделок.
type TradesParams struct {
	// Адрес кошелька
	User string
	// Маркет (condition_id)
	Market string
	// Token ID
	AssetID string
	// Лимит
	Limit int
	// Смещение
	Offset int
}
