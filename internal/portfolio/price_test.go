package portfolio

import (
	"testing"

	"event-driven-backtesting-engine/internal/domain"
)

func TestUpdatePrice(t *testing.T) {

	position := domain.NewPosition(
		1,
		1,
		"BTCUSDT",
		domain.BuyOrder,
		2,
		100,
		100,
	)

	UpdatePrice(&position, 120)

	if position.CurrentPrice != 120 {
		t.Fatalf(
			"expected current price 120, got %.2f",
			position.CurrentPrice,
		)
	}
}
