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

	order := domain.NewOrder(
		1,
		"BTCUSDT",
		domain.BuyOrder,
		1,
		50000,
		time.Now(),
	)

	orderEvent := events.NewOrderCreatedEvent(order)

	if err := engine.Consume(orderEvent); err != nil {
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

	if trade.Symbol != order.Symbol {
		t.Fatalf("expected symbol %s got %s", order.Symbol, trade.Symbol)
	}

	if trade.Side != order.Side {
		t.Fatalf("expected side %s got %s", order.Side, trade.Side)
	}

	if trade.Quantity != order.Quantity {
		t.Fatalf("expected quantity %.2f got %.2f",
			order.Quantity,
			trade.Quantity,
		)
	}

	if trade.ExecutedPrice != order.Price {
		t.Fatalf("expected executed price %.2f got %.2f",
			order.Price,
			trade.ExecutedPrice,
		)
	}
}
