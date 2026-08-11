package dashboard

import (
	"testing"
	"time"

	"event-driven-backtesting-engine/internal/events"
	"event-driven-backtesting-engine/internal/domain"
)

func TestConsumerConsume_BacktestCompletedEvent(t *testing.T) {
	consumer := NewConsumer()

	event := events.NewBacktestCompletedEvent(
		1,
		"EMA Cross",
		"BTCUSDT",
		"1d",
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
		0.25,
		0.60,
		1.50,
		-0.10,
	)

	if err := consumer.Consume(event); err != nil {
		t.Fatalf(
			"Consumer.Consume() error = %v",
			err,
		)
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
			"Symbol = %q, want BTCUSDT",
			result.Symbol,
		)
	}

	if result.TotalReturn != 0.25 {
		t.Errorf(
			"TotalReturn = %f, want 0.25",
			result.TotalReturn,
		)
	}

	if result.WinRate != 0.60 {
		t.Errorf(
			"WinRate = %f, want 0.60",
			result.WinRate,
		)
	}
}

func TestConsumerConsume_UnsupportedEvent(t *testing.T) {
	consumer := NewConsumer()
	candle := domain.Candle{
		Symbol:    "BTCUSDT",
		Timestamp: time.Now(),
		Close:     100,
	}

	event := events.NewCandleReceivedEvent(candle)

	if err := consumer.Consume(event); err == nil {
		t.Fatal(
			"expected error for unsupported event",
		)
	}
}
