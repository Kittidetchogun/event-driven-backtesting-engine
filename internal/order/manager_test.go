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

	if !orderEvent.Order.CreatedAt.Equal(signal.SignalTime) {
		t.Fatalf(
			"expected Order.CreatedAt=%v, got %v",
			signal.SignalTime,
			orderEvent.Order.CreatedAt,
		)
	}
}

func TestOrderManagerSizingUsesSignalPrice(t *testing.T) {
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

	signalTime := time.Date(
		2026, 8, 12, 0, 0, 0, 0,
		time.UTC,
	)

	// Price available at signal time.
	signalPrice := 100000.0

	signal := events.NewSignalGeneratedEvent(
		1,
		"BTCUSDT",
		domain.BuyOrder,
		1,
		signalPrice,
		signalTime,
	)

	if err := manager.Consume(signal); err != nil {
		t.Fatal(err)
	}

	event, ok := queue.Pop()
	if !ok {
		t.Fatal("expected OrderCreatedEvent")
	}

	orderEvent, ok := event.(events.OrderCreatedEvent)
	if !ok {
		t.Fatalf("expected OrderCreatedEvent, got %T", event)
	}

	expectedQuantity := 0.01

	if orderEvent.Order.Quantity != expectedQuantity {
		t.Fatalf(
			"expected quantity %.8f, got %.8f",
			expectedQuantity,
			orderEvent.Order.Quantity,
		)
	}

	if orderEvent.Order.Price != signalPrice {
		t.Fatalf(
			"expected order price %.2f, got %.2f",
			signalPrice,
			orderEvent.Order.Price,
		)
	}

	if !orderEvent.Order.CreatedAt.Equal(signalTime) {
		t.Fatalf(
			"expected CreatedAt=%v, got %v",
			signalTime,
			orderEvent.Order.CreatedAt,
		)
	}
}

func TestOrderManagerRejectsBuyWhenInsufficientCash(t *testing.T) {
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
		t.Fatalf("expected rejection to be handled without error, got %v", err)
	}

	if queue.Len() != 1 {
		t.Fatalf("expected 1 event, got %d", queue.Len())
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

	if rejectedEvent.Order.Status != domain.RejectedOrder {
		t.Fatalf(
			"expected status REJECTED, got %s",
			rejectedEvent.Order.Status,
		)
	}

	if rejectedEvent.Reason == "" {
		t.Fatal("expected rejection reason")
	}

	if rejectedEvent.Order.RejectReason == "" {
		t.Fatal("expected Order.RejectReason")
	}

	if rejectedEvent.Order.RejectedAt == nil {
		t.Fatal("expected Order.RejectedAt")
	}
}

func TestOrderManagerRejectsSellWhenNoPosition(t *testing.T) {
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

	signalTime := time.Date(
		2026, 8, 12, 0, 0, 0, 0,
		time.UTC,
	)

	signal := events.NewSignalGeneratedEvent(
		1,
		"BTCUSDT",
		domain.SellOrder,
		1,
		100000,
		signalTime,
	)

	if err := manager.Consume(signal); err != nil {
		t.Fatalf("expected rejection to be handled without error, got %v", err)
	}

	if queue.Len() != 1 {
		t.Fatalf("expected 1 event, got %d", queue.Len())
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

	if rejectedEvent.Order.Status != domain.RejectedOrder {
		t.Fatalf(
			"expected status REJECTED, got %s",
			rejectedEvent.Order.Status,
		)
	}

	if rejectedEvent.Reason == "" {
		t.Fatal("expected rejection reason")
	}

	if rejectedEvent.Order.RejectReason == "" {
		t.Fatal("expected Order.RejectReason")
	}

	if rejectedEvent.Order.RejectedAt == nil {
		t.Fatal("expected Order.RejectedAt")
	}
}

func TestOrderManagerDoesNotCreateOrderEventWhenRejected(t *testing.T) {
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

	signal := events.NewSignalGeneratedEvent(
		1,
		"BTCUSDT",
		domain.BuyOrder,
		1,
		1000,
		time.Now(),
	)

	if err := manager.Consume(signal); err != nil {
		t.Fatalf("expected rejection to be handled, got %v", err)
	}

	event, ok := queue.Pop()
	if !ok {
		t.Fatal("expected rejection event")
	}

	if _, ok := event.(events.OrderCreatedEvent); ok {
		t.Fatal("rejected order must not create OrderCreatedEvent")
	}

	if _, ok := event.(events.OrderRejectedEvent); !ok {
		t.Fatalf(
			"expected OrderRejectedEvent, got %T",
			event,
		)
	}
}
