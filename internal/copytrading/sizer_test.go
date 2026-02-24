package copytrading_test

import (
	"testing"

	"github.com/atlasdev/polytrade-bot/internal/copytrading"
)

func TestSizeCalculatorProportional(t *testing.T) {
	calc := copytrading.NewSizeCalculator("proportional", 10.0, 100.0)

	// Трейдер: баланс 1000 USD, позиция 100 USDC (10% баланса)
	// Наш баланс: 500 USD, allocation: 10%
	// Ожидаем: 100/1000 * 500 * 0.10 = 5.0 USD
	size := calc.Calculate(100.0, 1000.0, 500.0)
	if size != 5.0 {
		t.Errorf("proportional: expected 5.0, got %f", size)
	}
}

func TestSizeCalculatorFixedPct(t *testing.T) {
	calc := copytrading.NewSizeCalculator("fixed_pct", 5.0, 100.0)

	// allocation_pct=5%, наш баланс=200 USD → 5% от 200 = 10.0
	// Позиция/баланс трейдера не влияют на результат
	size := calc.Calculate(999.0, 1000.0, 200.0)
	if size != 10.0 {
		t.Errorf("fixed_pct: expected 10.0, got %f", size)
	}
}

func TestSizeCalculatorMaxCap(t *testing.T) {
	calc := copytrading.NewSizeCalculator("fixed_pct", 50.0, 30.0)

	// 50% от 100 = 50, но max_position_usd=30 → должно вернуть 30
	size := calc.Calculate(0, 0, 100.0)
	if size != 30.0 {
		t.Errorf("max cap: expected 30.0, got %f", size)
	}
}

func TestSizeCalculatorZeroTraderBalance(t *testing.T) {
	calc := copytrading.NewSizeCalculator("proportional", 10.0, 50.0)

	// traderBalance=0 → деление на ноль защищено, вернуть 0
	size := calc.Calculate(100.0, 0.0, 500.0)
	if size != 0.0 {
		t.Errorf("zero trader balance: expected 0, got %f", size)
	}
}

func TestSizeCalculatorNoMaxCap(t *testing.T) {
	calc := copytrading.NewSizeCalculator("fixed_pct", 100.0, 0.0)

	// max=0 означает без лимита: 100% от 200 = 200
	size := calc.Calculate(0, 0, 200.0)
	if size != 200.0 {
		t.Errorf("no cap: expected 200.0, got %f", size)
	}
}
