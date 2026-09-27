package matching

import (
	"testing"
	"time"

	"event-driven-backtesting-engine/internal/domain"
	"event-driven-backtesting-engine/internal/events"
)

func TestMatchingEngineConsume(t *testing.T) {
	queue := events.NewEventQueue()
	engine := NewEngine(queue)

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
	engine := NewEngine(queue)

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
