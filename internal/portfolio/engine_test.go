package portfolio

import (
	"errors"
	"fmt"
	"math"
	"testing"
	"time"

	"event-driven-backtesting-engine/internal/domain"
	"event-driven-backtesting-engine/internal/events"
)

func assertPortfolioAccounting(t *testing.T, engine *Engine, cash, quantity, average, value, equity, realized, unrealized float64) {
	t.Helper()
	const tolerance = 1e-9
	portfolio := engine.Portfolio()
	if math.Abs(portfolio.Cash-cash) > tolerance || math.Abs(portfolio.PositionValue-value) > tolerance || math.Abs(portfolio.Equity-equity) > tolerance || math.Abs(portfolio.RealizedPnL-realized) > tolerance || math.Abs(portfolio.UnrealizedPnL-unrealized) > tolerance {
		t.Fatalf("portfolio = cash %.12f value %.12f equity %.12f realized %.12f unrealized %.12f", portfolio.Cash, portfolio.PositionValue, portfolio.Equity, portfolio.RealizedPnL, portfolio.UnrealizedPnL)
	}
	position, ok := engine.Positions()["BTCUSDT"]
	if quantity == 0 {
		if ok {
			t.Fatalf("expected no open position, got %+v", position)
		}
		return
	}
	if !ok || math.Abs(position.Quantity-quantity) > tolerance || math.Abs(position.AveragePrice-average) > tolerance || math.Abs(position.CurrentValue-value) > tolerance {
		t.Fatalf("position = %+v, want quantity %.12f average %.12f value %.12f", position, quantity, average, value)
	}
}

func reserveTestTrade(t *testing.T, engine *Engine, trade domain.Trade) {
	t.Helper()
	amount := 0.0
	if trade.Side == domain.BuyOrder {
		amount = trade.ExecutedPrice*trade.Quantity + trade.TransactionCost
	}
	if err := engine.Reserve(domain.OrderID(trade.OrderID), amount); err != nil {
		t.Fatalf("Reserve() error = %v", err)
	}
}

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
	reserveTestTrade(t, engine, trade)

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
		OrderID: 2,
		Symbol:  "BTCUSDT",
		Side:    domain.BuyOrder,

		Quantity: 1,

		ExecutedPrice: 100,

		ExecutedTime: time.Now(),
	}

	event := events.NewTradeExecutedEvent(trade)
	reserveTestTrade(t, engine, trade)

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
	engine.recordSnapshot(first)
	engine.portfolio.Cash = 200
	engine.recordSnapshot(second)
	engine.portfolio.Cash = 300
	engine.recordSnapshot(first)

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
		OrderID:       3,
		Symbol:        "BTCUSDT",
		Side:          domain.BuyOrder,
		Quantity:      1,
		ExecutedPrice: 100,
		ExecutedTime:  time.Now(),
	}

	event := events.NewTradeExecutedEvent(trade)
	reserveTestTrade(t, engine, trade)

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
		OrderID: 4,
		Symbol:  "BTCUSDT",
		Side:    domain.BuyOrder,

		Quantity: 2,

		ExecutedPrice: 100,

		ExecutedTime: time.Now(),
	}

	event := events.NewTradeExecutedEvent(trade)
	reserveTestTrade(t, engine, trade)

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

	if err := engine.UpdateMarketPrices(signalCandle); err != nil {
		t.Fatal(err)
	}

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
	reserveTestTrade(t, engine, trade)

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

func TestPortfolioEngineReservationUsesAvailableCash(t *testing.T) {
	engine := NewEngine(events.NewEventQueue(), domain.NewPortfolio(1, 1000))

	if err := engine.Reserve(1, 700); err != nil {
		t.Fatal(err)
	}
	if got := engine.AvailableCash(); got != 300 {
		t.Fatalf("AvailableCash() = %.2f, want 300", got)
	}
	if err := engine.Reserve(2, 301); err == nil {
		t.Fatal("expected second reservation to fail against available cash")
	}
	if got := engine.ReservedCash(); got != 700 {
		t.Fatalf("ReservedCash() = %.2f, want 700", got)
	}
}

func TestPortfolioEngineReservationIncludesFee(t *testing.T) {
	engine := NewEngine(events.NewEventQueue(), domain.NewPortfolio(1, 100))
	if err := engine.Reserve(1, 100.01); err == nil {
		t.Fatal("expected reservation including fee to fail")
	}
	if err := engine.Reserve(1, 100); err != nil {
		t.Fatal(err)
	}
}

func TestPortfolioEngineReservationFloatingPointNeverBecomesNegative(t *testing.T) {
	engine := NewEngine(events.NewEventQueue(), domain.NewPortfolio(1, 1))
	if err := engine.Reserve(1, 0.1); err != nil {
		t.Fatal(err)
	}
	if err := engine.Reserve(2, 0.2); err != nil {
		t.Fatal(err)
	}
	if err := engine.Release(1); err != nil {
		t.Fatal(err)
	}
	if err := engine.Release(2); err != nil {
		t.Fatal(err)
	}
	if got := engine.ReservedCash(); got < 0 || got != 0 {
		t.Fatalf("ReservedCash() = %.18f, want 0", got)
	}
}

func TestPortfolioEngineSettlementReleasesReservationOnce(t *testing.T) {
	engine := NewEngine(events.NewEventQueue(), domain.NewPortfolio(1, 1000))
	trade := domain.NewTrade(1, 1, 7, "BTCUSDT", domain.BuyOrder, 1, 120, 0.12, time.Unix(2, 0))
	if err := engine.Reserve(domain.OrderID(trade.OrderID), 120.12); err != nil {
		t.Fatal(err)
	}

	if err := engine.Consume(events.NewTradeExecutedEvent(trade)); err != nil {
		t.Fatal(err)
	}
	if got := engine.Portfolio().Cash; got != 879.88 {
		t.Fatalf("Cash = %.2f, want 879.88", got)
	}
	if got := engine.ReservedCash(); got != 0 {
		t.Fatalf("ReservedCash() = %.2f, want 0", got)
	}
	if err := engine.Consume(events.NewTradeExecutedEvent(trade)); err == nil {
		t.Fatal("expected duplicate settlement to fail")
	}
	if got := engine.Portfolio().Cash; got != 879.88 {
		t.Fatalf("duplicate settlement changed Cash to %.2f", got)
	}
}

func TestPortfolioEngineSettlementMismatchDoesNotChangeState(t *testing.T) {
	engine := NewEngine(events.NewEventQueue(), domain.NewPortfolio(1, 1000))
	trade := domain.NewTrade(1, 1, 8, "BTCUSDT", domain.BuyOrder, 1, 120, 0.12, time.Unix(2, 0))
	if err := engine.Reserve(domain.OrderID(trade.OrderID), 100); err != nil {
		t.Fatal(err)
	}

	if err := engine.Consume(events.NewTradeExecutedEvent(trade)); err == nil {
		t.Fatal("expected settlement mismatch error")
	}
	if got := engine.Portfolio().Cash; got != 1000 {
		t.Fatalf("Cash = %.2f, want 1000 after failed settlement", got)
	}
	if len(engine.Positions()) != 0 {
		t.Fatal("failed settlement must not create a position")
	}
	if got := engine.ReservedCash(); got != 100 {
		t.Fatalf("reservation must remain after failed settlement, got %.2f", got)
	}
}

func TestPortfolioAccountingBuyWithFee(t *testing.T) {
	engine := NewEngine(events.NewEventQueue(), domain.NewPortfolio(1, 1000))
	trade := domain.NewTrade(1, 1, 1, "BTCUSDT", domain.BuyOrder, 1, 100, 0.10, time.Unix(1, 0))
	reserveTestTrade(t, engine, trade)
	if err := engine.Consume(events.NewTradeExecutedEvent(trade)); err != nil {
		t.Fatal(err)
	}

	assertPortfolioAccounting(t, engine, 899.90, 1, 100.10, 100, 999.90, 0, -0.10)
}

func TestPortfolioAccountingBuyThenFullSell(t *testing.T) {
	engine := NewEngine(events.NewEventQueue(), domain.NewPortfolio(1, 1000))
	trades := []domain.Trade{
		domain.NewTrade(1, 1, 1, "BTCUSDT", domain.BuyOrder, 1, 100, 0.10, time.Unix(1, 0)),
		domain.NewTrade(2, 1, 2, "BTCUSDT", domain.SellOrder, 1, 120, 0.12, time.Unix(2, 0)),
	}
	for _, trade := range trades {
		reserveTestTrade(t, engine, trade)
		if err := engine.Consume(events.NewTradeExecutedEvent(trade)); err != nil {
			t.Fatal(err)
		}
	}

	assertPortfolioAccounting(t, engine, 1019.78, 0, 0, 0, 1019.78, 19.78, 0)
}

func TestPortfolioAccountingMultipleBuysPartialAndFullClose(t *testing.T) {
	engine := NewEngine(events.NewEventQueue(), domain.NewPortfolio(1, 1000))
	trades := []domain.Trade{
		domain.NewTrade(1, 1, 1, "BTCUSDT", domain.BuyOrder, 1, 100, 0.10, time.Unix(1, 0)),
		domain.NewTrade(2, 1, 2, "BTCUSDT", domain.BuyOrder, 1, 120, 0.12, time.Unix(2, 0)),
		domain.NewTrade(3, 1, 3, "BTCUSDT", domain.SellOrder, 1, 130, 0.13, time.Unix(3, 0)),
	}
	for _, trade := range trades {
		reserveTestTrade(t, engine, trade)
		if err := engine.Consume(events.NewTradeExecutedEvent(trade)); err != nil {
			t.Fatal(err)
		}
	}

	assertPortfolioAccounting(t, engine, 909.65, 1, 110.11, 130, 1039.65, 19.76, 19.89)

	trade := domain.NewTrade(4, 1, 4, "BTCUSDT", domain.SellOrder, 1, 140, 0.14, time.Unix(4, 0))
	reserveTestTrade(t, engine, trade)
	if err := engine.Consume(events.NewTradeExecutedEvent(trade)); err != nil {
		t.Fatal(err)
	}
	assertPortfolioAccounting(t, engine, 1049.51, 0, 0, 0, 1049.51, 49.51, 0)
}

func TestPortfolioSettlementRejectsInvalidSellAtBoundary(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*testing.T, *Engine)
		quantity float64
		wantErr  error
		orderID  int
	}{
		{name: "no position", quantity: 1, wantErr: ErrNoPosition, orderID: 1},
		{name: "greater than position", quantity: 2, wantErr: ErrInsufficientPosition, orderID: 2, setup: func(t *testing.T, engine *Engine) {
			trade := domain.NewTrade(1, 1, 10, "BTCUSDT", domain.BuyOrder, 1, 100, 0.10, time.Unix(1, 0))
			reserveTestTrade(t, engine, trade)
			if err := engine.Consume(events.NewTradeExecutedEvent(trade)); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "zero quantity", quantity: 0, wantErr: ErrSettlementMismatch, orderID: 3},
		{name: "negative quantity", quantity: -1, wantErr: ErrSettlementMismatch, orderID: 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine(events.NewEventQueue(), domain.NewPortfolio(1, 1000))
			if tt.setup != nil {
				tt.setup(t, engine)
			}
			trade := domain.NewTrade(tt.orderID, 1, tt.orderID, "BTCUSDT", domain.SellOrder, tt.quantity, 120, 0.12, time.Unix(2, 0))
			reserveTestTrade(t, engine, trade)
			before := engine.Portfolio()
			beforePositions := len(engine.Positions())
			if err := engine.Consume(events.NewTradeExecutedEvent(trade)); !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want errors.Is(..., %v)", err, tt.wantErr)
			}
			after := engine.Portfolio()
			if after.Cash != before.Cash || after.RealizedPnL != before.RealizedPnL || len(engine.Positions()) != beforePositions {
				t.Fatal("invalid sell changed portfolio accounting")
			}
		})
	}
}

func TestPortfolioEngineMarkToMarket(t *testing.T) {
	queue := events.NewEventQueue()
	engine := NewEngine(queue, domain.NewPortfolio(1, 1000))
	trade := domain.NewTrade(1, 1, 1, "BTCUSDT", domain.BuyOrder, 1, 100, 0.10, time.Unix(1, 0))
	reserveTestTrade(t, engine, trade)
	if err := engine.Consume(events.NewTradeExecutedEvent(trade)); err != nil {
		t.Fatal(err)
	}
	for !queue.IsEmpty() {
		queue.Pop()
	}

	if err := engine.UpdateMarketPrices(domain.Candle{Timestamp: time.Unix(2, 0), Symbol: "BTCUSDT", Close: 120}); err != nil {
		t.Fatal(err)
	}
	assertPortfolioAccounting(t, engine, 899.90, 1, 100.10, 120, 1019.90, 0, 19.90)

	if err := engine.UpdateMarketPrices(domain.Candle{Timestamp: time.Unix(3, 0), Symbol: "BTCUSDT", Close: 90}); err != nil {
		t.Fatal(err)
	}
	assertPortfolioAccounting(t, engine, 899.90, 1, 100.10, 90, 989.90, 0, -10.10)
}

func TestPortfolioEngineRejectsInvalidMarketCloseWithoutMutation(t *testing.T) {
	invalidCloses := []float64{0, -1, math.NaN(), math.Inf(1), math.Inf(-1)}
	for _, close := range invalidCloses {
		t.Run(fmt.Sprintf("close_%v", close), func(t *testing.T) {
			queue := events.NewEventQueue()
			engine := NewEngine(queue, domain.NewPortfolio(1, 1000))
			trade := domain.NewTrade(1, 1, 1, "BTCUSDT", domain.BuyOrder, 1, 100, 0.10, time.Unix(1, 0))
			reserveTestTrade(t, engine, trade)
			if err := engine.Consume(events.NewTradeExecutedEvent(trade)); err != nil {
				t.Fatal(err)
			}
			for !queue.IsEmpty() {
				queue.Pop()
			}

			beforePortfolio := engine.Portfolio()
			beforePosition := engine.Positions()["BTCUSDT"]
			beforeSnapshots := len(engine.Snapshots())
			if err := engine.UpdateMarketPrices(domain.Candle{Timestamp: time.Unix(2, 0), Symbol: "BTCUSDT", Close: close}); !errors.Is(err, ErrInvalidMarketPrice) {
				t.Fatalf("error = %v, want ErrInvalidMarketPrice", err)
			}

			afterPortfolio := engine.Portfolio()
			afterPosition := engine.Positions()["BTCUSDT"]
			if afterPortfolio != beforePortfolio {
				t.Fatalf("portfolio mutated: before=%+v after=%+v", beforePortfolio, afterPortfolio)
			}
			if afterPosition != beforePosition {
				t.Fatalf("position mutated: before=%+v after=%+v", beforePosition, afterPosition)
			}
			if len(engine.Snapshots()) != beforeSnapshots || !queue.IsEmpty() {
				t.Fatal("invalid market price created a snapshot/event")
			}
		})
	}
}

func TestPortfolioEngineSameTimestampUsesCloseValuation(t *testing.T) {
	timestamp := time.Unix(10, 0)
	queue := events.NewEventQueue()
	engine := NewEngine(queue, domain.NewPortfolio(1, 1000))
	trade := domain.NewTrade(1, 1, 1, "BTCUSDT", domain.BuyOrder, 1, 100, 0.10, timestamp)
	reserveTestTrade(t, engine, trade)
	if err := engine.Consume(events.NewTradeExecutedEvent(trade)); err != nil {
		t.Fatal(err)
	}
	for !queue.IsEmpty() {
		queue.Pop()
	}

	if err := engine.UpdateMarketPrices(domain.Candle{Timestamp: timestamp, Symbol: "BTCUSDT", Close: 120}); err != nil {
		t.Fatal(err)
	}
	if len(engine.Snapshots()) != 1 {
		t.Fatalf("snapshot count = %d, want 1 after same-timestamp replacement", len(engine.Snapshots()))
	}
	snapshot := engine.Snapshots()[0]
	if math.Abs(snapshot.Equity-1019.90) > 1e-9 || math.Abs(snapshot.PositionValue-120) > 1e-9 || math.Abs(snapshot.UnrealizedPnL-19.90) > 1e-9 {
		t.Fatalf("snapshot = %+v, want close valuation equity=1019.90 value=120 unrealized=19.90", snapshot)
	}
}
