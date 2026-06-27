package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"event-driven-backtesting-engine/internal/storage/postgres"
)

func main() {
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
	candles, err := repository.GetCandles(
		ctx,
		"BTCUSDT",
		"1d",
		time.Unix(0, 0).UTC(),
		time.Now().UTC(),
	)
	if err != nil {
		log.Fatalf("get candles: %v", err)
	}

	fmt.Printf("BTCUSDT 1d candles loaded: %d\n", len(candles))
}
