package pipeline

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"event-driven-backtesting-engine/internal/domain"
)

type fakeCandleRepository struct {
	candles []domain.Candle
	err     error
	query   CandleQuery
	called  bool
}

func (r *fakeCandleRepository) GetCandles(
	ctx context.Context,
	symbol string,
	timeframe string,
	start time.Time,
	end time.Time,
) ([]domain.Candle, error) {
	r.called = true
	r.query = CandleQuery{
		Symbol:    symbol,
		Timeframe: timeframe,
		Start:     start,
		End:       end,
	}

	if r.err != nil {
		return nil, r.err
	}

	return r.candles, nil
}

func TestNewHistoricalCandlePipelineLoadsCandlesFromRepository(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	repository := &fakeCandleRepository{}
	query := CandleQuery{
		Symbol:    "BTCUSDT",
		Timeframe: "1d",
		Start:     start,
		End:       end,
	}

	_, err := NewHistoricalCandlePipeline(context.Background(), repository, query)
	if err != nil {
		t.Fatalf("NewHistoricalCandlePipeline() error = %v", err)
	}

	if !repository.called {
		t.Fatal("repository was not called")
	}

	if !reflect.DeepEqual(repository.query, query) {
		t.Fatalf("repository query = %+v, want %+v", repository.query, query)
	}
}

func TestHistoricalCandlePipelineReturnsCandlesSequentially(t *testing.T) {
	first := candleAt(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	second := candleAt(first.Timestamp.Add(24 * time.Hour))
	pipeline := newTestPipeline(t, []domain.Candle{first, second})

	gotFirst, ok, err := pipeline.Next()
	if err != nil || !ok {
		t.Fatalf("first Next() = (%+v, %v, %v), want candle, true, nil", gotFirst, ok, err)
	}
	if gotFirst != first {
		t.Fatalf("first candle = %+v, want %+v", gotFirst, first)
	}

	gotSecond, ok, err := pipeline.Next()
	if err != nil || !ok {
		t.Fatalf("second Next() = (%+v, %v, %v), want candle, true, nil", gotSecond, ok, err)
	}
	if gotSecond != second {
		t.Fatalf("second candle = %+v, want %+v", gotSecond, second)
	}
}

func TestHistoricalCandlePipelineDetectsEndOfData(t *testing.T) {
	pipeline := newTestPipeline(t, []domain.Candle{
		candleAt(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
	})

	if _, ok, err := pipeline.Next(); err != nil || !ok {
		t.Fatalf("first Next() ok = %v, err = %v; want true, nil", ok, err)
	}

	candle, ok, err := pipeline.Next()
	if err != nil {
		t.Fatalf("EOF Next() error = %v, want nil", err)
	}
	if ok {
		t.Fatalf("EOF Next() ok = true, candle = %+v; want false", candle)
	}
}

func TestHistoricalCandlePipelineDetectsDuplicateTimestamp(t *testing.T) {
	timestamp := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	pipeline := newTestPipeline(t, []domain.Candle{
		candleAt(timestamp),
		candleAt(timestamp),
	})

	if _, ok, err := pipeline.Next(); err != nil || !ok {
		t.Fatalf("first Next() ok = %v, err = %v; want true, nil", ok, err)
	}

	_, ok, err := pipeline.Next()
	if !errors.Is(err, ErrDuplicateTimestamp) {
		t.Fatalf("duplicate Next() error = %v, want ErrDuplicateTimestamp", err)
	}
	if ok {
		t.Fatal("duplicate Next() ok = true, want false")
	}
}

func TestHistoricalCandlePipelineDetectsOutOfOrderTimestamp(t *testing.T) {
	first := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	second := first.Add(-24 * time.Hour)
	pipeline := newTestPipeline(t, []domain.Candle{
		candleAt(first),
		candleAt(second),
	})

	if _, ok, err := pipeline.Next(); err != nil || !ok {
		t.Fatalf("first Next() ok = %v, err = %v; want true, nil", ok, err)
	}

	_, ok, err := pipeline.Next()
	if !errors.Is(err, ErrOutOfOrderCandle) {
		t.Fatalf("out-of-order Next() error = %v, want ErrOutOfOrderCandle", err)
	}
	if ok {
		t.Fatal("out-of-order Next() ok = true, want false")
	}
}

func TestNewHistoricalCandlePipelineReturnsRepositoryError(t *testing.T) {
	repositoryError := errors.New("repository failure")
	repository := &fakeCandleRepository{err: repositoryError}

	_, err := NewHistoricalCandlePipeline(context.Background(), repository, CandleQuery{})
	if !errors.Is(err, repositoryError) {
		t.Fatalf("NewHistoricalCandlePipeline() error = %v, want %v", err, repositoryError)
	}
}

func TestNewHistoricalCandlePipelineRequiresRepository(t *testing.T) {
	_, err := NewHistoricalCandlePipeline(context.Background(), nil, CandleQuery{})
	if err == nil {
		t.Fatal("NewHistoricalCandlePipeline() error = nil, want error")
	}
}

func newTestPipeline(t *testing.T, candles []domain.Candle) *HistoricalCandlePipeline {
	t.Helper()

	pipeline, err := NewHistoricalCandlePipeline(
		context.Background(),
		&fakeCandleRepository{candles: candles},
		CandleQuery{},
	)
	if err != nil {
		t.Fatalf("NewHistoricalCandlePipeline() error = %v", err)
	}

	return pipeline
}

func candleAt(timestamp time.Time) domain.Candle {
	return domain.Candle{
		Timestamp: timestamp,
		Symbol:    "BTCUSDT",
		Timeframe: "1d",
		Open:      1,
		High:      1,
		Low:       1,
		Close:     1,
		Volume:    1,
	}
}
