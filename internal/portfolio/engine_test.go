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

	expectedCost := (50000.0 * 1.0) + 50.0

	if position.AveragePrice != expectedCost {
		t.Fatalf(
			"expected average price %.2f, got %.2f",
			expectedCost,
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

func TestPortfolioEngine_RecordSnapshot_ReplacesDuplicateAndPreservesOrder(t *testing.T) {
	engine := NewEngine(
		events.NewEventQueue(),
		domain.NewPortfolio(1, 100000),
	)

	first := time.Unix(1000, 0).UTC()
	second := first.Add(time.Minute)

	engine.portfolio.Cash = 100
	engine.RecordSnapshot(first)
	engine.portfolio.Cash = 200
	engine.RecordSnapshot(second)
	engine.portfolio.Cash = 300
	engine.RecordSnapshot(first)

	snapshots := engine.Snapshots()
	if len(snapshots) != 2 {
		t.Fatalf("expected 2 snapshots, got %d", len(snapshots))
	}
	if !snapshots[0].Time.Equal(first) || snapshots[0].Equity != 300 {
		t.Fatalf("duplicate snapshot was not replaced with latest values: %+v", snapshots[0])
	}
	if !snapshots[1].Time.Equal(second) || snapshots[1].Equity != 200 {
		t.Fatalf("unique snapshot order or values changed: %+v", snapshots[1])
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

func TestPortfolioEngineCanBuy(t *testing.T) {
	p := domain.NewPortfolio(1, 1000)

	engine := NewEngine(
		events.NewEventQueue(),
		p,
	)

	order := domain.NewOrder(
		1,
		"BTCUSDT",
		domain.BuyOrder,
		0.01,
		50000,
		time.Now(),
	)

	err := engine.CanBuy(order)

	if err != nil {
		t.Fatalf("expected CanBuy to succeed, got %v", err)
	}
}

func TestPortfolioEngineCanBuy_InsufficientCash(t *testing.T) {
	p := domain.NewPortfolio(1, 1000)

	engine := NewEngine(
		events.NewEventQueue(),
		p,
	)

	order := domain.NewOrder(
		1,
		"BTCUSDT",
		domain.BuyOrder,
		1,
		2000,
		time.Now(),
	)

	err := engine.CanBuy(order)

	if err == nil {
		t.Fatal("expected CanBuy to fail due to insufficient cash")
	}
}

func TestPortfolioEngineCanBuy_MarketOrder(t *testing.T) {
	p := domain.NewPortfolio(1, 1000)

	engine := NewEngine(
		events.NewEventQueue(),
		p,
	)

	// Market order ยังไม่มี execution price
	order := domain.NewOrder(
		1,
		"BTCUSDT",
		domain.BuyOrder,
		1,
		0,
		time.Now(),
	)

	err := engine.CanBuy(order)

	if err != nil {
		t.Fatalf("expected market order to pass, got %v", err)
	}
}

func TestPortfolioEngineCanSell(t *testing.T) {
	p := domain.NewPortfolio(1, 1000)

	engine := NewEngine(
		events.NewEventQueue(),
		p,
	)

	// สร้าง position ที่มีอยู่ใน portfolio engine
	engine.positions["BTCUSDT"] = domain.Position{
		Symbol:   "BTCUSDT",
		Side:     domain.BuyOrder,
		Quantity: 1,
	}

	order := domain.NewOrder(
		1,
		"BTCUSDT",
		domain.SellOrder,
		0.5,
		0,
		time.Now(),
	)

	err := engine.CanSell(order)

	if err != nil {
		t.Fatalf("expected CanSell to succeed, got %v", err)
	}
}

func TestPortfolioEngineCanSell_NoPosition(t *testing.T) {
	p := domain.NewPortfolio(1, 1000)

	engine := NewEngine(
		events.NewEventQueue(),
		p,
	)

	order := domain.NewOrder(
		1,
		"BTCUSDT",
		domain.SellOrder,
		1,
		0,
		time.Now(),
	)

	err := engine.CanSell(order)

	if err == nil {
		t.Fatal("expected CanSell to fail when no position exists")
	}
}

func TestPortfolioEngineCanSell_InsufficientPosition(t *testing.T) {
	p := domain.NewPortfolio(1, 1000)

	engine := NewEngine(
		events.NewEventQueue(),
		p,
	)

	engine.positions["BTCUSDT"] = domain.Position{
		Symbol:   "BTCUSDT",
		Side:     domain.BuyOrder,
		Quantity: 0.5,
	}

	order := domain.NewOrder(
		1,
		"BTCUSDT",
		domain.SellOrder,
		1,
		0,
		time.Now(),
	)

	err := engine.CanSell(order)

	if err == nil {
		t.Fatal("expected CanSell to fail due to insufficient position")
	}
}

func TestPortfolioEngineDoesNotCreatePositionBeforeTradeExecution(t *testing.T) {
	queue := events.NewEventQueue()

	signalTime := time.Date(
		2026, 7, 16,
		0, 0, 0, 0,
		time.UTC,
	)

	executionTime := time.Date(
		2026, 7, 17,
		0, 0, 0, 0,
		time.UTC,
	)

	engine := NewEngine(
		queue,
		domain.NewPortfolio(1, 100000),
	)

	signalCandle := domain.Candle{
		Timestamp: signalTime,
		Symbol:    "BTCUSDT",
		Timeframe: "1d",
		Open:      100,
		High:      110,
		Low:       95,
		Close:     105,
		Volume:    1000,
	}

	// -------------------------------------------------
	// Day 16
	// Portfolio sees the signal candle.
	// There is no executed trade yet.
	// Therefore no position must exist.
	// -------------------------------------------------

	engine.UpdateMarketPrices(signalCandle)

	if _, ok := engine.Positions()["BTCUSDT"]; ok {
		t.Fatal("position must not exist before trade execution")
	}

	if engine.Portfolio().PositionValue != 0 {
		t.Fatalf(
			"expected position value 0 before trade execution, got %.2f",
			engine.Portfolio().PositionValue,
		)
	}

	// -------------------------------------------------
	// Day 17
	// Trade is actually executed.
	// Only now should the portfolio create the position.
	// -------------------------------------------------

	trade := domain.NewTrade(
		1,
		1,
		1,
		"BTCUSDT",
		domain.BuyOrder,
		1,
		120,
		0,
		executionTime,
	)

	tradeEvent := events.NewTradeExecutedEvent(trade)

	if err := engine.Consume(tradeEvent); err != nil {
		t.Fatal(err)
	}

	position, ok := engine.Positions()["BTCUSDT"]
	if !ok {
		t.Fatal("expected position after trade execution")
	}

	if position.Quantity != 1 {
		t.Fatalf(
			"expected position quantity 1, got %.2f",
			position.Quantity,
		)
	}

	if position.AveragePrice != 120 {
		t.Fatalf(
			"expected average price 120, got %.2f",
			position.AveragePrice,
		)
	}

	if engine.Portfolio().PositionValue != 120 {
		t.Fatalf(
			"expected position value 120, got %.2f",
			engine.Portfolio().PositionValue,
		)
	}
}
