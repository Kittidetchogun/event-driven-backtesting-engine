package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"

	"event-driven-backtesting-engine/internal/backtest"
	"event-driven-backtesting-engine/internal/pipeline"
	"event-driven-backtesting-engine/internal/storage/postgres"
)

func main() {
	loadedEnv := false
	for _, envPath := range []string{
		".env",
		filepath.Join("..", ".env"),
		filepath.Join("..", "..", ".env"),
	} {
		if err := godotenv.Load(envPath); err == nil {
			loadedEnv = true
			break
		}
	}
	if !loadedEnv {
		log.Println("warning: .env file not loaded")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	symbol := os.Getenv("MARKET_DATA_SYMBOL")
	if symbol == "" {
		log.Fatal("MARKET_DATA_SYMBOL is required")
	}

	timeframe := os.Getenv("MARKET_DATA_INTERVAL")
	if timeframe == "" {
		log.Fatal("MARKET_DATA_INTERVAL is required")
	}

	backtestStartStr := "2023-01-01"
	backtestEndStr := "2023-12-31"
	if backtestStartStr == "" || backtestEndStr == "" {
		log.Fatal("BACKTEST_START_DATE and BACKTEST_END_DATE are required")
	}

	backtestStart, err := time.Parse("2006-01-02", backtestStartStr)
	if err != nil {
		log.Fatalf("invalid BACKTEST_START_DATE %q: %v", backtestStartStr, err)
	}

	backtestEnd, err := time.Parse("2006-01-02", backtestEndStr)
	if err != nil {
		log.Fatalf("invalid BACKTEST_END_DATE %q: %v", backtestEndStr, err)
	}

	backtestStart = backtestStart.UTC()
	backtestEnd = backtestEnd.Add(24*time.Hour - time.Nanosecond).UTC()

	if !backtestStart.Before(backtestEnd) {
		log.Fatal("BACKTEST_START_DATE must be before BACKTEST_END_DATE")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := postgres.NewPool(ctx, databaseURL)
	if err != nil {
		log.Fatalf("connect to PostgreSQL: %v", err)
	}
	defer pool.Close()

	candleRepository := postgres.NewCandleRepository(pool)
	backtestRepository := postgres.NewBacktestRepository(pool)

	candles, err := candleRepository.GetCandles(
		ctx,
		symbol,
		timeframe,
		backtestStart,
		backtestEnd,
	)
	if err != nil {
		log.Fatalf("load candles from PostgreSQL: %v", err)
	}

	if len(candles) == 0 {
		log.Fatalf("no candles found for %s %s in the selected range", symbol, timeframe)
	}

	testStart := candles[0].Timestamp
	testEnd := candles[len(candles)-1].Timestamp

	candlePipeline, err := pipeline.NewHistoricalCandlePipeline(
		ctx,
		candleRepository,
		pipeline.CandleQuery{
			Symbol:    symbol,
			Timeframe: timeframe,
			Start:     testStart,
			End:       testEnd,
		},
	)
	if err != nil {
		log.Fatalf("create historical candle pipeline: %v", err)
	}

	runID := int(time.Now().UnixNano() % 1_000_000_000)

	runner, err := backtest.NewRunner(
		backtest.RunnerConfig{
			RunID:          runID,
			StrategyName:   "EMA Cross",
			Symbol:         symbol,
			Timeframe:      timeframe,
			StartDate:      testStart,
			EndDate:        testEnd,
			InitialCapital: 10_000,
		},
		candlePipeline,
		backtestRepository,
	)
	if err != nil {
		log.Fatalf("create runner: %v", err)
	}

	result, err := runner.Run()
	if err != nil {
		log.Fatalf("run backtest: %v", err)
	}

	fmt.Println("Backtest completed successfully")
	fmt.Printf("RunID: %d\n", result.RunID)
	fmt.Printf("Strategy: %s\n", result.StrategyName)
	fmt.Printf("Symbol: %s\n", result.Symbol)
	fmt.Printf("Timeframe: %s\n", result.Timeframe)
	fmt.Printf("StartDate: %s\n", result.StartDate.Format(time.RFC3339))
	fmt.Printf("EndDate: %s\n", result.EndDate.Format(time.RFC3339))
	fmt.Printf("TotalReturn: %.8f\n", result.TotalReturn)
	fmt.Printf("WinRate: %.8f\n", result.WinRate)
	fmt.Printf("SharpeRatio: %.8f\n", result.SharpeRatio)
	fmt.Printf("MaxDrawdown: %.8f\n", result.MaxDrawdown)
}
