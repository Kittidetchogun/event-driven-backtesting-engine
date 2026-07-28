package events

import (
	"testing"
	"time"

	"event-driven-backtesting-engine/internal/domain"
)

func TestNewOrderCreatedEvent(t *testing.T) {
	now := time.Now()

	order := domain.NewOrder(
		1,
		"BTCUSDT",
		domain.BuyOrder,
		0.5,
		100000,
		now,
	)

	event := NewOrderCreatedEvent(order)

	if event.Type() != OrderCreatedEventType {
		t.Fatalf("expected event type %s, got %s",
			OrderCreatedEventType,
			event.Type(),
		)
	}

	if !event.Timestamp().Equal(now) {
		t.Fatal("timestamp mismatch")
	}

	if event.Order.Symbol != "BTCUSDT" {
		t.Fatalf("expected BTCUSDT, got %s", event.Order.Symbol)
	}

	if event.Order.Side != domain.BuyOrder {
		t.Fatalf("expected BUY, got %s", event.Order.Side)
	}
}
