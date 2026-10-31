package backtest

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"event-driven-backtesting-engine/internal/domain"
	"event-driven-backtesting-engine/internal/events"
	"event-driven-backtesting-engine/internal/statistics"
)

const runID = 1

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

// fakeBacktestResultRepository is used by the integration test
// to verify that Runner persists the completed BacktestResult.
type fakeBacktestResultRepository struct {
	results []statistics.BacktestResult
}

func (r *fakeBacktestResultRepository) SaveBacktestResult(
	ctx context.Context,
	result statistics.BacktestResult,
) error {
	r.results = append(r.results, result)
	return nil
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
			Open:      100,
			Close:     100,
		})
	}

	// Price rises sharply so Fast EMA crosses above Slow EMA.
	for i := 21; i < 25; i++ {
		candles = append(candles, domain.Candle{
			Symbol:    "BTCUSDT",
			Timestamp: start.AddDate(0, 0, i),
			Open:      200,
			Close:     200,
		})
	}

	p := &fakeCandlePipeline{
		candles: candles,
	}

	repository := &fakeBacktestResultRepository{
		results: make([]statistics.BacktestResult, 0),
	}

	runner, err := NewRunner(
		RunnerConfig{
			RunID:          runID,
			StrategyName:   "EMA Cross",
			Symbol:         "BTCUSDT",
			Timeframe:      "1d",
			StartDate:      candles[0].Timestamp,
			EndDate:        candles[len(candles)-1].Timestamp,
			InitialCapital: 10000,
		},
		p,
		repository,
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
	// Statistics consistency
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

	// -------------------------
	// Phase 10.7
	// BacktestResult persistence
	// -------------------------

	if len(repository.results) != 1 {
		t.Fatalf(
			"saved results = %d, want 1",
			len(repository.results),
		)
	}

	saved := repository.results[0]

	if saved.RunID != result.RunID {
		t.Errorf(
			"saved RunID = %d, want %d",
			saved.RunID,
			result.RunID,
		)
	}

	if saved.StrategyName != result.StrategyName {
		t.Errorf(
			"saved StrategyName = %q, want %q",
			saved.StrategyName,
			result.StrategyName,
		)
	}

	if saved.Symbol != result.Symbol {
		t.Errorf(
			"saved Symbol = %q, want %q",
			saved.Symbol,
			result.Symbol,
		)
	}

	if saved.Timeframe != result.Timeframe {
		t.Errorf(
			"saved Timeframe = %q, want %q",
			saved.Timeframe,
			result.Timeframe,
		)
	}

	if !saved.StartDate.Equal(result.StartDate) {
		t.Errorf(
			"saved StartDate = %v, want %v",
			saved.StartDate,
			result.StartDate,
		)
	}

	if !saved.EndDate.Equal(result.EndDate) {
		t.Errorf(
			"saved EndDate = %v, want %v",
			saved.EndDate,
			result.EndDate,
		)
	}

	if saved.TotalReturn != result.TotalReturn {
		t.Errorf(
			"saved TotalReturn = %f, want %f",
			saved.TotalReturn,
			result.TotalReturn,
		)
	}

	if saved.WinRate != result.WinRate {
		t.Errorf(
			"saved WinRate = %f, want %f",
			saved.WinRate,
			result.WinRate,
		)
	}

	if saved.SharpeRatio != result.SharpeRatio {
		t.Errorf(
			"saved SharpeRatio = %f, want %f",
			saved.SharpeRatio,
			result.SharpeRatio,
		)
	}

	if saved.MaxDrawdown != result.MaxDrawdown {
		t.Errorf(
			"saved MaxDrawdown = %f, want %f",
			saved.MaxDrawdown,
			result.MaxDrawdown,
		)
	}
}

func TestRunnerRun_RecordsInitialFlatAndFinalSnapshots(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 2)
	candles := []domain.Candle{
		{Symbol: "BTCUSDT", Timestamp: start, Open: 100, Close: 100},
		{Symbol: "BTCUSDT", Timestamp: start.AddDate(0, 0, 1), Open: 100, Close: 100},
	}

	runner, err := NewRunner(
		RunnerConfig{
			RunID:          1,
			StrategyName:   "EMA Cross",
			Symbol:         "BTCUSDT",
			Timeframe:      "1d",
			StartDate:      start,
			EndDate:        end,
			InitialCapital: 10000,
		},
		&fakeCandlePipeline{candles: candles},
		&fakeBacktestResultRepository{},
	)
	if err != nil {
		t.Fatalf("NewRunner() error = %v", err)
	}

	if _, err := runner.Run(); err != nil {
		t.Fatalf("Runner.Run() error = %v", err)
	}

	snapshots := runner.statistics.Snapshots()
	if len(snapshots) != len(candles)+1 {
		t.Fatalf("snapshot count = %d, want %d", len(snapshots), len(candles)+1)
	}

	if !snapshots[0].Time.Equal(start) {
		t.Errorf("initial snapshot time = %v, want %v", snapshots[0].Time, start)
	}

	if snapshots[0].Equity != 10000 {
		t.Errorf("initial equity = %.2f, want 10000", snapshots[0].Equity)
	}

	last := snapshots[len(snapshots)-1]
	if !last.Time.Equal(end) {
		t.Errorf("final snapshot time = %v, want %v", last.Time, end)
	}

	if last.Equity != 10000 {
		t.Errorf("final equity = %.2f, want 10000", last.Equity)
	}

	if runner.statistics.EquityCurve()[0] != 10000 {
		t.Errorf(
			"equity curve initial value = %.2f, want 10000",
			runner.statistics.EquityCurve()[0],
		)
	}

	performance, err := runner.statistics.Performance()
	if err != nil {
		t.Fatal(err)
	}
	if performance.SharpeRatio != 0 {
		t.Errorf(
			"SharpeRatio = %f, want 0 for a flat equity curve",
			performance.SharpeRatio,
		)
	}

	if performance.MaxDrawdown != 0 {
		t.Errorf(
			"MaxDrawdown = %f, want 0 for a flat equity curve",
			performance.MaxDrawdown,
		)
	}
}

func TestRunnerRun_OpenPositionFinalEquityUsesFinalClose(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	candles := make([]domain.Candle, 0, 23)
	for i := 0; i < 21; i++ {
		candles = append(candles, domain.Candle{
			Symbol: "BTCUSDT", Timestamp: start.AddDate(0, 0, i), Open: 100, Close: 100,
		})
	}
	candles = append(candles,
		domain.Candle{Symbol: "BTCUSDT", Timestamp: start.AddDate(0, 0, 21), Open: 200, Close: 200},
		domain.Candle{Symbol: "BTCUSDT", Timestamp: start.AddDate(0, 0, 22), Open: 200, Close: 300},
	)

	runner, err := NewRunner(
		RunnerConfig{
			RunID: 4001, StrategyName: "EMA Cross", Symbol: "BTCUSDT", Timeframe: "1d",
			StartDate: start, EndDate: candles[len(candles)-1].Timestamp, InitialCapital: 10000,
		},
		&fakeCandlePipeline{candles: candles},
		&fakeBacktestResultRepository{},
	)
	if err != nil {
		t.Fatal(err)
	}
	result, err := runner.Run()
	if err != nil {
		t.Fatal(err)
	}

	if len(runner.matching.Trades()) == 0 {
		t.Fatal("expected an executed trade")
	}
	if len(runner.portfolio.Positions()) != 1 {
		t.Fatal("expected an open position at the final candle")
	}
	position := runner.portfolio.Positions()["BTCUSDT"]
	if position.CurrentPrice != 300 {
		t.Fatalf("CurrentPrice = %.2f, want final Close 300", position.CurrentPrice)
	}
	portfolioState := runner.portfolio.Portfolio()
	if portfolioState.PositionValue != position.Quantity*300 {
		t.Fatalf("PositionValue = %.2f, want %.2f", portfolioState.PositionValue, position.Quantity*300)
	}
	if portfolioState.Equity != portfolioState.Cash+portfolioState.PositionValue {
		t.Fatalf("Equity = %.2f, want Cash + PositionValue = %.2f", portfolioState.Equity, portfolioState.Cash+portfolioState.PositionValue)
	}

	snapshots := runner.statistics.Snapshots()
	finalSnapshot := snapshots[len(snapshots)-1]
	if finalSnapshot.Equity != portfolioState.Equity || finalSnapshot.PositionValue != portfolioState.PositionValue {
		t.Fatalf("final Statistics snapshot = %+v, want Portfolio = %+v", finalSnapshot, portfolioState)
	}
	wantReturn := (portfolioState.Equity - 10000) / 10000
	if result.TotalReturn != wantReturn {
		t.Fatalf("TotalReturn = %.12f, want %.12f from final equity %.2f", result.TotalReturn, wantReturn, portfolioState.Equity)
	}
}

func TestRunnerRun_PropagatesPerformanceValidationError(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	runner, err := NewRunner(
		RunnerConfig{
			RunID: 5001, StrategyName: "EMA Cross", Symbol: "BTCUSDT", Timeframe: "1d",
			StartDate: start, EndDate: start, InitialCapital: 1000,
		},
		&fakeCandlePipeline{candles: nil},
		&fakeBacktestResultRepository{},
	)
	if err != nil {
		t.Fatal(err)
	}

	valid := domain.NewPortfolio(1, 1000)
	valid.UpdatedAt = start.Add(-time.Hour)
	if err := runner.statistics.Consume(events.NewPortfolioUpdatedEvent(valid)); err != nil {
		t.Fatal(err)
	}
	runner.statistics.Snapshots()[0].Equity = math.NaN()

	if _, err := runner.Run(); !errors.Is(err, statistics.ErrInvalidEquityCurve) {
		t.Fatalf("Runner.Run() error = %v, want ErrInvalidEquityCurve", err)
	}
}

func TestRunnerRun_RejectsPriceGapWithoutChangingPortfolio(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	candles := make([]domain.Candle, 0, 23)
	for i := 0; i < 21; i++ {
		candles = append(candles, domain.Candle{
			Symbol: "BTCUSDT", Timestamp: start.AddDate(0, 0, i),
			Open: 100, Close: 100,
		})
	}
	candles = append(candles,
		domain.Candle{Symbol: "BTCUSDT", Timestamp: start.AddDate(0, 0, 21), Open: 200, Close: 200},
		domain.Candle{Symbol: "BTCUSDT", Timestamp: start.AddDate(0, 0, 22), Open: 2000, Close: 2000},
	)

	runner, err := NewRunner(
		RunnerConfig{
			RunID: 3001, StrategyName: "EMA Cross", Symbol: "BTCUSDT", Timeframe: "1d",
			StartDate: start, EndDate: candles[len(candles)-1].Timestamp, InitialCapital: 1000,
		},
		&fakeCandlePipeline{candles: candles},
		&fakeBacktestResultRepository{},
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runner.Run(); err != nil {
		t.Fatal(err)
	}

	if len(runner.matching.Trades()) != 0 {
		t.Fatal("price gap should reject the order without creating a trade")
	}
	if got := runner.portfolio.Portfolio().Cash; got != 1000 {
		t.Fatalf("Cash = %.2f, want 1000", got)
	}
	if len(runner.portfolio.Positions()) != 0 {
		t.Fatal("rejected order must not create a position")
	}
}
