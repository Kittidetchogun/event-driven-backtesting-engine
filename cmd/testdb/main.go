package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"

	"event-driven-backtesting-engine/internal/pipeline"
	"event-driven-backtesting-engine/internal/storage/postgres"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()

	pool, err := postgres.NewPool(ctx, databaseURL)
	if err != nil {
		log.Fatalf("connect to PostgreSQL: %v", err)
	}
	defer pool.Close()

	repository := postgres.NewCandleRepository(pool)
	symbol := os.Getenv("MARKET_DATA_SYMBOL")
	interval := os.Getenv("MARKET_DATA_INTERVAL")
	candlePipeline, err := pipeline.NewHistoricalCandlePipeline(
		ctx,
		repository,
		pipeline.CandleQuery{
			Symbol:    symbol,
			Timeframe: interval,
			Start:     time.Unix(0, 0).UTC(),
			End:       time.Now().UTC(),
		},
	)
	if err != nil {
		log.Fatalf("create historical candle pipeline: %v", err)
	}

	for {
		candle, ok, err := candlePipeline.Next()
		if err != nil {
			log.Fatalf("read candle from pipeline: %v", err)
		}
		if !ok {
			fmt.Println("End of Historical Data")
			return
		}

		fmt.Printf(
			"%s %s open=%.8f high=%.8f low=%.8f close=%.8f volume=%.8f\n",
			candle.Timestamp.Format(time.RFC3339),
			candle.Symbol,
			candle.Open,
			candle.High,
			candle.Low,
			candle.Close,
			candle.Volume,
		)
	}
}
