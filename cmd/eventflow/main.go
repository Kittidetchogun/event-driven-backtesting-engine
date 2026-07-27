package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"

	"event-driven-backtesting-engine/internal/events"
	"event-driven-backtesting-engine/internal/pipeline"
	"event-driven-backtesting-engine/internal/storage/postgres"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()

	fmt.Println("Connecting Database...")

	pool, err := postgres.NewPool(ctx, databaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer pool.Close()

	fmt.Println("Database Connected")

	repository := postgres.NewCandleRepository(pool)

	fmt.Println("Loading Historical Data...")

	p, err := pipeline.NewHistoricalCandlePipeline(
		ctx,
		repository,
		pipeline.CandleQuery{
			Symbol:    "BTCUSDT",
			Timeframe: "1d",
			Start:     time.Unix(0, 0).UTC(),
			End:       time.Now().UTC(),
		},
	)
	if err != nil {
		log.Fatalf("create pipeline: %v", err)
	}

	queue := events.NewEventQueue()
	dispatcher := events.NewEventDispatcher()
	printConsumer := events.NewPrintConsumer()

	dispatcher.Register(
		events.CandleReceivedEventType,
		printConsumer.Consume,
	)

	fmt.Println("Starting Event Flow...")

	for {
		candle, ok, err := p.Next()

		if err != nil {
			log.Fatalf("pipeline error: %v", err)
		}

		if !ok {
			fmt.Println("End of Historical Data")
			break
		}

		// Create Event
		event := events.NewCandleReceivedEvent(candle)

		// Push into Queue
		queue.Push(event)

		// Pop from Queue
		queuedEvent, ok := queue.Pop()
		if !ok {
			continue
		}

		// Dispatch
		if err := dispatcher.Dispatch(queuedEvent); err != nil {
			log.Printf("dispatch error: %v", err)
		}
	}

	fmt.Println("Event Flow Completed")
}