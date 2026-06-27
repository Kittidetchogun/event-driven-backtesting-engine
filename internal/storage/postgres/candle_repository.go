package postgres

import (
	"context"
	"time"

	"event-driven-backtesting-engine/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CandleRepository struct {
	pool *pgxpool.Pool
}

func NewCandleRepository(pool *pgxpool.Pool) *CandleRepository {
	return &CandleRepository{pool: pool}
}

func (r *CandleRepository) GetCandles(
	ctx context.Context,
	symbol string,
	timeframe string,
	start time.Time,
	end time.Time,
) ([]domain.Candle, error) {
	const query = `
		SELECT timestamp, symbol, timeframe, open, high, low, close, volume
		FROM candles
		WHERE symbol = $1
			AND timeframe = $2
			AND timestamp >= $3
			AND timestamp <= $4
		ORDER BY timestamp ASC
	`

	rows, err := r.pool.Query(ctx, query, symbol, timeframe, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	candles := make([]domain.Candle, 0)
	for rows.Next() {
		var candle domain.Candle
		if err := rows.Scan(
			&candle.Timestamp,
			&candle.Symbol,
			&candle.Timeframe,
			&candle.Open,
			&candle.High,
			&candle.Low,
			&candle.Close,
			&candle.Volume,
		); err != nil {
			return nil, err
		}

		candles = append(candles, candle)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return candles, nil
}
