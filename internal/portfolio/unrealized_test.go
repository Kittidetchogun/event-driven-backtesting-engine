package portfolio

import (
	"testing"

	"event-driven-backtesting-engine/internal/domain"
)

func TestUpdateUnrealizedPnL_Profit(t *testing.T) {

	position := domain.NewPosition(
		1,
		1,
		"BTCUSDT",
		domain.BuyOrder,
		2,
		100,
		120,
	)

	position.UnrealizedPnL = 0

	UpdateUnrealizedPnL(&position)

	expected := 40.0

	if position.UnrealizedPnL != expected {
		t.Fatalf(
			"expected %.2f got %.2f",
			expected,
			position.UnrealizedPnL,
		)
	}
}

func TestUpdateUnrealizedPnL_Loss(t *testing.T) {

	position := domain.NewPosition(
		1,
		1,
		"BTCUSDT",
		domain.BuyOrder,
		2,
		100,
		80,
	)

	position.UnrealizedPnL = 0

	UpdateUnrealizedPnL(&position)

	expected := -40.0

	if position.UnrealizedPnL != expected {
		t.Fatalf(
			"expected %.2f got %.2f",
			expected,
			position.UnrealizedPnL,
		)
	}
}

func TestUpdateUnrealizedPnL_Breakeven(t *testing.T) {

	position := domain.NewPosition(
		1,
		1,
		"BTCUSDT",
		domain.BuyOrder,
		5,
		100,
		100,
	)

	position.UnrealizedPnL = 999

	UpdateUnrealizedPnL(&position)

	if position.UnrealizedPnL != 0 {
		t.Fatalf(
			"expected 0 got %.2f",
			position.UnrealizedPnL,
		)
	}
}

func TestUpdateUnrealizedPnL_ZeroQuantity(t *testing.T) {

	position := domain.NewPosition(
		1,
		1,
		"BTCUSDT",
		domain.BuyOrder,
		0,
		100,
		120,
	)

	UpdateUnrealizedPnL(&position)

	if position.UnrealizedPnL != 0 {
		t.Fatalf(
			"expected 0 got %.2f",
			position.UnrealizedPnL,
		)
	}
}
