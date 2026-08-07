package statistics

import (
	"testing"
	"time"

	"event-driven-backtesting-engine/internal/domain"
	"event-driven-backtesting-engine/internal/events"
)

func TestNewStatisticsEngine(t *testing.T) {

	engine := NewEngine()

	if engine == nil {
		t.Fatal("expected engine")
	}

	if len(engine.Snapshots()) != 0 {
		t.Fatalf(
			"expected 0 snapshots, got %d",
			len(engine.Snapshots()),
		)
	}
}

func TestStatisticsEngineConsume(t *testing.T) {

	engine := NewEngine()

	portfolio := domain.NewPortfolio(
		1,
		100000,
	)

	event := events.NewPortfolioUpdatedEvent(
		portfolio,
	)

	if err := engine.Consume(event); err != nil {
		t.Fatalf(
			"unexpected error %v",
			err,
		)
	}

	if len(engine.Snapshots()) != 1 {
		t.Fatalf(
			"expected 1 snapshot, got %d",
			len(engine.Snapshots()),
		)
	}

	snapshot := engine.Snapshots()[0]

	if snapshot.Cash != portfolio.Cash {
		t.Fatalf(
			"expected cash %.2f, got %.2f",
			portfolio.Cash,
			snapshot.Cash,
		)
	}

	if snapshot.Equity != portfolio.Equity {
		t.Fatalf(
			"expected equity %.2f, got %.2f",
			portfolio.Equity,
			snapshot.Equity,
		)
	}
}

func TestStatisticsEngineConsume_InvalidEvent(t *testing.T) {

	engine := NewEngine()

	candle := domain.Candle{
		Timestamp: time.Now(),
		Symbol:    "BTCUSDT",
		Timeframe: "1m",
		Open:      100,
		High:      110,
		Low:       90,
		Close:     105,
		Volume:    1000,
	}

	event := events.NewCandleReceivedEvent(candle)

	err := engine.Consume(event)

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestStatisticsEngineAppendMultipleSnapshots(t *testing.T) {

	engine := NewEngine()

	values := []float64{
		100000,
		101000,
		99000,
	}

	for _, equity := range values {

		portfolio := domain.NewPortfolio(
			1,
			100000,
		)

		portfolio.Equity = equity

		event := events.NewPortfolioUpdatedEvent(
			portfolio,
		)

		if err := engine.Consume(event); err != nil {
			t.Fatal(err)
		}
	}

	if len(engine.Snapshots()) != len(values) {
		t.Fatalf(
			"expected %d snapshots, got %d",
			len(values),
			len(engine.Snapshots()),
		)
	}
}

func TestStatisticsEngineEquityCurve(t *testing.T) {

	engine := NewEngine()

	values := []float64{
		100000,
		101000,
		99500,
		105000,
	}

	for _, equity := range values {

		portfolio := domain.NewPortfolio(
			1,
			100000,
		)

		portfolio.Equity = equity

		event := events.NewPortfolioUpdatedEvent(
			portfolio,
		)

		if err := engine.Consume(event); err != nil {
			t.Fatal(err)
		}
	}

	curve := engine.EquityCurve()

	if len(curve) != len(values) {
		t.Fatalf(
			"expected %d values, got %d",
			len(values),
			len(curve),
		)
	}

	for i := range values {

		if curve[i] != values[i] {
			t.Fatalf(
				"expected %.2f, got %.2f",
				values[i],
				curve[i],
			)
		}
	}
}
