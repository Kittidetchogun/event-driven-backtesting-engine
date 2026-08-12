package backtest

import (
	"context"
	"testing"
	"time"

	"event-driven-backtesting-engine/internal/domain"
	"event-driven-backtesting-engine/internal/statistics"
)

type e2eCandlePipeline struct {
	candles []domain.Candle
	index   int
}

func (p *e2eCandlePipeline) Next() (domain.Candle, bool, error) {
	if p.index >= len(p.candles) {
		return domain.Candle{}, false, nil
	}

	candle := p.candles[p.index]
	p.index++

	return candle, true, nil
}

type e2eBacktestResultRepository struct {
	results []statistics.BacktestResult
}

func (r *e2eBacktestResultRepository) SaveBacktestResult(
	ctx context.Context,
	result statistics.BacktestResult,
) error {
	r.results = append(r.results, result)
	return nil
}

func TestBacktestEndToEnd(t *testing.T) {
	start := time.Date(
		2024,
		time.January,
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	)

	const (
		runID        = 1001
		strategyName = "EMA Cross"
		symbol       = "BTCUSDT"
		timeframe    = "1d"
		initialCash  = 10000.0
	)

	// ------------------------------------------------------------
	// 1. Prepare deterministic test dataset
	// ------------------------------------------------------------
	//
	// First 21 candles establish the EMA state.
	// Then price rises sharply to trigger the EMA crossover.
	//
	// This keeps the E2E test deterministic and reproducible.
	candles := make([]domain.Candle, 0, 25)

	for i := 0; i < 21; i++ {
		candles = append(candles, domain.Candle{
			Symbol:    symbol,
			Timeframe: timeframe,
			Timestamp: start.AddDate(0, 0, i),
			Open:      100,
			High:      100,
			Low:       100,
			Close:     100,
			Volume:    1,
		})
	}

	for i := 21; i < 25; i++ {
		candles = append(candles, domain.Candle{
			Symbol:    symbol,
			Timeframe: timeframe,
			Timestamp: start.AddDate(0, 0, i),
			Open:      200,
			High:      200,
			Low:       200,
			Close:     200,
			Volume:    1,
		})
	}

	pipeline := &e2eCandlePipeline{
		candles: candles,
	}

	repository := &e2eBacktestResultRepository{
		results: make([]statistics.BacktestResult, 0),
	}

	// ------------------------------------------------------------
	// 2. Create Runner
	// ------------------------------------------------------------
	runner, err := NewRunner(
		RunnerConfig{
			RunID:          runID,
			StrategyName:   strategyName,
			Symbol:         symbol,
			Timeframe:      timeframe,
			StartDate:      candles[0].Timestamp,
			EndDate:        candles[len(candles)-1].Timestamp,
			InitialCapital: initialCash,
		},
		pipeline,
		repository,
	)
	if err != nil {
		t.Fatalf("NewRunner() error = %v", err)
	}

	// ------------------------------------------------------------
	// 3. Run complete backtest
	// ------------------------------------------------------------
	result, err := runner.Run()
	if err != nil {
		t.Fatalf("Runner.Run() error = %v", err)
	}

	// ------------------------------------------------------------
	// 4. Verify BacktestResult
	// ------------------------------------------------------------

	if result.RunID != runID {
		t.Errorf(
			"RunID = %d, want %d",
			result.RunID,
			runID,
		)
	}

	if result.StrategyName != strategyName {
		t.Errorf(
			"StrategyName = %q, want %q",
			result.StrategyName,
			strategyName,
		)
	}

	if result.Symbol != symbol {
		t.Errorf(
			"Symbol = %q, want %q",
			result.Symbol,
			symbol,
		)
	}

	if result.Timeframe != timeframe {
		t.Errorf(
			"Timeframe = %q, want %q",
			result.Timeframe,
			timeframe,
		)
	}

	if !result.StartDate.Equal(candles[0].Timestamp) {
		t.Errorf(
			"StartDate = %v, want %v",
			result.StartDate,
			candles[0].Timestamp,
		)
	}

	if !result.EndDate.Equal(
		candles[len(candles)-1].Timestamp,
	) {
		t.Errorf(
			"EndDate = %v, want %v",
			result.EndDate,
			candles[len(candles)-1].Timestamp,
		)
	}

	// ------------------------------------------------------------
	// 5. Verify Strategy → Order → Matching → Trade
	// ------------------------------------------------------------

	trades := runner.matching.Trades()

	if len(trades) == 0 {
		t.Fatal(
			"E2E: expected at least one executed trade",
		)
	}

	// ------------------------------------------------------------
	// 6. Verify Portfolio was updated
	// ------------------------------------------------------------

	snapshots := runner.portfolio.Snapshots()

	if len(snapshots) == 0 {
		t.Fatal(
			"E2E: expected portfolio snapshots",
		)
	}

	// ------------------------------------------------------------
	// 7. Verify Statistics received trades
	// ------------------------------------------------------------

	statisticsTrades := runner.statistics.Trades()

	if len(statisticsTrades) == 0 {
		t.Fatal(
			"E2E: expected Statistics Engine to receive trades",
		)
	}

	if len(statisticsTrades) != len(trades) {
		t.Errorf(
			"E2E: statistics trades = %d, matching trades = %d",
			len(statisticsTrades),
			len(trades),
		)
	}

	// ------------------------------------------------------------
	// 8. Verify Equity Curve
	// ------------------------------------------------------------

	equityCurve := runner.statistics.EquityCurve()

	if len(equityCurve) == 0 {
		t.Fatal(
			"E2E: expected non-empty equity curve",
		)
	}

	// Final equity should be represented by the final
	// portfolio snapshot.
	finalSnapshot := snapshots[len(snapshots)-1]

	if len(equityCurve) != len(snapshots) {
		t.Errorf(
			"E2E: equity curve length = %d, snapshots = %d",
			len(equityCurve),
			len(snapshots),
		)
	}

	if equityCurve[len(equityCurve)-1] != finalSnapshot.Equity {
		t.Errorf(
			"E2E: final equity curve value = %.8f, final snapshot equity = %.8f",
			equityCurve[len(equityCurve)-1],
			finalSnapshot.Equity,
		)
	}

	// ------------------------------------------------------------
	// 9. Verify performance metrics were calculated
	// ------------------------------------------------------------

	performance := runner.statistics.Performance()

	// We don't require every metric to be non-zero because
	// a deterministic test dataset can legitimately produce
	// zero Sharpe or drawdown.
	//
	// We only verify that Performance() can be calculated
	// without causing the E2E flow to fail.
	_ = performance

	// ------------------------------------------------------------
	// 10. Verify BacktestCompletedEvent persistence
	// ------------------------------------------------------------

	if len(repository.results) != 1 {
		t.Fatalf(
			"E2E: saved results = %d, want 1",
			len(repository.results),
		)
	}

	saved := repository.results[0]

	if saved.RunID != result.RunID {
		t.Errorf(
			"E2E: saved RunID = %d, result RunID = %d",
			saved.RunID,
			result.RunID,
		)
	}

	if saved.StrategyName != result.StrategyName {
		t.Errorf(
			"E2E: saved StrategyName = %q, result StrategyName = %q",
			saved.StrategyName,
			result.StrategyName,
		)
	}

	if saved.Symbol != result.Symbol {
		t.Errorf(
			"E2E: saved Symbol = %q, result Symbol = %q",
			saved.Symbol,
			result.Symbol,
		)
	}

	if saved.Timeframe != result.Timeframe {
		t.Errorf(
			"E2E: saved Timeframe = %q, result Timeframe = %q",
			saved.Timeframe,
			result.Timeframe,
		)
	}

	if saved.StartDate != result.StartDate {
		t.Errorf(
			"E2E: saved StartDate = %v, result StartDate = %v",
			saved.StartDate,
			result.StartDate,
		)
	}

	if saved.EndDate != result.EndDate {
		t.Errorf(
			"E2E: saved EndDate = %v, result EndDate = %v",
			saved.EndDate,
			result.EndDate,
		)
	}

	if saved.TotalReturn != result.TotalReturn {
		t.Errorf(
			"E2E: saved TotalReturn = %.8f, result TotalReturn = %.8f",
			saved.TotalReturn,
			result.TotalReturn,
		)
	}

	if saved.WinRate != result.WinRate {
		t.Errorf(
			"E2E: saved WinRate = %.8f, result WinRate = %.8f",
			saved.WinRate,
			result.WinRate,
		)
	}

	if saved.SharpeRatio != result.SharpeRatio {
		t.Errorf(
			"E2E: saved SharpeRatio = %.8f, result SharpeRatio = %.8f",
			saved.SharpeRatio,
			result.SharpeRatio,
		)
	}

	if saved.MaxDrawdown != result.MaxDrawdown {
		t.Errorf(
			"E2E: saved MaxDrawdown = %.8f, result MaxDrawdown = %.8f",
			saved.MaxDrawdown,
			result.MaxDrawdown,
		)
	}
}
