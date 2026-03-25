package copytrading

import "math"

// SizeCalculator вычисляет размер нашей позиции при копировании сделки трейдера.
type SizeCalculator struct {
	mode           string  // "proportional" или "fixed_pct"
	allocationPct  float64 // % нашего баланса, выделяемый данному трейдеру
	maxPositionUSD float64 // максимальный размер одной позиции в USD (0 = без лимита)
}

// NewSizeCalculator создаёт калькулятор размера позиции.
func NewSizeCalculator(mode string, allocationPct, maxPositionUSD float64) *SizeCalculator {
	return &SizeCalculator{
		mode:           mode,
		allocationPct:  allocationPct,
		maxPositionUSD: maxPositionUSD,
	}
}

// Calculate вычисляет размер нашей позиции в USD.
// Возвращает округлённое до 2 знаков значение. Минимум $1.00.
func (c *SizeCalculator) Calculate(traderPositionUSD, traderTotalBalance, myBalance float64) float64 {
	var size float64

	switch c.mode {
	case "fixed_pct":
		size = myBalance * c.allocationPct / 100.0
	default: // "proportional"
		if traderTotalBalance <= 0 {
			return 0
		}
		ratio := traderPositionUSD / traderTotalBalance
		size = ratio * myBalance * c.allocationPct / 100.0
	}

	if c.maxPositionUSD > 0 && size > c.maxPositionUSD {
		size = c.maxPositionUSD
	}

	// Округление до 2 знаков (центов)
	size = math.Floor(size*100) / 100.0

	// Polymarket обычно требует минимум $1.00 для ордера.
	if size < 1.0 {
		return 0
	}

	return size
}

