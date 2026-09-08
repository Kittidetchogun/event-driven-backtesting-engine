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

func TestOrderManagerEmitsExplicitOrderRejectedEvent(t *testing.T) {
	queue := events.NewEventQueue()

	initialPortfolio := domain.NewPortfolio(1, 100)
	initialPortfolio.Cash = 5

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

	signalTime := time.Date(
		2026, 8, 12, 0, 0, 0, 0,
		time.UTC,
	)

	signal := events.NewSignalGeneratedEvent(
		1,
		"BTCUSDT",
		domain.BuyOrder,
		1,
		1000,
		signalTime,
	)

	if err := manager.Consume(signal); err != nil {
		t.Fatalf(
			"expected rejection to be handled without error, got %v",
			err,
		)
	}

	event, ok := queue.Pop()
	if !ok {
		t.Fatal("expected OrderRejectedEvent")
	}

	rejectedEvent, ok := event.(events.OrderRejectedEvent)
	if !ok {
		t.Fatalf(
			"expected OrderRejectedEvent, got %T",
			event,
		)
	}

	if rejectedEvent.Type() != events.OrderRejectedEventType {
		t.Fatalf(
			"expected event type %s, got %s",
			events.OrderRejectedEventType,
			rejectedEvent.Type(),
		)
	}

	if rejectedEvent.Order.Status != domain.RejectedOrder {
		t.Fatalf(
			"expected order status REJECTED, got %s",
			rejectedEvent.Order.Status,
		)
	}

	if rejectedEvent.Reason == "" {
		t.Fatal("expected explicit rejection reason")
	}

	if rejectedEvent.Order.RejectReason == "" {
		t.Fatal("expected Order.RejectReason")
	}

	if rejectedEvent.Order.RejectedAt == nil {
		t.Fatal("expected Order.RejectedAt")
	}

	if !rejectedEvent.Order.RejectedAt.Equal(signalTime) {
		t.Fatalf(
			"expected RejectedAt=%v, got %v",
			signalTime,
			*rejectedEvent.Order.RejectedAt,
		)
	}

	// Display rejection result clearly in test output.
	t.Log("========================================")
	t.Log("ORDER REJECTION RESULT")
	t.Log("========================================")
	t.Logf("Event Type       : %s", rejectedEvent.Type())
	t.Logf("Symbol           : %s", rejectedEvent.Order.Symbol)
	t.Logf("Side             : %s", rejectedEvent.Order.Side)
	t.Logf("Order Status     : %s", rejectedEvent.Order.Status)
	t.Logf("Order Quantity   : %.8f", rejectedEvent.Order.Quantity)
	t.Logf("Order Price      : %.2f", rejectedEvent.Order.Price)
	t.Logf("Reason           : %s", rejectedEvent.Reason)
	t.Logf("Order RejectReason: %s", rejectedEvent.Order.RejectReason)
	t.Logf("Rejected At      : %v", *rejectedEvent.Order.RejectedAt)
	t.Log("========================================")
}
