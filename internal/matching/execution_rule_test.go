package matching

import (
	"testing"
	"time"

	"event-driven-backtesting-engine/internal/domain"
)

func TestExecutionPriceUsesCurrentClosePrice(t *testing.T) {

	order := domain.NewOrder(
		1,
		"BTCUSDT",
		domain.BuyOrder,
		1,
		50000,
		time.Now(),
	)

	price := executionPrice(order)

	if price != order.Price {
		t.Fatalf(
			"expected %.2f got %.2f",
			order.Price,
			price,
		)
	}
}