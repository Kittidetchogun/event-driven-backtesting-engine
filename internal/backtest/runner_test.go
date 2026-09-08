package backtest

import (
	"testing"
	"time"

	"event-driven-backtesting-engine/internal/domain"
	"event-driven-backtesting-engine/internal/events"
)

type mockCandlePipeline struct {
	candles []domain.Candle
	index   int
}

func (p *mockCandlePipeline) Next() (domain.Candle, bool, error) {
	if p.index >= len(p.candles) {
		return domain.Candle{}, false, nil
	}

	candle := p.candles[p.index]
	p.index++

	return candle, true, nil
}

func TestRunnerRequiresPipeline(t *testing.T) {
    runner, err := NewRunner(
        RunnerConfig{
            RunID:          1,
            InitialCapital: 10000,
        },
        nil,
        nil,
    )

    if err == nil {
        t.Fatal("expected error when candle pipeline is nil")
    }

    if runner != nil {
        t.Fatal("expected nil runner")
    }
}

func TestRunnerContinuesAfterOrderRejectedEvent(t *testing.T) {
	queue := events.NewEventQueue()
	dispatcher := events.NewEventDispatcher()

	runner := &Runner{
		queue:      queue,
		dispatcher: dispatcher,
	}

	// Register the same rejection handler used by Runner.
	dispatcher.Register(
		events.OrderRejectedEventType,
		func(event events.Event) error {
			return nil
		},
	)

	// Event after the rejected order.
	continued := false

	const testEventType = "TestBacktestContinuesEvent"

	dispatcher.Register(
		testEventType,
		func(event events.Event) error {
			continued = true
			return nil
		},
	)

	now := time.Now()

	orderToReject := domain.NewOrder(
		1,
		"BTCUSDT",
		domain.BuyOrder,
		1,
		100000,
		now,
	)

	if err := orderToReject.Reject(
		"insufficient cash",
		now,
	); err != nil {
		t.Fatalf("failed to reject order: %v", err)
	}

	queue.Push(
		events.NewOrderRejectedEvent(
			orderToReject,
			"insufficient cash",
		),
	)

	queue.Push(
		events.NewBaseEvent(
			testEventType,
			now,
		),
	)

	if err := runner.processQueue(); err != nil {
		t.Fatalf(
			"processQueue() stopped after OrderRejectedEvent: %v",
			err,
		)
	}

	if !continued {
		t.Fatal(
			"backtest did not continue after OrderRejectedEvent",
		)
	}

	if !queue.IsEmpty() {
		t.Fatalf(
			"expected event queue to be empty, got %d events",
			queue.Len(),
		)
	}
}

type testBacktestContinuesEvent struct {
	events.BaseEvent
}

func (testBacktestContinuesEvent) Type() string {
	return "TestBacktestContinuesEvent"
}
