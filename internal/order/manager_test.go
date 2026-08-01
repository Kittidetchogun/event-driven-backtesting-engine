package order

import (
	"testing"
	"time"

	"event-driven-backtesting-engine/internal/domain"
	"event-driven-backtesting-engine/internal/events"
)

func TestOrderManagerConsume(t *testing.T) {
	queue := events.NewEventQueue()

	manager := NewManager(
		domain.DummyPortfolioChecker{},
		queue,
	)

	signal := events.NewSignalGeneratedEvent(
		1,               // RunID
		"BTCUSDT",       // Symbol
		domain.BuyOrder, // SignalType
		1,               // Quantity
		time.Now(),      // SignalTime
	)

	if err := manager.Consume(signal); err != nil {
		t.Fatal(err)
	}

	if queue.Len() != 1 {
		t.Fatalf("expected queue length 1 got %d", queue.Len())
	}

	event, ok := queue.Pop()
	if !ok {
		t.Fatal("expected event")
	}

	orderEvent, ok := event.(events.OrderCreatedEvent)
	if !ok {
		t.Fatal("expected OrderCreatedEvent")
	}

	if orderEvent.Order.Symbol != "BTCUSDT" {
		t.Fatalf("expected symbol BTCUSDT got %s", orderEvent.Order.Symbol)
	}

	if orderEvent.Order.Side != domain.BuyOrder {
		t.Fatalf("expected side BUY got %s", orderEvent.Order.Side)
	}

	if orderEvent.Order.Status != domain.PendingOrder {
		t.Fatalf("expected status PENDING got %s", orderEvent.Order.Status)
	}
}