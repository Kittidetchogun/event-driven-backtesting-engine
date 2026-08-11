package backtest

import (
	"testing"
	"time"

	"event-driven-backtesting-engine/internal/domain"
)

type fakeCandlePipeline struct {
	candles []domain.Candle
	index   int
}

func (p *fakeCandlePipeline) Next() (domain.Candle, bool, error) {
	if p.index >= len(p.candles) {
		return domain.Candle{}, false, nil
	}

	candle := p.candles[p.index]
	p.index++

	return candle, true, nil
}

func TestRunnerRun_Integration(t *testing.T) {
	start := time.Date(
		2024, 1, 1,
		0, 0, 0,
		0,
		time.UTC,
	)

	candles := make([]domain.Candle, 0, 30)

	// Initial candles: establish EMA state.
	for i := 0; i < 21; i++ {
		candles = append(candles, domain.Candle{
			Symbol:    "BTCUSDT",
			Timestamp: start.AddDate(0, 0, i),
			Close:     100,
		})
	}

	// Price rises sharply so Fast EMA crosses above Slow EMA.
	for i := 21; i < 25; i++ {
		candles = append(candles, domain.Candle{
			Symbol:    "BTCUSDT",
			Timestamp: start.AddDate(0, 0, i),
			Close:     200,
		})
	}

	p := &fakeCandlePipeline{
		candles: candles,
	}

	runner, err := NewRunner(
		RunnerConfig{
			RunID:          1,
			StrategyName:   "EMA Cross",
			Symbol:         "BTCUSDT",
			Timeframe:      "1d",
			StartDate:      candles[0].Timestamp,
			EndDate:        candles[len(candles)-1].Timestamp,
			InitialCapital: 10000,
		},
		p,
	)
	if err != nil {
		t.Fatalf("NewRunner() error = %v", err)
	}

	result, err := runner.Run()
	if err != nil {
		t.Fatalf("Runner.Run() error = %v", err)
	}

	// -------------------------
	// BacktestResult metadata
	// -------------------------

	if result.RunID != 1 {
		t.Errorf("RunID = %d, want 1", result.RunID)
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

	if result.Timeframe != "1d" {
		t.Errorf(
			"Timeframe = %q, want 1d",
			result.Timeframe,
		)
	}

	if !result.StartDate.Equal(candles[0].Timestamp) {
		t.Errorf(
			"StartDate = %v, want %v",
			result.StartDate,
			candles[0].Timestamp,
		)
	}

	if !result.EndDate.Equal(candles[len(candles)-1].Timestamp) {
		t.Errorf(
			"EndDate = %v, want %v",
			result.EndDate,
			candles[len(candles)-1].Timestamp,
		)
	}

	// -------------------------
	// Event-driven flow
	// -------------------------

	if len(runner.matching.Trades()) == 0 {
		t.Fatal("expected at least one executed trade")
	}

	if len(runner.portfolio.Snapshots()) == 0 {
		t.Fatal("expected portfolio snapshots")
	}

	if len(runner.statistics.Snapshots()) == 0 {
		t.Fatal("expected statistics snapshots")
	}

	// -------------------------
	// Result consistency
	// -------------------------

	if len(runner.statistics.Trades()) == 0 {
		t.Fatal("expected statistics engine to receive executed trades")
	}

	if len(runner.statistics.EquityCurve()) == 0 {
		t.Fatal("expected statistics equity curve")
	}

	if result.TotalReturn == 0 {
		t.Error("TotalReturn = 0, expected non-zero result")
	}
}
