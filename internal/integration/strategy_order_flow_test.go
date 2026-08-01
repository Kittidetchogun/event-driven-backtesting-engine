package integration

import (
	"testing"
	"time"

	"event-driven-backtesting-engine/internal/domain"
	"event-driven-backtesting-engine/internal/events"
	"event-driven-backtesting-engine/internal/order"
	"event-driven-backtesting-engine/internal/strategy"
)

func TestStrategyToOrderManagerFlow(t *testing.T) {

	queue := events.NewEventQueue()

	dispatcher := events.NewEventDispatcher()

	manager := order.NewManager(
		domain.DummyPortfolioChecker{},
		queue,
	)

	dispatcher.Register(
		events.SignalGeneratedEventType,
		manager.Consume,
	)

	ema := strategy.NewEmaCross(2, 3)

	ema.SetDispatcher(dispatcher)

	if err := ema.Initialize(); err != nil {
		t.Fatal(err)
	}

	candles := []domain.Candle{
		newCandle(1),
		newCandle(1),
		newCandle(1),
		newCandle(10),
	}

	for _, candle := range candles {
		ema.OnData(candle)
	}

	if queue.Len() != 1 {
		t.Fatalf("expected 1 OrderCreatedEvent got %d", queue.Len())
	}

	event, ok := queue.Pop()
	if !ok {
		t.Fatal("expected event")
	}

	orderEvent, ok := event.(events.OrderCreatedEvent)
	if !ok {
		t.Fatalf("expected OrderCreatedEvent got %T", event)
	}

	if orderEvent.Order.Symbol != "BTCUSDT" {
		t.Fatalf("expected BTCUSDT got %s", orderEvent.Order.Symbol)
	}

	if orderEvent.Order.Side != domain.BuyOrder {
		t.Fatalf("expected BUY got %s", orderEvent.Order.Side)
	}
}

func newCandle(close float64) domain.Candle {
	return domain.Candle{
		Symbol:    "BTCUSDT",
		Timestamp: time.Now(),
		Close:     close,
	}
}