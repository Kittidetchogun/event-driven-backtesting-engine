package portfolio

import (
	"testing"
	"time"

	"event-driven-backtesting-engine/internal/domain"
)

func TestApplyCashUpdate_Buy(t *testing.T) {

	portfolio := domain.NewPortfolio(
		1,
		100000,
	)

	trade := domain.NewTrade(
		1,
		1,
		1,
		"BTCUSDT",
		domain.BuyOrder,
		1,
		50000,
		50,
		time.Now(),
	)

	ApplyCashUpdate(&portfolio, trade)

	expected := 49950.0

	if portfolio.Cash != expected {
		t.Fatalf(
			"expected %.2f got %.2f",
			expected,
			portfolio.Cash,
		)
	}
}

func TestApplyCashUpdate_Sell(t *testing.T) {

	portfolio := domain.NewPortfolio(
		1,
		100000,
	)

	trade := domain.NewTrade(
		1,
		1,
		1,
		"BTCUSDT",
		domain.SellOrder,
		1,
		50000,
		50,
		time.Now(),
	)

	ApplyCashUpdate(&portfolio, trade)

	expected := 149950.0

	if portfolio.Cash != expected {
		t.Fatalf(
			"expected %.2f got %.2f",
			expected,
			portfolio.Cash,
		)
	}
}
