package portfolio

import (
	"time"
	"testing"

	"event-driven-backtesting-engine/internal/domain"
)

func TestUpdateRealizedPnL_Profit(t *testing.T) {

	position := domain.NewPosition(
		1,
		1,
		"BTCUSDT",
		domain.BuyOrder,
		10,
		100,
		100,
	)

	trade := domain.NewTrade(
		1,
		1,
		1,
		"BTCUSDT",
		domain.SellOrder,
		5,
		120,
		5,
		time.Now(),
	)

	UpdateRealizedPnL(&position, trade)

	expected := (120.0-100.0)*5 - 5

	if position.RealizedPnL != expected {
		t.Fatalf(
			"expected %.2f got %.2f",
			expected,
			position.RealizedPnL,
		)
	}
}

func TestUpdateRealizedPnL_Loss(t *testing.T) {

	position := domain.NewPosition(
		1,
		1,
		"BTCUSDT",
		domain.BuyOrder,
		10,
		100,
		100,
	)

	trade := domain.NewTrade(
		1,
		1,
		1,
		"BTCUSDT",
		domain.SellOrder,
		5,
		80,
		5,
		time.Now(),
	)

	UpdateRealizedPnL(&position, trade)

	expected := (80.0-100.0)*5 - 5

	if position.RealizedPnL != expected {
		t.Fatalf(
			"expected %.2f got %.2f",
			expected,
			position.RealizedPnL,
		)
	}
}

func TestUpdateRealizedPnL_BuyTrade(t *testing.T) {

	position := domain.NewPosition(
		1,
		1,
		"BTCUSDT",
		domain.BuyOrder,
		10,
		100,
		100,
	)

	trade := domain.NewTrade(
		1,
		1,
		1,
		"BTCUSDT",
		domain.BuyOrder,
		5,
		120,
		5,
		time.Now(),
	)

	UpdateRealizedPnL(&position, trade)

	if position.RealizedPnL != 0 {
		t.Fatalf(
			"expected 0 got %.2f",
			position.RealizedPnL,
		)
	}
}
