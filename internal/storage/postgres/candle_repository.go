package postgres

import (
	"context"
	"time"

	"event-driven-backtesting-engine/internal/domain"

	"github.com/jackc/pgx/v5"
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
			AND timestamp < $4
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

func (r *CandleRepository) GetLatestTimestamp(
	ctx context.Context,
	symbol string,
	timeframe string,
) (time.Time, bool, error) {
	const query = `
		SELECT timestamp
		FROM candles
		WHERE symbol = $1
			AND timeframe = $2
		ORDER BY timestamp DESC
		LIMIT 1
	`

	var timestamp time.Time
	if err := r.pool.QueryRow(ctx, query, symbol, timeframe).Scan(&timestamp); err != nil {
		if err == pgx.ErrNoRows {
			return time.Time{}, false, nil
		}
		return time.Time{}, false, err
	}

	return timestamp, true, nil
}

const insertChunkSize = 2000

func (r *CandleRepository) InsertCandles(
	ctx context.Context,
	candles []domain.Candle,
) (int64, error) {
	if len(candles) == 0 {
		return 0, nil
	}

	const query = `
		INSERT INTO candles (
			timestamp, symbol, timeframe, open, high, low, close, volume
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (timestamp, symbol, timeframe) DO NOTHING
	`

	var totalInserted int64
	for i := 0; i < len(candles); i += insertChunkSize {
		end := i + insertChunkSize
		if end > len(candles) {
			end = len(candles)
		}
		chunk := candles[i:end]

		batch := &pgx.Batch{}
		for _, candle := range chunk {
			batch.Queue(
				query,
				candle.Timestamp,
				candle.Symbol,
				candle.Timeframe,
				candle.Open,
				candle.High,
				candle.Low,
				candle.Close,
				candle.Volume,
			)
		}

		results := r.pool.SendBatch(ctx, batch)
		for range chunk {
			commandTag, err := results.Exec()
			if err != nil {
				results.Close()
				return totalInserted, err
			}
			totalInserted += commandTag.RowsAffected()
		}
		results.Close()
	}

	return totalInserted, nil
}
