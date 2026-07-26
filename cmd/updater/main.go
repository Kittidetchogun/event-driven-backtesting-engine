package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"event-driven-backtesting-engine/internal/marketdata"
	"event-driven-backtesting-engine/internal/storage/postgres"
)

func main() {
	logger := log.New(os.Stdout, "", 0)
	ctx := context.Background()

	config, err := marketdata.LoadConfigFromEnv()
	if err != nil {
		logger.Printf("Market data sync failed: load config: %v", err)
		os.Exit(1)
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		logger.Println("Market data sync failed: DATABASE_URL is required")
		os.Exit(1)
	}

	pool, err := postgres.NewPool(ctx, databaseURL)
	if err != nil {
		logger.Printf("Market data sync failed: connect to PostgreSQL/TimescaleDB: %v", err)
		os.Exit(1)
	}
	defer pool.Close()

	repository := postgres.NewCandleRepository(pool)

	before, beforeExists, err := repository.GetLatestTimestamp(ctx, config.Symbol, config.Interval)
	if err != nil {
		logger.Printf("Market data sync failed: get latest timestamp before sync: %v", err)
		os.Exit(1)
	}
	logger.Printf("Latest timestamp before sync: %s", formatLatestTimestamp(before, beforeExists))

	updater, err := marketdata.NewUpdater(repository, config, logger)
	if err != nil {
		logger.Printf("Market data sync failed: create updater: %v", err)
		os.Exit(1)
	}

	if err := updater.Sync(ctx); err != nil {
		logger.Printf("Market data sync failed: %v", err)
		os.Exit(1)
	}

	after, afterExists, err := repository.GetLatestTimestamp(ctx, config.Symbol, config.Interval)
	if err != nil {
		logger.Printf("Market data sync failed: get latest timestamp after sync: %v", err)
		os.Exit(1)
	}

	result := updater.LastResult()
	logger.Println("Sync summary:")
	logger.Printf("Downloaded candles: %d", result.Downloaded)
	logger.Printf("Inserted candles: %d", result.Inserted)
	logger.Printf("Latest timestamp after sync: %s", formatLatestTimestamp(after, afterExists))
	logger.Println("Market data sync completed successfully.")
}

func formatLatestTimestamp(timestamp time.Time, exists bool) string {
	if !exists {
		return "none"
	}

	return fmt.Sprintf("%s (%s)", timestamp.Format(time.RFC3339), timestamp.Format("2006-01-02"))
}
