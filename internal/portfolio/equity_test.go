package portfolio

import (
	"testing"

	"event-driven-backtesting-engine/internal/domain"
)

func TestUpdateEquity(t *testing.T) {

	portfolio := domain.NewPortfolio(
		1,
		100000,
	)

	portfolio.Cash = 40000
	portfolio.PositionValue = 75000

	UpdateEquity(&portfolio)

	expected := 115000.0

	if portfolio.Equity != expected {
		t.Fatalf(
			"expected %.2f got %.2f",
			expected,
			portfolio.Equity,
		)
	}
}

func TestUpdateEquity_OnlyCash(t *testing.T) {

	portfolio := domain.NewPortfolio(
		1,
		100000,
	)

	portfolio.Cash = 50000
	portfolio.PositionValue = 0

	UpdateEquity(&portfolio)

	if portfolio.Equity != 50000 {
		t.Fatalf(
			"expected %.2f got %.2f",
			50000.0,
			portfolio.Equity,
		)
	}
}

func TestUpdateEquity_OnlyPosition(t *testing.T) {

	portfolio := domain.NewPortfolio(
		1,
		0,
	)

	portfolio.Cash = 0
	portfolio.PositionValue = 12000

	UpdateEquity(&portfolio)

	if portfolio.Equity != 12000 {
		t.Fatalf(
			"expected %.2f got %.2f",
			12000.0,
			portfolio.Equity,
		)
	}
}

func TestUpdateEquity_Zero(t *testing.T) {

	portfolio := domain.NewPortfolio(
		1,
		0,
	)

	UpdateEquity(&portfolio)

	if portfolio.Equity != 0 {
		t.Fatalf(
			"expected 0 got %.2f",
			portfolio.Equity,
		)
	}
}
