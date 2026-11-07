package main

import (
	"context"
	"flag"
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

	// CLI flag overrides (applied after .env defaults are loaded)
	intervalFlag := flag.String("interval", "", "market data timeframe/interval (e.g. 5m, 1h, 1d)")
	startFlag := flag.String("start", "", "start date YYYY-MM-DD (overrides .env MARKET_DATA_START_DATE)")
	endFlag := flag.String("end", "", "exclusive end date YYYY-MM-DD (required with -backfill)")
	symbolFlag := flag.String("symbol", "", "trading symbol (e.g. BTCUSDT)")
	backfillFlag := flag.Bool("backfill", false, "download only missing candles in the explicit [start,end) range")
	flag.Parse()

	if *intervalFlag != "" {
		config.Interval = *intervalFlag
	}
	if *startFlag != "" {
		parsedStart, parseErr := time.Parse("2006-01-02", *startFlag)
		if parseErr != nil {
			logger.Printf("Market data sync failed: invalid -start date %q: %v", *startFlag, parseErr)
			os.Exit(1)
		}
		config.StartDate = parsedStart.UTC()
	}
	if *symbolFlag != "" {
		config.Symbol = *symbolFlag
	}

	// Re-validate after flag overrides (catches e.g. unsupported interval)
	if err := marketdata.ValidateConfig(config); err != nil {
		logger.Printf("Market data sync failed: invalid config: %v", err)
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

	var endDate time.Time
	if *backfillFlag {
		if *endFlag == "" {
			logger.Println("Market data sync failed: -end is required with -backfill")
			os.Exit(1)
		}
		endDate, err = time.Parse("2006-01-02", *endFlag)
		if err != nil {
			logger.Printf("Market data sync failed: invalid -end date %q: %v", *endFlag, err)
			os.Exit(1)
		}
		endDate = endDate.UTC()
		if !config.StartDate.Before(endDate) {
			logger.Println("Market data sync failed: -start must be before -end for backfill")
			os.Exit(1)
		}
	}

	if *backfillFlag {
		err = updater.Backfill(ctx, config.StartDate, endDate)
	} else {
		err = updater.Sync(ctx)
	}
	if err != nil {
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
