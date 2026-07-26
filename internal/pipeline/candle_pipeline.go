package pipeline

import (
	"context"
	"errors"
	"fmt"
	"time"

	"event-driven-backtesting-engine/internal/domain"
)

var (
	ErrDuplicateTimestamp = errors.New("duplicate candle timestamp")
	ErrOutOfOrderCandle   = errors.New("out-of-order candle timestamp")
)

type CandleRepository interface {
	GetCandles(
		ctx context.Context,
		symbol string,
		timeframe string,
		start time.Time,
		end time.Time,
	) ([]domain.Candle, error)
}

type CandlePipeline interface {
	Next() (domain.Candle, bool, error)
}

type CandleQuery struct {
	Symbol    string
	Timeframe string
	Start     time.Time
	End       time.Time
}

type HistoricalCandlePipeline struct {
	candles     []domain.Candle
	index       int
	previous    time.Time
	hasPrevious bool
}

func NewHistoricalCandlePipeline(
	ctx context.Context,
	repository CandleRepository,
	query CandleQuery,
) (*HistoricalCandlePipeline, error) {
	if repository == nil {
		return nil, errors.New("candle repository is required")
	}

	candles, err := repository.GetCandles(
		ctx,
		query.Symbol,
		query.Timeframe,
		query.Start,
		query.End,
	)
	if err != nil {
		return nil, err
	}

	return &HistoricalCandlePipeline{
		candles: candles,
	}, nil
}

func (p *HistoricalCandlePipeline) Next() (domain.Candle, bool, error) {
	if p.index >= len(p.candles) {
		return domain.Candle{}, false, nil
	}

	candle := p.candles[p.index]
	if p.hasPrevious {
		if candle.Timestamp.Equal(p.previous) {
			return domain.Candle{}, false, fmt.Errorf(
				"%w: %s",
				ErrDuplicateTimestamp,
				candle.Timestamp.Format(time.RFC3339Nano),
			)
		}

		if candle.Timestamp.Before(p.previous) {
			return domain.Candle{}, false, fmt.Errorf(
				"%w: previous=%s current=%s",
				ErrOutOfOrderCandle,
				p.previous.Format(time.RFC3339Nano),
				candle.Timestamp.Format(time.RFC3339Nano),
			)
		}
	}

	p.index++
	p.previous = candle.Timestamp
	p.hasPrevious = true

	return candle, true, nil
}
