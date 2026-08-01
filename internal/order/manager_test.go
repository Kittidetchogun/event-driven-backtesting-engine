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
		1,
		"TestStrategy",
		"BTCUSDT",
		"1d",
		events.SignalBuy,
		time.Now(),
		1,
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

	_, ok = event.(events.OrderCreatedEvent)

	if !ok {
		t.Fatal("expected OrderCreatedEvent")
	}
}
