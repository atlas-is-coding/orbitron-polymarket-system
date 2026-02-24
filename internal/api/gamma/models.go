// Package gamma реализует клиент для Polymarket Gamma API.
// Base URL: https://gamma-api.polymarket.com
// Gamma API предоставляет метаданные рынков, событий и не требует аутентификации.
package gamma

// Market — расширенные метаданные рынка из Gamma API.
type Market struct {
	ID          string  `json:"id"`
	ConditionID string  `json:"conditionId"`
	QuestionID  string  `json:"questionId"`
	Question    string  `json:"question"`
	Description string  `json:"description"`
	// Категория рынка (политика, спорт, крипто и т.д.)
	Category    string  `json:"category"`
	// Теги
	Tags        []Tag   `json:"tags"`
	// Статус
	Active      bool    `json:"active"`
	Closed      bool    `json:"closed"`
	Archived    bool    `json:"archived"`
	// Дата резолюции
	EndDateISO  string  `json:"endDateIso"`
	// Ликвидность (в USDC)
	Liquidity   float64 `json:"liquidity"`
	// Объём (в USDC)
	Volume      float64 `json:"volume"`
	// Текущие вероятности YES/NO (от 0 до 1)
	OutcomePrices []string `json:"outcomePrices"`
	// Названия исходов
	Outcomes    []string `json:"outcomes"`
	// token_id для каждого исхода
	ClobTokenIDs []string `json:"clobTokenIds"`
	// Обложка рынка
	Image       string  `json:"image"`
	// Ссылка на источник резолюции
	ResolutionSource string `json:"resolutionSource"`
	// Создатель
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
	// Связанное событие
	EventID     string  `json:"eventId"`
	// neg_risk
	NegRisk     bool    `json:"negRisk"`
}

// Tag — тег рынка.
type Tag struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Slug  string `json:"slug"`
}

// Event — событие, объединяющее несколько рынков.
type Event struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Slug        string   `json:"slug"`
	Category    string   `json:"category"`
	// Рынки в этом событии
	Markets     []Market `json:"markets"`
	// Статус события
	Active      bool     `json:"active"`
	Closed      bool     `json:"closed"`
	// Общий объём события
	Volume      float64  `json:"volume"`
	Liquidity   float64  `json:"liquidity"`
	// Изображение события
	Image       string   `json:"image"`
	StartDate   string   `json:"startDate"`
	EndDate     string   `json:"endDate"`
}

// MarketsParams — параметры фильтрации для GET /markets
type MarketsParams struct {
	// Статус: active, closed
	Active      *bool
	// Категория
	Category    string
	// Сортировка: volume, liquidity, endDate
	SortBy      string
	// Порядок: ASC, DESC
	SortOrder   string
	// Пагинация
	Offset      int
	Limit       int
}

// EventsParams — параметры для GET /events
type EventsParams struct {
	Active    *bool
	Category  string
	Offset    int
	Limit     int
}
