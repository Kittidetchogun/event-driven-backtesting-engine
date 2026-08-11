package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"event-driven-backtesting-engine/internal/statistics"

	"github.com/jackc/pgx/v5/pgxpool"
)

func newTestBacktestRepository(t *testing.T) (*BacktestRepository, *pgxpool.Pool) {
	t.Helper()

	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		t.Skip("DATABASE_URL is not set")
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("failed to create postgres pool: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("failed to ping postgres: %v", err)
	}

	repository := NewBacktestRepository(pool)

	return repository, pool
}

func TestBacktestRepositorySaveAndGet(t *testing.T) {
	repository, pool := newTestBacktestRepository(t)

	ctx := context.Background()

	startDate := time.Date(
		2026,
		1,
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	)

	endDate := time.Date(
		2026,
		3,
		31,
		0,
		0,
		0,
		0,
		time.UTC,
	)

	result := statistics.BacktestResult{
		RunID:        999001,
		StrategyName: "Test Strategy",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		StartDate:    startDate,
		EndDate:      endDate,
		TotalReturn:  0.15,
		WinRate:      0.60,
		SharpeRatio:  1.25,
		MaxDrawdown:  0.20,
	}

	// Cleanup test data first.
	t.Cleanup(func() {
		pool.Close()
	})

	t.Cleanup(func() {
		_, err := pool.Exec(
			ctx,
			"DELETE FROM backtest_runs WHERE run_id = $1",
			result.RunID,
		)

		if err != nil {
			t.Errorf("failed to cleanup test backtest: %v", err)
		}
	})

	// Save
	err := repository.SaveBacktestResult(ctx, result)
	if err != nil {
		t.Fatalf("failed to save backtest result: %v", err)
	}

	// Get
	got, err := repository.GetBacktestResult(ctx, result.RunID)
	if err != nil {
		t.Fatalf("failed to get backtest result: %v", err)
	}

	if got == nil {
		t.Fatal("expected backtest result, got nil")
	}

	if got.RunID != result.RunID {
		t.Errorf(
			"expected RunID %d, got %d",
			result.RunID,
			got.RunID,
		)
	}

	if got.StrategyName != result.StrategyName {
		t.Errorf(
			"expected strategy name %q, got %q",
			result.StrategyName,
			got.StrategyName,
		)
	}

	if got.Symbol != result.Symbol {
		t.Errorf(
			"expected symbol %q, got %q",
			result.Symbol,
			got.Symbol,
		)
	}

	if got.Timeframe != result.Timeframe {
		t.Errorf(
			"expected timeframe %q, got %q",
			result.Timeframe,
			got.Timeframe,
		)
	}

	if !got.StartDate.Equal(result.StartDate) {
		t.Errorf(
			"expected start date %v, got %v",
			result.StartDate,
			got.StartDate,
		)
	}

	if !got.EndDate.Equal(result.EndDate) {
		t.Errorf(
			"expected end date %v, got %v",
			result.EndDate,
			got.EndDate,
		)
	}

	if got.TotalReturn != result.TotalReturn {
		t.Errorf(
			"expected total return %.4f, got %.4f",
			result.TotalReturn,
			got.TotalReturn,
		)
	}

	if got.WinRate != result.WinRate {
		t.Errorf(
			"expected win rate %.4f, got %.4f",
			result.WinRate,
			got.WinRate,
		)
	}

	if got.SharpeRatio != result.SharpeRatio {
		t.Errorf(
			"expected sharpe ratio %.4f, got %.4f",
			result.SharpeRatio,
			got.SharpeRatio,
		)
	}

	if got.MaxDrawdown != result.MaxDrawdown {
		t.Errorf(
			"expected max drawdown %.4f, got %.4f",
			result.MaxDrawdown,
			got.MaxDrawdown,
		)
	}
}

func TestBacktestRepositoryGetNotFound(t *testing.T) {
	repository, pool := newTestBacktestRepository(t)

	t.Cleanup(func() {
		pool.Close()
	})

	ctx := context.Background()

	got, err := repository.GetBacktestResult(
		ctx,
		999999,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got != nil {
		t.Fatalf(
			"expected nil result for unknown run ID, got %+v",
			got,
		)
	}
}
