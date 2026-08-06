package portfolio

import (
	"testing"

	"event-driven-backtesting-engine/internal/domain"
)

func TestUpdateMarketValue(t *testing.T) {

	position := domain.NewPosition(
		1,
		1,
		"BTCUSDT",
		domain.BuyOrder,
		2,
		100,
		120,
	)

	position.CurrentValue = 0

	UpdateMarketValue(&position)

	expected := 240.0

	if position.CurrentValue != expected {
		t.Fatalf(
			"expected %.2f got %.2f",
			expected,
			position.CurrentValue,
		)
	}
}

func TestUpdateMarketValue_ZeroQuantity(t *testing.T) {

	position := domain.NewPosition(
		1,
		1,
		"BTCUSDT",
		domain.BuyOrder,
		0,
		100,
		120,
	)

	UpdateMarketValue(&position)

	if position.CurrentValue != 0 {
		t.Fatalf(
			"expected 0 got %.2f",
			position.CurrentValue,
		)
	}
}
