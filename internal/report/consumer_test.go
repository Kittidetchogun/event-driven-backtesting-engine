package report

import (
	"testing"
	"time"

	"event-driven-backtesting-engine/internal/events"
	"event-driven-backtesting-engine/internal/domain"
)

func TestConsumerConsume_BacktestCompletedEvent(t *testing.T) {
	consumer := NewConsumer()

	startDate := time.Date(
		2024,
		1,
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	)

	endDate := time.Date(
		2024,
		1,
		31,
		0,
		0,
		0,
		0,
		time.UTC,
	)

	event := events.NewBacktestCompletedEvent(
		1,
		"EMA Cross",
		"BTCUSDT",
		"1d",
		startDate,
		endDate,
		0.10,
		0.60,
		1.20,
		-0.05,
	)

	if err := consumer.Consume(event); err != nil {
		t.Fatalf("Consume() error = %v", err)
	}

	results := consumer.Results()

	if len(results) != 1 {
		t.Fatalf(
			"results = %d, want 1",
			len(results),
		)
	}

	result := results[0]

	if result.RunID != 1 {
		t.Errorf(
			"RunID = %d, want 1",
			result.RunID,
		)
	}

	if result.StrategyName != "EMA Cross" {
		t.Errorf(
			"StrategyName = %q, want %q",
			result.StrategyName,
			"EMA Cross",
		)
	}

	if result.Symbol != "BTCUSDT" {
		t.Errorf(
			"Symbol = %q, want %q",
			result.Symbol,
			"BTCUSDT",
		)
	}

	if result.TotalReturn != 0.10 {
		t.Errorf(
			"TotalReturn = %f, want 0.10",
			result.TotalReturn,
		)
	}
}

func TestConsumerConsume_UnsupportedEvent(t *testing.T) {
	consumer := NewConsumer()

	trade := domain.NewTrade(
        1,
        1,
        1,
        "BTCUSDT",
        domain.BuyOrder,
        1,
        50000,
        10,
        time.Now(),
    )

    event := events.NewTradeExecutedEvent(trade)

	if err := consumer.Consume(event); err == nil {
		t.Fatal("expected error for unsupported event")
	}
}
