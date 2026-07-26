package marketdata

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"event-driven-backtesting-engine/internal/domain"

	"github.com/joho/godotenv"
)

const (
	defaultRequestLimit = 1000
	defaultHTTPTimeout  = 30 * time.Second
)

var (
	ErrDuplicateTimestamp = errors.New("duplicate downloaded candle timestamp")
	ErrOutOfOrderCandle   = errors.New("out-of-order downloaded candle timestamp")
)

type CandleRepository interface {
	GetLatestTimestamp(ctx context.Context, symbol string, timeframe string) (time.Time, bool, error)
	InsertCandles(ctx context.Context, candles []domain.Candle) (int64, error)
}

type Config struct {
	Symbol       string
	Interval     string
	StartDate    time.Time
	BinanceURL   string
	RequestLimit int
}

type SyncResult struct {
	Downloaded int
	Inserted   int64
}

type Updater struct {
	repository CandleRepository
	client     *http.Client
	logger     *log.Logger
	config     Config
	lastResult SyncResult
}

func NewUpdater(repository CandleRepository, config Config, logger *log.Logger) (*Updater, error) {
	if repository == nil {
		return nil, errors.New("candle repository is required")
	}
	if logger == nil {
		logger = log.New(io.Discard, "", 0)
	}
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	return &Updater{
		repository: repository,
		client: &http.Client{
			Timeout: defaultHTTPTimeout,
		},
		logger: logger,
		config: config,
	}, nil
}

func LoadConfigFromEnv() (Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("load .env: %w", err)
	}

	startDate, err := time.Parse("2006-01-02", os.Getenv("MARKET_DATA_START_DATE"))
	if err != nil {
		return Config{}, fmt.Errorf("parse MARKET_DATA_START_DATE: %w", err)
	}

	requestLimit := defaultRequestLimit
	if value := os.Getenv("BINANCE_REQUEST_LIMIT"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return Config{}, fmt.Errorf("parse BINANCE_REQUEST_LIMIT: %w", err)
		}
		requestLimit = parsed
	}

	config := Config{
		Symbol:       os.Getenv("MARKET_DATA_SYMBOL"),
		Interval:     os.Getenv("MARKET_DATA_INTERVAL"),
		StartDate:    startDate.UTC(),
		BinanceURL:   os.Getenv("BINANCE_KLINES_URL"),
		RequestLimit: requestLimit,
	}
	if err := validateConfig(config); err != nil {
		return Config{}, err
	}

	return config, nil
}

func (u *Updater) Sync(ctx context.Context) error {
	u.lastResult = SyncResult{}

	u.logger.Println("Loading latest timestamp...")

	latestTimestamp, exists, err := u.repository.GetLatestTimestamp(
		ctx,
		u.config.Symbol,
		u.config.Interval,
	)
	if err != nil {
		return fmt.Errorf("get latest candle timestamp: %w", err)
	}

	intervalDuration, err := intervalDuration(u.config.Interval)
	if err != nil {
		return err
	}

	start := u.config.StartDate
	if exists {
		u.logger.Printf("Latest candle:\n%s\n", latestTimestamp.Format("2006-01-02"))
		start = latestTimestamp.Add(intervalDuration)
	} else {
		u.logger.Println("Latest candle:\nnone")
	}

	end := lastClosedOpen(time.Now().UTC(), intervalDuration)
	if start.After(end) {
		u.logger.Println("Downloading new candles...")
		u.logger.Println("Downloaded:")
		u.logger.Println("0 candles")
		u.logger.Println("Inserted:")
		u.logger.Println("0 candles")
		u.logger.Println("Database synchronized successfully.")
		return nil
	}

	u.logger.Println("Downloading new candles...")
	candles, err := u.downloadCandles(ctx, start, end)
	if err != nil {
		return err
	}

	u.logger.Println("Downloaded:")
	u.logger.Printf("%d candles\n", len(candles))

	if err := ValidateCandles(candles); err != nil {
		return fmt.Errorf("validate downloaded candles: %w", err)
	}

	inserted, err := u.repository.InsertCandles(ctx, candles)
	if err != nil {
		return fmt.Errorf("insert downloaded candles: %w", err)
	}

	u.logger.Println("Inserted:")
	u.logger.Printf("%d candles\n", inserted)
	u.logger.Println("Database synchronized successfully.")
	u.lastResult = SyncResult{
		Downloaded: len(candles),
		Inserted:   inserted,
	}

	return nil
}

func (u *Updater) LastResult() SyncResult {
	return u.lastResult
}

func (u *Updater) downloadCandles(ctx context.Context, start time.Time, end time.Time) ([]domain.Candle, error) {
	startMS := toMilliseconds(start)
	endMS := toMilliseconds(end)
	candles := make([]domain.Candle, 0)

	for startMS <= endMS {
		batch, err := u.fetchKlines(ctx, startMS, endMS)
		if err != nil {
			return nil, err
		}
		if len(batch) == 0 {
			break
		}

		for _, kline := range batch {
			candle, err := parseKline(kline, u.config.Symbol, u.config.Interval)
			if err != nil {
				return nil, err
			}
			candles = append(candles, candle)
		}

		nextStart := batch[len(batch)-1].OpenTime + 1
		if nextStart <= startMS {
			return nil, errors.New("binance pagination did not advance")
		}
		startMS = nextStart
	}

	return candles, nil
}

func (u *Updater) fetchKlines(ctx context.Context, startMS int64, endMS int64) ([]binanceKline, error) {
	endpoint, err := url.Parse(u.config.BinanceURL)
	if err != nil {
		return nil, fmt.Errorf("parse BINANCE_KLINES_URL: %w", err)
	}

	values := endpoint.Query()
	values.Set("symbol", u.config.Symbol)
	values.Set("interval", u.config.Interval)
	values.Set("startTime", strconv.FormatInt(startMS, 10))
	values.Set("endTime", strconv.FormatInt(endMS, 10))
	values.Set("limit", strconv.Itoa(u.config.RequestLimit))
	endpoint.RawQuery = values.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create binance request: %w", err)
	}

	response, err := u.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("call binance klines api: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read binance response: %w", err)
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("binance klines api returned status %d: %s", response.StatusCode, string(body))
	}

	var raw [][]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decode binance response: %w", err)
	}

	klines := make([]binanceKline, 0, len(raw))
	for index, row := range raw {
		kline, err := decodeKline(row)
		if err != nil {
			return nil, fmt.Errorf("decode binance kline %d: %w", index, err)
		}
		klines = append(klines, kline)
	}

	return klines, nil
}

func ValidateCandles(candles []domain.Candle) error {
	seen := make(map[time.Time]struct{}, len(candles))
	var previous time.Time

	for index, candle := range candles {
		if _, ok := seen[candle.Timestamp]; ok {
			return fmt.Errorf("%w: index=%d timestamp=%s", ErrDuplicateTimestamp, index, candle.Timestamp.Format(time.RFC3339Nano))
		}

		if index > 0 && !candle.Timestamp.After(previous) {
			return fmt.Errorf(
				"%w: previous=%s current=%s",
				ErrOutOfOrderCandle,
				previous.Format(time.RFC3339Nano),
				candle.Timestamp.Format(time.RFC3339Nano),
			)
		}

		seen[candle.Timestamp] = struct{}{}
		previous = candle.Timestamp
	}

	return nil
}

type binanceKline struct {
	OpenTime int64
	Open     string
	High     string
	Low      string
	Close    string
	Volume   string
}

func decodeKline(row []json.RawMessage) (binanceKline, error) {
	if len(row) < 6 {
		return binanceKline{}, fmt.Errorf("expected at least 6 fields, got %d", len(row))
	}

	openTime, err := parseRawInt64(row[0])
	if err != nil {
		return binanceKline{}, fmt.Errorf("open time: %w", err)
	}

	open, err := parseRawString(row[1])
	if err != nil {
		return binanceKline{}, fmt.Errorf("open: %w", err)
	}
	high, err := parseRawString(row[2])
	if err != nil {
		return binanceKline{}, fmt.Errorf("high: %w", err)
	}
	low, err := parseRawString(row[3])
	if err != nil {
		return binanceKline{}, fmt.Errorf("low: %w", err)
	}
	closePrice, err := parseRawString(row[4])
	if err != nil {
		return binanceKline{}, fmt.Errorf("close: %w", err)
	}
	volume, err := parseRawString(row[5])
	if err != nil {
		return binanceKline{}, fmt.Errorf("volume: %w", err)
	}

	return binanceKline{
		OpenTime: openTime,
		Open:     open,
		High:     high,
		Low:      low,
		Close:    closePrice,
		Volume:   volume,
	}, nil
}

func parseKline(kline binanceKline, symbol string, timeframe string) (domain.Candle, error) {
	open, err := strconv.ParseFloat(kline.Open, 64)
	if err != nil {
		return domain.Candle{}, fmt.Errorf("parse open price: %w", err)
	}
	high, err := strconv.ParseFloat(kline.High, 64)
	if err != nil {
		return domain.Candle{}, fmt.Errorf("parse high price: %w", err)
	}
	low, err := strconv.ParseFloat(kline.Low, 64)
	if err != nil {
		return domain.Candle{}, fmt.Errorf("parse low price: %w", err)
	}
	closePrice, err := strconv.ParseFloat(kline.Close, 64)
	if err != nil {
		return domain.Candle{}, fmt.Errorf("parse close price: %w", err)
	}
	volume, err := strconv.ParseFloat(kline.Volume, 64)
	if err != nil {
		return domain.Candle{}, fmt.Errorf("parse volume: %w", err)
	}

	return domain.Candle{
		Timestamp: time.UnixMilli(kline.OpenTime).UTC(),
		Symbol:    symbol,
		Timeframe: timeframe,
		Open:      open,
		High:      high,
		Low:       low,
		Close:     closePrice,
		Volume:    volume,
	}, nil
}

func parseRawInt64(raw json.RawMessage) (int64, error) {
	var value int64
	if err := json.Unmarshal(raw, &value); err == nil {
		return value, nil
	}

	var text string
	if err := json.Unmarshal(raw, &text); err != nil {
		return 0, err
	}

	return strconv.ParseInt(text, 10, 64)
}

func parseRawString(raw json.RawMessage) (string, error) {
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return text, nil
	}

	var number json.Number
	if err := json.Unmarshal(raw, &number); err != nil {
		return "", err
	}

	return number.String(), nil
}

func validateConfig(config Config) error {
	if config.Symbol == "" {
		return errors.New("MARKET_DATA_SYMBOL is required")
	}
	if config.Interval == "" {
		return errors.New("MARKET_DATA_INTERVAL is required")
	}
	if config.StartDate.IsZero() {
		return errors.New("MARKET_DATA_START_DATE is required")
	}
	if config.BinanceURL == "" {
		return errors.New("BINANCE_KLINES_URL is required")
	}
	if config.RequestLimit <= 0 || config.RequestLimit > defaultRequestLimit {
		return fmt.Errorf("BINANCE_REQUEST_LIMIT must be between 1 and %d", defaultRequestLimit)
	}
	if _, err := intervalDuration(config.Interval); err != nil {
		return err
	}

	return nil
}

func intervalDuration(interval string) (time.Duration, error) {
	durations := map[string]time.Duration{
		"1m":  1 * time.Minute,
		"3m":  3 * time.Minute,
		"5m":  5 * time.Minute,
		"15m": 15 * time.Minute,
		"30m": 30 * time.Minute,
		"1h":  1 * time.Hour,
		"2h":  2 * time.Hour,
		"4h":  4 * time.Hour,
		"6h":  6 * time.Hour,
		"8h":  8 * time.Hour,
		"12h": 12 * time.Hour,
		"1d":  24 * time.Hour,
		"3d":  72 * time.Hour,
		"1w":  7 * 24 * time.Hour,
	}

	duration, ok := durations[interval]
	if !ok {
		return 0, fmt.Errorf("unsupported MARKET_DATA_INTERVAL: %s", interval)
	}

	return duration, nil
}

func lastClosedOpen(now time.Time, interval time.Duration) time.Time {
	return now.Truncate(interval).Add(-interval)
}

func toMilliseconds(value time.Time) int64 {
	return value.UTC().UnixMilli()
}
