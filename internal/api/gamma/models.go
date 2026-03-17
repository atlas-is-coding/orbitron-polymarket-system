// Package gamma реализует клиент для Polymarket Gamma API.
// Base URL: https://gamma-api.polymarket.com
// Gamma API предоставляет метаданные рынков, событий и не требует аутентификации.
package gamma

import (
	"encoding/json"
	"strconv"
)

type flexFloat64 float64

func (f *flexFloat64) UnmarshalJSON(b []byte) error {
	if len(b) > 0 && b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		if s == "" {
			*f = 0
			return nil
		}
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}
		*f = flexFloat64(v)
		return nil
	}
	var v float64
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	*f = flexFloat64(v)
	return nil
}

type flexStringSlice []string

func (f *flexStringSlice) UnmarshalJSON(b []byte) error {
	if len(b) > 0 && b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		if s == "" {
			return nil
		}
		var slice []string
		if err := json.Unmarshal([]byte(s), &slice); err != nil {
			return err
		}
		*f = flexStringSlice(slice)
		return nil
	}
	
	if string(b) == "null" {
		return nil
	}
	
	var slice []string
	if err := json.Unmarshal(b, &slice); err != nil {
		return err
	}
	*f = flexStringSlice(slice)
	return nil
}

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
	Liquidity   flexFloat64 `json:"liquidity"`
	// Объём (в USDC)
	Volume      flexFloat64 `json:"volume"`
	// Текущие вероятности YES/NO (от 0 до 1)
	OutcomePrices flexStringSlice `json:"outcomePrices"`
	// Названия исходов
	Outcomes    flexStringSlice `json:"outcomes"`
	// token_id для каждого исхода
	ClobTokenIDs flexStringSlice `json:"clobTokenIds"`
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
	Volume      flexFloat64  `json:"volume"`
	Liquidity   flexFloat64  `json:"liquidity"`
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
	Order     string // "volume_24hr", "volume", "liquidity", "end_date"
	Ascending bool   // sort direction; false = descending (default)
	Closed    *bool  // explicit closed filter
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
	Order     string
	Ascending bool
}
