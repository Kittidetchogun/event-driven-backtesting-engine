package matching

import (
	"errors"
	"math"
	"strings"
	"testing"
	"time"

	"event-driven-backtesting-engine/internal/domain"
	"event-driven-backtesting-engine/internal/events"
	"event-driven-backtesting-engine/internal/portfolio"
)

type testCashReserver struct {
	reservations map[domain.OrderID]float64
}

func newTestCashReserver() *testCashReserver {
	return &testCashReserver{reservations: make(map[domain.OrderID]float64)}
}

func mustNewEngine(t *testing.T, queue *events.EventQueue, reserver CashReserver) *Engine {
	t.Helper()
	engine, err := NewEngine(queue, reserver)
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}
	return engine
}

func TestNewEngineRequiresCashReserver(t *testing.T) {
	engine, err := NewEngine(events.NewEventQueue(), nil)
	if !errors.Is(err, ErrCashReserverRequired) {
		t.Fatalf("NewEngine() error = %v, want ErrCashReserverRequired", err)
	}
	if engine != nil {
		t.Fatal("expected nil engine when CashReserver is missing")
	}
}

func (r *testCashReserver) Reserve(orderID domain.OrderID, amount float64) error {
	if _, ok := r.reservations[orderID]; ok {
		return errors.New("duplicate reservation")
	}
	r.reservations[orderID] = amount
	return nil
}

func (r *testCashReserver) Release(orderID domain.OrderID) error {
	if _, ok := r.reservations[orderID]; !ok {
		return errors.New("reservation not found")
	}
	delete(r.reservations, orderID)
	return nil
}

func TestMatchingEngineConsume(t *testing.T) {
	queue := events.NewEventQueue()
	engine := mustNewEngine(t, queue, newTestCashReserver())

	signalTime1 := time.Date(
		2026, 7, 16,
		0, 0, 0, 0,
		time.UTC,
	)

	order1 := domain.NewOrder(
		1,
		"BTCUSDT",
		domain.BuyOrder,
		1,
		50000,
		signalTime1,
	)
	order1.ID = 41

	orderEvent1 := events.NewOrderCreatedEvent(order1)

	// Consume should only put the order into the pending state.
	if err := engine.Consume(orderEvent1); err != nil {
		t.Fatal(err)
	}

	// The order must NOT execute immediately.
	if queue.Len() != 0 {
		t.Fatalf("expected queue length 0 after Consume, got %d", queue.Len())
	}

	if len(engine.Trades()) != 0 {
		t.Fatalf("expected 0 trades after Consume, got %d", len(engine.Trades()))
	}

	// -------------------------------------------------
	// Execute pending order on the next candle
	// -------------------------------------------------

	nextCandle1 := domain.Candle{
		Timestamp: time.Date(
			2026, 7, 17,
			0, 0, 0, 0,
			time.UTC,
		),
		Symbol:    "BTCUSDT",
		Timeframe: "1d",
		Open:      51000,
		High:      52000,
		Low:       50000,
		Close:     51500,
		Volume:    1000,
	}

	if err := engine.ExecutePending(nextCandle1); err != nil {
		t.Fatal(err)
	}

	if queue.Len() != 1 {
		t.Fatalf("expected queue length 1 after execution, got %d", queue.Len())
	}

	event, ok := queue.Pop()
	if !ok {
		t.Fatal("expected TradeExecutedEvent")
	}

	tradeEvent, ok := event.(events.TradeExecutedEvent)
	if !ok {
		t.Fatalf("expected TradeExecutedEvent, got %T", event)
	}

	trade := tradeEvent.Trade

	if trade.TradeID != 1 {
		t.Fatalf("expected trade ID 1, got %d", trade.TradeID)
	}

	if trade.OrderID != int(order1.ID) {
		t.Fatalf("expected trade order ID %d, got %d", order1.ID, trade.OrderID)
	}

	if trade.Symbol != order1.Symbol {
		t.Fatalf(
			"expected symbol %s got %s",
			order1.Symbol,
			trade.Symbol,
		)
	}

	if trade.Side != order1.Side {
		t.Fatalf(
			"expected side %s got %s",
			order1.Side,
			trade.Side,
		)
	}

	if trade.Quantity != order1.Quantity {
		t.Fatalf(
			"expected quantity %.2f got %.2f",
			order1.Quantity,
			trade.Quantity,
		)
	}

	// Execution must use the next candle Open,
	// not the signal candle Close.
	if trade.ExecutedPrice != nextCandle1.Open {
		t.Fatalf(
			"expected executed price %.2f got %.2f",
			nextCandle1.Open,
			trade.ExecutedPrice,
		)
	}

	if !trade.ExecutedTime.Equal(nextCandle1.Timestamp) {
		t.Fatalf(
			"expected executed time %v got %v",
			nextCandle1.Timestamp,
			trade.ExecutedTime,
		)
	}

	if len(engine.Trades()) != 1 {
		t.Fatalf("expected 1 trade, got %d", len(engine.Trades()))
	}

	// -------------------------------------------------
	// Second order -> should generate TradeID = 2
	// -------------------------------------------------

	signalTime2 := time.Date(
		2026, 7, 18,
		0, 0, 0, 0,
		time.UTC,
	)

	order2 := domain.NewOrder(
		2,
		"BTCUSDT",
		domain.SellOrder,
		1,
		51500,
		signalTime2,
	)

	orderEvent2 := events.NewOrderCreatedEvent(order2)

	if err := engine.Consume(orderEvent2); err != nil {
		t.Fatal(err)
	}

	// Second order is pending, so it must not execute yet.
	if queue.Len() != 0 {
		t.Fatalf(
			"expected queue length 0 after second Consume, got %d",
			queue.Len(),
		)
	}

	if len(engine.Trades()) != 1 {
		t.Fatalf(
			"expected 1 trade before second execution, got %d",
			len(engine.Trades()),
		)
	}

	nextCandle2 := domain.Candle{
		Timestamp: time.Date(
			2026, 7, 19,
			0, 0, 0, 0,
			time.UTC,
		),
		Symbol:    "BTCUSDT",
		Timeframe: "1d",
		Open:      52000,
		High:      53000,
		Low:       51000,
		Close:     52500,
		Volume:    1000,
	}

	if err := engine.ExecutePending(nextCandle2); err != nil {
		t.Fatal(err)
	}

	if queue.Len() != 1 {
		t.Fatalf(
			"expected queue length 1 after second execution, got %d",
			queue.Len(),
		)
	}

	event, ok = queue.Pop()
	if !ok {
		t.Fatal("expected second TradeExecutedEvent")
	}

	tradeEvent, ok = event.(events.TradeExecutedEvent)
	if !ok {
		t.Fatalf(
			"expected second TradeExecutedEvent, got %T",
			event,
		)
	}

	trade2 := tradeEvent.Trade

	if trade2.TradeID != 2 {
		t.Fatalf(
			"expected trade ID 2, got %d",
			trade2.TradeID,
		)
	}

	if trade2.ExecutedPrice != nextCandle2.Open {
		t.Fatalf(
			"expected second executed price %.2f got %.2f",
			nextCandle2.Open,
			trade2.ExecutedPrice,
		)
	}

	if !trade2.ExecutedTime.Equal(nextCandle2.Timestamp) {
		t.Fatalf(
			"expected second executed time %v got %v",
			nextCandle2.Timestamp,
			trade2.ExecutedTime,
		)
	}

	if len(engine.Trades()) != 2 {
		t.Fatalf(
			"expected 2 trades, got %d",
			len(engine.Trades()),
		)
	}

	// Verify stored trades
	storedTrades := engine.Trades()

	if storedTrades[0].TradeID != 1 {
		t.Fatalf(
			"expected first trade ID 1, got %d",
			storedTrades[0].TradeID,
		)
	}

	if storedTrades[1].TradeID != 2 {
		t.Fatalf(
			"expected second trade ID 2, got %d",
			storedTrades[1].TradeID,
		)
	}
}

func TestMatchingEngineExecutesOrderOnNextCandleOpen(t *testing.T) {
	queue := events.NewEventQueue()
	engine := mustNewEngine(t, queue, newTestCashReserver())

	signalTime := time.Date(
		2026, 7, 16,
		0, 0, 0, 0,
		time.UTC,
	)

	nextCandleTime := time.Date(
		2026, 7, 17,
		0, 0, 0, 0,
		time.UTC,
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

	nextCandle := domain.Candle{
		Timestamp: nextCandleTime,
		Symbol:    "BTCUSDT",
		Timeframe: "1d",
		Open:      120,
		High:      130,
		Low:       115,
		Close:     125,
		Volume:    1000,
	}

	order := domain.NewOrder(
		1,
		"BTCUSDT",
		domain.BuyOrder,
		1,
		signalCandle.Close,
		signalCandle.Timestamp,
	)

	orderEvent := events.NewOrderCreatedEvent(order)

	// Order is created from the signal on Day 16.
	if err := engine.Consume(orderEvent); err != nil {
		t.Fatal(err)
	}

	// -------------------------------------------------
	// Day 16
	// The order must NOT execute on the same candle.
	// -------------------------------------------------

	if err := engine.ExecutePending(signalCandle); err != nil {
		t.Fatal(err)
	}

	if len(engine.Trades()) != 0 {
		t.Fatalf(
			"expected no trade on signal candle, got %d",
			len(engine.Trades()),
		)
	}

	if queue.Len() != 0 {
		t.Fatalf(
			"expected no TradeExecutedEvent on signal candle, got %d",
			queue.Len(),
		)
	}

	// -------------------------------------------------
	// Day 17
	// The order must execute at Day 17 Open.
	// -------------------------------------------------

	if err := engine.ExecutePending(nextCandle); err != nil {
		t.Fatal(err)
	}

	if len(engine.Trades()) != 1 {
		t.Fatalf(
			"expected 1 trade on next candle, got %d",
			len(engine.Trades()),
		)
	}

	if queue.Len() != 1 {
		t.Fatalf(
			"expected 1 TradeExecutedEvent, got %d",
			queue.Len(),
		)
	}

	event, ok := queue.Pop()
	if !ok {
		t.Fatal("expected TradeExecutedEvent")
	}

	tradeEvent, ok := event.(events.TradeExecutedEvent)
	if !ok {
		t.Fatalf(
			"expected TradeExecutedEvent, got %T",
			event,
		)
	}

	trade := tradeEvent.Trade

	// Signal/reference price must remain Day 16 Close.
	if order.Price != signalCandle.Close {
		t.Fatalf(
			"expected order price %.2f, got %.2f",
			signalCandle.Close,
			order.Price,
		)
	}

	// Execution price must be Day 17 Open.
	if trade.ExecutedPrice != nextCandle.Open {
		t.Fatalf(
			"expected executed price %.2f, got %.2f",
			nextCandle.Open,
			trade.ExecutedPrice,
		)
	}

	// Execution time must be Day 17.
	if !trade.ExecutedTime.Equal(nextCandle.Timestamp) {
		t.Fatalf(
			"expected executed time %v, got %v",
			nextCandle.Timestamp,
			trade.ExecutedTime,
		)
	}

	// Trade must never execute at the signal candle.
	if trade.ExecutedTime.Equal(signalCandle.Timestamp) {
		t.Fatal("trade executed on the same candle as the signal")
	}
}

func TestMatchingEngineRejectsInvalidExecutionValues(t *testing.T) {
	tests := []struct {
		name          string
		open          float64
		quantity      float64
		wantErrReason error
	}{
		{name: "zero price", open: 0, quantity: 1, wantErrReason: ErrInvalidExecutionPrice},
		{name: "negative price", open: -1, quantity: 1, wantErrReason: ErrInvalidExecutionPrice},
		{name: "nan price", open: math.NaN(), quantity: 1, wantErrReason: ErrInvalidExecutionPrice},
		{name: "infinite price", open: math.Inf(1), quantity: 1, wantErrReason: ErrInvalidExecutionPrice},
		{name: "nan quantity", open: 100, quantity: math.NaN(), wantErrReason: ErrInvalidExecutionQuantity},
		{name: "infinite quantity", open: 100, quantity: math.Inf(1), wantErrReason: ErrInvalidExecutionQuantity},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			queue := events.NewEventQueue()
			engine := mustNewEngine(t, queue, newTestCashReserver())
			order := domain.NewOrder(1, "BTCUSDT", domain.BuyOrder, test.quantity, 100, time.Unix(1, 0))
			if err := engine.Consume(events.NewOrderCreatedEvent(order)); err != nil {
				t.Fatal(err)
			}
			if err := engine.ExecutePending(domain.Candle{Timestamp: time.Unix(2, 0), Open: test.open}); err != nil {
				t.Fatal(err)
			}
			if len(engine.Trades()) != 0 {
				t.Fatal("invalid execution must not create a trade")
			}
			event, ok := queue.Pop()
			if !ok {
				t.Fatal("expected OrderRejectedEvent")
			}
			rejected, ok := event.(events.OrderRejectedEvent)
			if !ok || !strings.Contains(rejected.Reason, test.wantErrReason.Error()) {
				t.Fatalf("expected rejection reason containing %q, got %#v", test.wantErrReason, event)
			}
		})
	}
}

func TestMatchingEngineRejectedEventUsesExecutionTimestamp(t *testing.T) {
	queue := events.NewEventQueue()
	engine := mustNewEngine(t, queue, newTestCashReserver())
	createdAt := time.Unix(1, 0).UTC()
	rejectedAt := time.Unix(2, 0).UTC()
	order := domain.NewOrder(1, "BTCUSDT", domain.BuyOrder, 1, 100, createdAt)
	if err := engine.Consume(events.NewOrderCreatedEvent(order)); err != nil {
		t.Fatal(err)
	}
	if err := engine.ExecutePending(domain.Candle{Timestamp: rejectedAt, Open: 0}); err != nil {
		t.Fatal(err)
	}
	event, ok := queue.Pop()
	if !ok {
		t.Fatal("expected OrderRejectedEvent")
	}
	rejected, ok := event.(events.OrderRejectedEvent)
	if !ok {
		t.Fatalf("event type = %T, want OrderRejectedEvent", event)
	}
	if rejected.Order.RejectedAt == nil || !rejected.Order.RejectedAt.Equal(rejectedAt) {
		t.Fatalf("RejectedAt = %v, want %v", rejected.Order.RejectedAt, rejectedAt)
	}
	if !rejected.Timestamp().Equal(rejectedAt) {
		t.Fatalf("event timestamp = %v, want %v", rejected.Timestamp(), rejectedAt)
	}
}

func TestExecutePendingErrorKeepsOnlyUnprocessedOrders(t *testing.T) {
	queue := events.NewEventQueue()
	engine := mustNewEngine(t, queue, newTestCashReserver())
	timestamp := time.Unix(1, 0).UTC()
	first := domain.NewOrder(1, "BTCUSDT", domain.BuyOrder, 1, 100, timestamp)
	first.ID = 1
	second := domain.NewOrder(1, "BTCUSDT", domain.BuyOrder, 1, 100, timestamp)
	second.ID = 2
	if err := second.Fill(timestamp); err != nil {
		t.Fatal(err)
	}
	third := domain.NewOrder(1, "BTCUSDT", domain.BuyOrder, 1, 100, timestamp)
	third.ID = 3
	for _, order := range []domain.Order{first, second, third} {
		if err := engine.Consume(events.NewOrderCreatedEvent(order)); err != nil {
			t.Fatal(err)
		}
	}

	if err := engine.ExecutePending(domain.Candle{Timestamp: timestamp.Add(time.Minute), Open: 120}); err == nil {
		t.Fatal("expected Fill error for already-filled order")
	}
	if len(engine.Trades()) != 1 {
		t.Fatalf("trades = %d, want 1", len(engine.Trades()))
	}
	if len(engine.pendingOrders) != 1 || engine.pendingOrders[0].ID != third.ID {
		t.Fatalf("pending orders after error = %#v, want only order %d", engine.pendingOrders, third.ID)
	}
}

func TestMatchingEngineRejectsWhenReservationFails(t *testing.T) {
	queue := events.NewEventQueue()
	reserver := newTestCashReserver()
	reserver.reservations[99] = 1
	engine := mustNewEngine(t, queue, reserver)

	order := domain.NewOrder(1, "BTCUSDT", domain.BuyOrder, 1, 100, time.Unix(1, 0))
	order.ID = 99
	if err := engine.Consume(events.NewOrderCreatedEvent(order)); err != nil {
		t.Fatal(err)
	}
	if err := engine.ExecutePending(domain.Candle{Timestamp: time.Unix(2, 0), Open: 120}); err != nil {
		t.Fatal(err)
	}
	if len(engine.Trades()) != 0 {
		t.Fatal("reservation failure must not create a trade")
	}
	if queue.Len() != 1 {
		t.Fatalf("expected one rejection event, got %d", queue.Len())
	}
}

func TestMatchingEngineRejectsInvalidQuantityAtConsume(t *testing.T) {
	for _, quantity := range []float64{0, -1} {
		queue := events.NewEventQueue()
		engine := mustNewEngine(t, queue, newTestCashReserver())
		order := domain.NewOrder(1, "BTCUSDT", domain.BuyOrder, quantity, 100, time.Unix(1, 0))
		if err := engine.Consume(events.NewOrderCreatedEvent(order)); err != nil {
			t.Fatalf("quantity %.2f: Consume() error = %v", quantity, err)
		}
		if len(engine.Trades()) != 0 || queue.Len() != 1 {
			t.Fatalf("quantity %.2f: expected one rejection and no trade", quantity)
		}
		event, _ := queue.Pop()
		if _, ok := event.(events.OrderRejectedEvent); !ok {
			t.Fatalf("quantity %.2f: expected OrderRejectedEvent", quantity)
		}
	}
}

func TestMatchingEngineExecutionReservationWithPortfolio(t *testing.T) {
	tests := []struct {
		name       string
		cash       float64
		open       float64
		orders     int
		wantTrades int
		wantCash   float64
	}{
		{name: "gap but affordable", cash: 1000, open: 120, orders: 1, wantTrades: 1, wantCash: 879.88},
		{name: "gap unaffordable", cash: 1000, open: 1000, orders: 1, wantTrades: 0, wantCash: 1000},
		{name: "fee unaffordable", cash: 100, open: 100, orders: 1, wantTrades: 0, wantCash: 100},
		{name: "multiple orders reserve cumulatively", cash: 150, open: 100, orders: 2, wantTrades: 1, wantCash: 49.9},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			queue := events.NewEventQueue()
			portfolioEngine := portfolio.NewEngine(queue, domain.NewPortfolio(1, test.cash))
			engine := mustNewEngine(t, queue, portfolioEngine)
			for index := 0; index < test.orders; index++ {
				order := domain.NewOrder(1, "BTCUSDT", domain.BuyOrder, 1, 100, time.Unix(1, 0))
				order.ID = domain.OrderID(index + 1)
				if err := engine.Consume(events.NewOrderCreatedEvent(order)); err != nil {
					t.Fatal(err)
				}
			}

			if err := engine.ExecutePending(domain.Candle{Timestamp: time.Unix(2, 0), Open: test.open}); err != nil {
				t.Fatal(err)
			}
			for !queue.IsEmpty() {
				event, _ := queue.Pop()
				switch event.(type) {
				case events.OrderRejectedEvent, events.PortfolioUpdatedEvent:
					continue
				case events.TradeExecutedEvent:
					if err := portfolioEngine.Consume(event); err != nil {
						t.Fatal(err)
					}
				}
			}

			if len(engine.Trades()) != test.wantTrades {
				t.Fatalf("trades = %d, want %d", len(engine.Trades()), test.wantTrades)
			}
			if got := portfolioEngine.Portfolio().Cash; math.Abs(got-test.wantCash) > 1e-9 {
				t.Fatalf("cash = %.12f, want %.12f", got, test.wantCash)
			}
			if len(portfolioEngine.Positions()) != test.wantTrades {
				t.Fatalf("positions = %d, want %d", len(portfolioEngine.Positions()), test.wantTrades)
			}
		})
	}
}
