package order

import (
	"testing"
	"time"

	"event-driven-backtesting-engine/internal/domain"
	"event-driven-backtesting-engine/internal/events"
	"event-driven-backtesting-engine/internal/portfolio"
	"event-driven-backtesting-engine/internal/sizing"
)

func TestOrderManagerConsume(t *testing.T) {
	queue := events.NewEventQueue()

	initialPortfolio := domain.NewPortfolio(1, 10000)

	portfolioEngine := portfolio.NewEngine(
		queue,
		initialPortfolio,
	)

	sizer := sizing.NewFixedFractional(0.10)

	manager := NewManager(
		portfolioEngine,
		queue,
		sizer,
	)

	signal := events.NewSignalGeneratedEvent(
		1,               // RunID
		"BTCUSDT",       // Symbol
		domain.BuyOrder, // SignalType
		1,               // Quantity
		100000,          // Price
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

	if orderEvent.Order.Quantity != 0.01 {
		t.Fatalf(
			"expected quantity 0.01 got %.8f",
			orderEvent.Order.Quantity,
		)
	}
}

func TestOrderManagerConsumeSell_UsesAvailablePosition(t *testing.T) {
	queue := events.NewEventQueue()

	initialPortfolio := domain.NewPortfolio(1, 10000)

	portfolioEngine := portfolio.NewEngine(
		queue,
		initialPortfolio,
	)

	portfolioEngine.Positions()["BTCUSDT"] = domain.Position{
		Symbol:   "BTCUSDT",
		Side:     domain.BuyOrder,
		Quantity: 0.5,
	}

	sizer := sizing.NewFixedFractional(0.10)

	manager := NewManager(
		portfolioEngine,
		queue,
		sizer,
	)

	signal := events.NewSignalGeneratedEvent(
		1,
		"BTCUSDT",
		domain.SellOrder,
		1,
		1000,
		time.Now(),
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
		t.Fatalf("expected OrderCreatedEvent got %T", event)
	}

	if orderEvent.Order.Side != domain.SellOrder {
		t.Fatalf("expected side SELL got %s", orderEvent.Order.Side)
	}

	if orderEvent.Order.Quantity != 0.5 {
		t.Fatalf("expected quantity 0.5 got %.8f", orderEvent.Order.Quantity)
	}
}

func TestOrderManagerConsumeSell_NoPositionIgnored(t *testing.T) {
	queue := events.NewEventQueue()

	initialPortfolio := domain.NewPortfolio(1, 10000)

	portfolioEngine := portfolio.NewEngine(
		queue,
		initialPortfolio,
	)

	sizer := sizing.NewFixedFractional(0.10)

	manager := NewManager(
		portfolioEngine,
		queue,
		sizer,
	)

	signal := events.NewSignalGeneratedEvent(
		1,
		"BTCUSDT",
		domain.SellOrder,
		1,
		1000,
		time.Now(),
	)

	if err := manager.Consume(signal); err != nil {
		t.Fatal(err)
	}

	if queue.Len() != 0 {
		t.Fatalf("expected queue length 0 got %d", queue.Len())
	}
}
