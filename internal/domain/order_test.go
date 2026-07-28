package domain

import (
	"testing"
	"time"
)

func TestNewOrder(t *testing.T) {
	now := time.Now()

	order := NewOrder(
		1,
		"BTCUSDT",
		BuyOrder,
		0.5,
		100000,
		now,
	)

	if order.RunID != 1 {
		t.Fatalf("expected RunID=1, got %d", order.RunID)
	}

	if order.Symbol != "BTCUSDT" {
		t.Fatalf("expected symbol BTCUSDT, got %s", order.Symbol)
	}

	if order.Side != BuyOrder {
		t.Fatalf("expected BUY, got %s", order.Side)
	}

	if order.Type != MarketOrder {
		t.Fatalf("expected MARKET, got %s", order.Type)
	}

	if order.Quantity != 0.5 {
		t.Fatalf("expected quantity 0.5, got %f", order.Quantity)
	}

	if order.Price != 100000 {
		t.Fatalf("expected price 100000, got %f", order.Price)
	}

	if order.Status != PendingOrder {
		t.Fatalf("expected status PENDING, got %s", order.Status)
	}

	if !order.CreatedAt.Equal(now) {
		t.Fatal("created time mismatch")
	}

	if order.FilledAt != nil {
		t.Fatal("FilledAt should be nil")
	}
}
