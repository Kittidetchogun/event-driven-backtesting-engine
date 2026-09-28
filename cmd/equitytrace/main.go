package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"

	"event-driven-backtesting-engine/internal/backtest"
	"event-driven-backtesting-engine/internal/pipeline"
	"event-driven-backtesting-engine/internal/statistics"
	"event-driven-backtesting-engine/internal/storage/postgres"
)

type discardResultRepository struct{}

func (discardResultRepository) SaveBacktestResult(
	_ context.Context,
	_ statistics.BacktestResult,
) error {
	return nil
}

func main() {
	_ = godotenv.Load()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	symbolDefault := os.Getenv("MARKET_DATA_SYMBOL")
	if symbolDefault == "" {
		symbolDefault = "BTCUSDT"
	}

	timeframeDefault := os.Getenv("MARKET_DATA_INTERVAL")
	if timeframeDefault == "" {
		timeframeDefault = "1d"
	}

	startDefault := os.Getenv("BACKTEST_START_DATE")
	if startDefault == "" {
		startDefault = "2019-01-01"
	}

	endDefault := os.Getenv("BACKTEST_END_DATE")
	if endDefault == "" {
		endDefault = "2020-01-01"
	}

	symbol := flag.String("symbol", symbolDefault, "candle symbol")
	timeframe := flag.String("timeframe", timeframeDefault, "candle timeframe")
	startText := flag.String("start", startDefault, "inclusive start date, YYYY-MM-DD")
	endText := flag.String("end", endDefault, "exclusive end date, YYYY-MM-DD")
	output := flag.String("output", "data/go_equity_trace.csv", "CSV output path")
	flag.Parse()

	start := parseDate(*startText)
	end := parseDate(*endText)

	ctx := context.Background()
	pool, err := postgres.NewPool(ctx, databaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer pool.Close()

	repository := postgres.NewCandleRepository(pool)
	candlePipeline, err := pipeline.NewHistoricalCandlePipeline(
		ctx,
		repository,
		pipeline.CandleQuery{
			Symbol:    *symbol,
			Timeframe: *timeframe,
			Start:     start,
			End:       end,
		},
	)
	if err != nil {
		log.Fatalf("create candle pipeline: %v", err)
	}

	runner, err := backtest.NewRunner(
		backtest.RunnerConfig{
			RunID:          int(time.Now().UnixNano() % 1_000_000_000),
			StrategyName:   "EMA Cross",
			Symbol:         *symbol,
			Timeframe:      *timeframe,
			StartDate:      start,
			EndDate:        end,
			InitialCapital: 10_000,
		},
		candlePipeline,
		discardResultRepository{},
	)
	if err != nil {
		log.Fatalf("create runner: %v", err)
	}

	if _, err := runner.Run(); err != nil {
		log.Fatalf("run backtest: %v", err)
	}

	if len(runner.EquityTrace()) == 0 {
		log.Fatal("backtest produced no equity trace rows")
	}

	file, err := os.Create(*output)
	if err != nil {
		log.Fatalf("create output: %v", err)
	}
	defer file.Close()

	if err := runner.WriteEquityTraceCSV(file); err != nil {
		log.Fatalf("write equity trace: %v", err)
	}

	log.Printf("wrote %d equity trace rows to %s", len(runner.EquityTrace()), *output)
}

func parseDate(value string) time.Time {
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		log.Fatalf("invalid date %q: %v", value, err)
	}
	return parsed.UTC()
}
