package marketdata

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"event-driven-backtesting-engine/internal/domain"
)

type fakeCandleRepository struct {
	mu      sync.Mutex
	candles []domain.Candle
}

func (r *fakeCandleRepository) GetLatestTimestamp(
	ctx context.Context,
	symbol string,
	timeframe string,
) (time.Time, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var latest time.Time
	found := false
	for _, candle := range r.candles {
		if candle.Symbol != symbol || candle.Timeframe != timeframe {
			continue
		}
		if !found || candle.Timestamp.After(latest) {
			latest = candle.Timestamp
			found = true
		}
	}

	return latest, found, nil
}

func (r *fakeCandleRepository) InsertCandles(
	ctx context.Context,
	candles []domain.Candle,
) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing := make(map[string]struct{}, len(r.candles))
	for _, candle := range r.candles {
		existing[candleKey(candle)] = struct{}{}
	}

	var inserted int64
	for _, candle := range candles {
		key := candleKey(candle)
		if _, ok := existing[key]; ok {
			continue
		}
		r.candles = append(r.candles, candle)
		existing[key] = struct{}{}
		inserted++
	}

	return inserted, nil
}

func TestUpdaterSyncDownloadsOnlyMissingCandles(t *testing.T) {
	interval := 24 * time.Hour
	latestAvailable := lastClosedOpen(time.Now().UTC(), interval)
	existing := latestAvailable.Add(-interval)
	repository := &fakeCandleRepository{
		candles: []domain.Candle{testCandle(existing)},
	}

	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		startTime, err := strconv.ParseInt(r.URL.Query().Get("startTime"), 10, 64)
		if err != nil {
			t.Fatalf("parse startTime: %v", err)
		}

		if startTime > latestAvailable.UnixMilli() {
			_, _ = io.WriteString(w, "[]")
			return
		}

		_, _ = io.WriteString(w, `[[`+strconv.FormatInt(latestAvailable.UnixMilli(), 10)+`,"1","2","0.5","1.5","10"]]`)
	}))
	defer server.Close()

	updater := newTestUpdater(t, repository, server.URL, existing)
	if err := updater.Sync(context.Background()); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	if requestCount != 1 {
		t.Fatalf("request count = %d, want 1", requestCount)
	}
	if len(repository.candles) != 2 {
		t.Fatalf("repository candle count = %d, want 2", len(repository.candles))
	}
	if !repository.candles[1].Timestamp.Equal(latestAvailable) {
		t.Fatalf("inserted timestamp = %s, want %s", repository.candles[1].Timestamp, latestAvailable)
	}
}

func TestUpdaterSyncIsIdempotent(t *testing.T) {
	interval := 24 * time.Hour
	latestAvailable := lastClosedOpen(time.Now().UTC(), interval)
	repository := &fakeCandleRepository{
		candles: []domain.Candle{testCandle(latestAvailable)},
	}

	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		_, _ = io.WriteString(w, `[[`+strconv.FormatInt(latestAvailable.UnixMilli(), 10)+`,"1","2","0.5","1.5","10"]]`)
	}))
	defer server.Close()

	updater := newTestUpdater(t, repository, server.URL, latestAvailable)
	if err := updater.Sync(context.Background()); err != nil {
		t.Fatalf("first Sync() error = %v", err)
	}
	if err := updater.Sync(context.Background()); err != nil {
		t.Fatalf("second Sync() error = %v", err)
	}

	if requestCount != 0 {
		t.Fatalf("request count = %d, want 0", requestCount)
	}
	if len(repository.candles) != 1 {
		t.Fatalf("repository candle count = %d, want 1", len(repository.candles))
	}
}

func TestValidateCandlesDetectsDuplicateTimestamp(t *testing.T) {
	timestamp := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	err := ValidateCandles([]domain.Candle{
		testCandle(timestamp),
		testCandle(timestamp),
	})

	if !errors.Is(err, ErrDuplicateTimestamp) {
		t.Fatalf("ValidateCandles() error = %v, want ErrDuplicateTimestamp", err)
	}
}

func TestValidateCandlesDetectsOutOfOrderTimestamp(t *testing.T) {
	first := time.Date(2026, 7, 2, 0, 0, 0, 0, time.UTC)
	second := first.Add(-24 * time.Hour)
	err := ValidateCandles([]domain.Candle{
		testCandle(first),
		testCandle(second),
	})

	if !errors.Is(err, ErrOutOfOrderCandle) {
		t.Fatalf("ValidateCandles() error = %v, want ErrOutOfOrderCandle", err)
	}
}

func newTestUpdater(
	t *testing.T,
	repository CandleRepository,
	binanceURL string,
	startDate time.Time,
) *Updater {
	t.Helper()

	updater, err := NewUpdater(
		repository,
		Config{
			Symbol:       "BTCUSDT",
			Interval:     "1d",
			StartDate:    startDate,
			BinanceURL:   binanceURL,
			RequestLimit: 1000,
		},
		log.New(io.Discard, "", 0),
	)
	if err != nil {
		t.Fatalf("NewUpdater() error = %v", err)
	}

	return updater
}

func testCandle(timestamp time.Time) domain.Candle {
	return domain.Candle{
		Timestamp: timestamp,
		Symbol:    "BTCUSDT",
		Timeframe: "1d",
		Open:      1,
		High:      2,
		Low:       0.5,
		Close:     1.5,
		Volume:    10,
	}
}

func candleKey(candle domain.Candle) string {
	return candle.Symbol + "|" + candle.Timeframe + "|" + candle.Timestamp.Format(time.RFC3339Nano)
}
