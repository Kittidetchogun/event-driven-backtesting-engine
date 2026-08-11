package statistics

import "context"

type BacktestResultRepository interface {
	SaveBacktestResult(
		ctx context.Context,
		result BacktestResult,
	) error
}
