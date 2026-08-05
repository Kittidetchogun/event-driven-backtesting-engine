package portfolio

import (
	"testing"
	"time"

	"event-driven-backtesting-engine/internal/domain"
)

func TestUpdatePosition_Create(t *testing.T) {

	positions := make(map[string]domain.Position)

	trade := domain.NewTrade(
		1,
		1,
		1,
		"BTCUSDT",
		domain.BuyOrder,
		2,
		50000,
		0,
		time.Now(),
	)

	UpdatePosition(
		positions,
		1,
		trade,
	)

	position, ok := positions["BTCUSDT"]

	if !ok {
		t.Fatal("expected position")
	}

	if position.Quantity != 2 {
		t.Fatalf("expected quantity 2 got %.2f", position.Quantity)
	}

	if position.AveragePrice != 50000 {
		t.Fatalf("expected avg price 50000")
	}
}

func TestUpdatePosition_BuyMore(t *testing.T) {

	positions := make(map[string]domain.Position)

	positions["BTCUSDT"] = domain.NewPosition(
		1,
		1,
		"BTCUSDT",
		domain.BuyOrder,
		1,
		50000,
		50000,
	)

	trade := domain.NewTrade(
		2,
		1,
		2,
		"BTCUSDT",
		domain.BuyOrder,
		1,
		60000,
		0,
		time.Now(),
	)

	UpdatePosition(
		positions,
		1,
		trade,
	)

	position := positions["BTCUSDT"]

	if position.Quantity != 2 {
		t.Fatalf("expected quantity 2")
	}

	expectedAvg := 55000.0

	if position.AveragePrice != expectedAvg {
		t.Fatalf(
			"expected avg %.2f got %.2f",
			expectedAvg,
			position.AveragePrice,
		)
	}
}

func TestUpdatePosition_SellPartial(t *testing.T) {

	positions := make(map[string]domain.Position)

	positions["BTCUSDT"] = domain.NewPosition(
		1,
		1,
		"BTCUSDT",
		domain.BuyOrder,
		2,
		50000,
		50000,
	)

	trade := domain.NewTrade(
		2,
		1,
		2,
		"BTCUSDT",
		domain.SellOrder,
		1,
		52000,
		0,
		time.Now(),
	)

	UpdatePosition(
		positions,
		1,
		trade,
	)

	position := positions["BTCUSDT"]

	if position.Quantity != 1 {
		t.Fatalf("expected quantity 1")
	}
}

func TestUpdatePosition_ClosePosition(t *testing.T) {

	positions := make(map[string]domain.Position)

	positions["BTCUSDT"] = domain.NewPosition(
		1,
		1,
		"BTCUSDT",
		domain.BuyOrder,
		1,
		50000,
		50000,
	)

	trade := domain.NewTrade(
		2,
		1,
		2,
		"BTCUSDT",
		domain.SellOrder,
		1,
		51000,
		0,
		time.Now(),
	)

	UpdatePosition(
		positions,
		1,
		trade,
	)

	if _, ok := positions["BTCUSDT"]; ok {
		t.Fatal("expected position removed")
	}
}
