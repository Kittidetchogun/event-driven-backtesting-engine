package backtest

import (
	"context"
	"errors"
	"testing"
	"time"

	"event-driven-backtesting-engine/internal/domain"
	"event-driven-backtesting-engine/internal/events"
	"event-driven-backtesting-engine/internal/statistics"
)

type e2eErrorPipeline struct {
	candles []domain.Candle
	index   int
	err     error
}

func (p *e2eErrorPipeline) Next() (domain.Candle, bool, error) {
	if p.err != nil {
		return domain.Candle{}, false, p.err
	}

	if p.index >= len(p.candles) {
		return domain.Candle{}, false, nil
	}

	candle := p.candles[p.index]
	p.index++

	return candle, true, nil
}

type e2eErrorRepository struct {
	results []statistics.BacktestResult
	err     error
}

func (r *e2eErrorRepository) SaveBacktestResult(
	ctx context.Context,
	result statistics.BacktestResult,
) error {
	if r.err != nil {
		return r.err
	}

	r.results = append(r.results, result)

	return nil
}

// ------------------------------------------------------------
// Empty Dataset
// ------------------------------------------------------------

func TestBacktestEndToEnd_EmptyDataset(t *testing.T) {
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

	pipeline := &e2eErrorPipeline{
		candles: []domain.Candle{},
	}

	repository := &e2eErrorRepository{
		results: make([]statistics.BacktestResult, 0),
	}

	runner, err := NewRunner(
		RunnerConfig{
			RunID:          2001,
			StrategyName:   "EMA Cross",
			Symbol:         "BTCUSDT",
			Timeframe:      "1d",
			StartDate:      start,
			EndDate:        start,
			InitialCapital: 10000,
		},
		pipeline,
		repository,
	)

	if err != nil {
		t.Fatalf("NewRunner() error = %v", err)
	}

	result, err := runner.Run()

	if err != nil {
		t.Fatalf(
			"Runner.Run() returned unexpected error for empty dataset: %v",
			err,
		)
	}

	if result.RunID != 2001 {
		t.Errorf(
			"RunID = %d, want 2001",
			result.RunID,
		)
	}

	// Even with no candles, the runner should complete
	// without crashing and publish the final result.
	if len(repository.results) != 1 {
		t.Fatalf(
			"saved results = %d, want 1",
			len(repository.results),
		)
	}
}

// ------------------------------------------------------------
// Pipeline Error
// ------------------------------------------------------------

func TestBacktestEndToEnd_PipelineError(t *testing.T) {
	expectedErr := errors.New("historical candle source failed")

	pipeline := &e2eErrorPipeline{
		err: expectedErr,
	}

	repository := &e2eErrorRepository{
		results: make([]statistics.BacktestResult, 0),
	}

	runner, err := NewRunner(
		RunnerConfig{
			RunID:          2002,
			StrategyName:   "EMA Cross",
			Symbol:         "BTCUSDT",
			Timeframe:      "1d",
			InitialCapital: 10000,
		},
		pipeline,
		repository,
	)

	if err != nil {
		t.Fatalf("NewRunner() error = %v", err)
	}

	_, err = runner.Run()

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"Runner.Run() error = %v, want %v",
			err,
			expectedErr,
		)
	}

	// Backtest should stop immediately.
	if len(repository.results) != 0 {
		t.Errorf(
			"saved results = %d, want 0",
			len(repository.results),
		)
	}
}

// ------------------------------------------------------------
// Repository Error
// ------------------------------------------------------------

func TestBacktestEndToEnd_RepositoryError(t *testing.T) {
	expectedErr := errors.New("database unavailable")

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

	pipeline := &e2eErrorPipeline{
		candles: []domain.Candle{},
	}

	repository := &e2eErrorRepository{
		results: make([]statistics.BacktestResult, 0),
		err:     expectedErr,
	}

	runner, err := NewRunner(
		RunnerConfig{
			RunID:          2003,
			StrategyName:   "EMA Cross",
			Symbol:         "BTCUSDT",
			Timeframe:      "1d",
			StartDate:      start,
			EndDate:        start,
			InitialCapital: 10000,
		},
		pipeline,
		repository,
	)

	if err != nil {
		t.Fatalf("NewRunner() error = %v", err)
	}

	_, err = runner.Run()

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"Runner.Run() error = %v, want %v",
			err,
			expectedErr,
		)
	}
}

// ------------------------------------------------------------
// Invalid Event
// ------------------------------------------------------------

func TestBacktestEndToEnd_InvalidEvent(t *testing.T) {
	dispatcher := events.NewEventDispatcher()

	err := dispatcher.Dispatch(nil)

	if err == nil {
		t.Fatal(
			"expected error when dispatching nil event",
		)
	}
}
