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

	order1 := domain.NewOrder(
		1,
		"BTCUSDT",
		domain.BuyOrder,
		1,
		50000,
		time.Now(),
	)

	orderEvent1 := events.NewOrderCreatedEvent(order1)

	if err := engine.Consume(orderEvent1); err != nil {
		t.Fatal(err)
	}

	if queue.Len() != 1 {
		t.Fatalf("expected queue length 1 got %d", queue.Len())
	}

	event, ok := queue.Pop()
	if !ok {
		t.Fatal("expected event")
	}

	tradeEvent, ok := event.(events.TradeExecutedEvent)
	if !ok {
		t.Fatal("expected TradeExecutedEvent")
	}

	trade := tradeEvent.Trade

	if trade.TradeID != 1 {
		t.Fatalf("expected trade ID 1, got %d", trade.TradeID)
	}

	if trade.Symbol != order1.Symbol {
		t.Fatalf("expected symbol %s got %s",
			order1.Symbol,
			trade.Symbol,
		)
	}

	if trade.Side != order1.Side {
		t.Fatalf("expected side %s got %s",
			order1.Side,
			trade.Side,
		)
	}

	if trade.Quantity != order1.Quantity {
		t.Fatalf("expected quantity %.2f got %.2f",
			order1.Quantity,
			trade.Quantity,
		)
	}

	if trade.ExecutedPrice != order1.Price {
		t.Fatalf("expected executed price %.2f got %.2f",
			order1.Price,
			trade.ExecutedPrice,
		)
	}

	if len(engine.Trades()) != 1 {
		t.Fatalf("expected 1 trade, got %d", len(engine.Trades()))
	}

	// -------------------------------------------------
	// Second order -> should generate TradeID = 2
	// -------------------------------------------------

	order2 := domain.NewOrder(
		2,
		"BTCUSDT",
		domain.SellOrder,
		1,
		51000,
		time.Now().Add(time.Minute),
	)

	orderEvent2 := events.NewOrderCreatedEvent(order2)

	if err := engine.Consume(orderEvent2); err != nil {
		t.Fatal(err)
	}

	if queue.Len() != 1 {
		t.Fatalf("expected queue length 1 got %d", queue.Len())
	}

	event, ok = queue.Pop()
	if !ok {
		t.Fatal("expected second event")
	}

	tradeEvent, ok = event.(events.TradeExecutedEvent)
	if !ok {
		t.Fatal("expected second TradeExecutedEvent")
	}

	trade2 := tradeEvent.Trade

	if trade2.TradeID != 2 {
		t.Fatalf("expected trade ID 2, got %d", trade2.TradeID)
	}

	if len(engine.Trades()) != 2 {
		t.Fatalf("expected 2 trades, got %d", len(engine.Trades()))
	}

	// Verify stored trades
	storedTrades := engine.Trades()

	if storedTrades[0].TradeID != 1 {
		t.Fatalf("expected first trade ID 1, got %d",
			storedTrades[0].TradeID)
	}

	if storedTrades[1].TradeID != 2 {
		t.Fatalf("expected second trade ID 2, got %d",
			storedTrades[1].TradeID)
	}
}
