package portfolio

import (
	"testing"
	"time"

	"event-driven-backtesting-engine/internal/domain"
	"event-driven-backtesting-engine/internal/events"
)

func TestPortfolioEngineConsume(t *testing.T) {

	queue := events.NewEventQueue()

	portfolio := domain.NewPortfolio(
		1,
		100000,
	)

	engine := NewEngine(
		queue,
		portfolio,
	)

	trade := domain.NewTrade(
		1,
		1,
		1,
		"BTCUSDT",
		domain.BuyOrder,
		1,
		50000,
		50,
		time.Now(),
	)

	tradeEvent := events.NewTradeExecutedEvent(trade)

	if err := engine.Consume(tradeEvent); err != nil {
		t.Fatal(err)
	}

	// PortfolioUpdatedEvent ต้องถูก Push
	if queue.Len() != 1 {
		t.Fatalf(
			"expected queue length 1, got %d",
			queue.Len(),
		)
	}

	event, ok := queue.Pop()
	if !ok {
		t.Fatal("expected PortfolioUpdatedEvent")
	}

	portfolioEvent, ok := event.(events.PortfolioUpdatedEvent)
	if !ok {
		t.Fatal("expected PortfolioUpdatedEvent")
	}

	// Cash ต้องลดลง
	expectedCash := 100000.0 - (50000.0 * 1.0) - 50.0

	if portfolioEvent.Portfolio.Cash != expectedCash {
		t.Fatalf(
			"expected cash %.2f, got %.2f",
			expectedCash,
			portfolioEvent.Portfolio.Cash,
		)
	}

	// Position ต้องถูกสร้าง
	position, ok := engine.Positions()["BTCUSDT"]
	if !ok {
		t.Fatal("expected BTCUSDT position")
	}

	if position.Quantity != 1 {
		t.Fatalf(
			"expected quantity 1, got %.2f",
			position.Quantity,
		)
	}

	if position.AveragePrice != 50000 {
		t.Fatalf(
			"expected average price 50000, got %.2f",
			position.AveragePrice,
		)
	}

	if position.CurrentPrice != 50000 {
		t.Fatalf(
			"expected market price 50000, got %.2f",
			position.CurrentPrice,
		)
	}

	// Equity ต้องถูกคำนวณใหม่
	expectedEquity :=
		expectedCash +
			position.CurrentValue

	if portfolioEvent.Portfolio.Equity != expectedEquity {
		t.Fatalf(
			"expected equity %.2f, got %.2f",
			expectedEquity,
			portfolioEvent.Portfolio.Equity,
		)
	}
}

func TestPortfolioEngineConsume_InvalidEvent(t *testing.T) {

	queue := events.NewEventQueue()

	portfolio := domain.NewPortfolio(1, 100000)

	engine := NewEngine(queue, portfolio)

	event := events.NewPortfolioUpdatedEvent(portfolio)

	err := engine.Consume(event)

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPortfolioEngine_AddSnapshot(t *testing.T) {

	queue := events.NewEventQueue()

	engine := NewEngine(
		queue,
		domain.NewPortfolio(1, 100000),
	)

	trade := domain.Trade{
		Symbol: "BTCUSDT",
		Side: domain.BuyOrder,

		Quantity: 1,

		ExecutedPrice: 100,

		ExecutedTime: time.Now(),
	}

	event := events.NewTradeExecutedEvent(trade)

	_ = engine.Consume(event)

	if len(engine.Snapshots()) != 1 {
		t.Fatalf(
			"expected snapshot 1, got %d",
			len(engine.Snapshots()),
		)
	}
}

func TestPortfolioEngine_PushPortfolioUpdatedEvent(t *testing.T) {

	queue := events.NewEventQueue()

	engine := NewEngine(
		queue,
		domain.NewPortfolio(1, 100000),
	)

	trade := domain.Trade{
		Symbol:          "BTCUSDT",
		Side:            domain.BuyOrder,
		Quantity:        1,
		ExecutedPrice:   100,
		ExecutedTime:    time.Now(),
	}

	event := events.NewTradeExecutedEvent(trade)

	if err := engine.Consume(event); err != nil {
		t.Fatal(err)
	}

	e, ok := queue.Pop()

	if !ok {
		t.Fatal("expected event in queue")
	}

	_, ok = e.(events.PortfolioUpdatedEvent)

	if !ok {
		t.Fatal("expected PortfolioUpdatedEvent")
	}
}

func TestPortfolioEngine_UpdatePositionValue(t *testing.T) {

	queue := events.NewEventQueue()

	engine := NewEngine(
		queue,
		domain.NewPortfolio(1, 100000),
	)

	trade := domain.Trade{
		Symbol: "BTCUSDT",
		Side: domain.BuyOrder,

		Quantity: 2,

		ExecutedPrice: 100,

		ExecutedTime: time.Now(),
	}

	event := events.NewTradeExecutedEvent(trade)

	_ = engine.Consume(event)

	if engine.Portfolio().PositionValue != 200 {
		t.Fatalf(
			"expected 200 got %.2f",
			engine.Portfolio().PositionValue,
		)
	}
}
