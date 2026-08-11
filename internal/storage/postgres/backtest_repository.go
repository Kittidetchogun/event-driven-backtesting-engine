package postgres

import (
	"context"

	"event-driven-backtesting-engine/internal/statistics"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BacktestRepository struct {
	pool *pgxpool.Pool
}

func NewBacktestRepository(pool *pgxpool.Pool) *BacktestRepository {
	return &BacktestRepository{
		pool: pool,
	}
}

// SaveBacktestResult stores the summary result of a completed backtest.
func (r *BacktestRepository) SaveBacktestResult(
	ctx context.Context,
	result statistics.BacktestResult,
) error {

	const query = `
		INSERT INTO backtest_runs (
			run_id,
			strategy_name,
			symbol,
			timeframe,
			start_date,
			end_date,
			total_return,
			win_rate,
			sharpe_ratio,
			max_drawdown
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10
		)
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		result.RunID,
		result.StrategyName,
		result.Symbol,
		result.Timeframe,
		result.StartDate,
		result.EndDate,
		result.TotalReturn,
		result.WinRate,
		result.SharpeRatio,
		result.MaxDrawdown,
	)

	return err
}

// GetBacktestResult returns a saved backtest result by run ID.
func (r *BacktestRepository) GetBacktestResult(
	ctx context.Context,
	runID int,
) (*statistics.BacktestResult, error) {

	const query = `
		SELECT
			run_id,
			strategy_name,
			symbol,
			timeframe,
			start_date,
			end_date,
			total_return,
			win_rate,
			sharpe_ratio,
			max_drawdown
		FROM backtest_runs
		WHERE run_id = $1
	`

	var result statistics.BacktestResult

	err := r.pool.QueryRow(
		ctx,
		query,
		runID,
	).Scan(
		&result.RunID,
		&result.StrategyName,
		&result.Symbol,
		&result.Timeframe,
		&result.StartDate,
		&result.EndDate,
		&result.TotalReturn,
		&result.WinRate,
		&result.SharpeRatio,
		&result.MaxDrawdown,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &result, nil
}
