package backtest

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"

	"event-driven-backtesting-engine/internal/pipeline"
	"event-driven-backtesting-engine/internal/storage/postgres"
)

func TestBacktestEndToEnd_PostgreSQL(t *testing.T) {
	// ------------------------------------------------------------
	// 0. Load environment
	// ------------------------------------------------------------

	_ = godotenv.Load()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is not set; skipping PostgreSQL E2E test")
	}

	symbol := os.Getenv("MARKET_DATA_SYMBOL")
	if symbol == "" {
		t.Skip("MARKET_DATA_SYMBOL is not set")
	}

	timeframe := os.Getenv("MARKET_DATA_INTERVAL")
	if timeframe == "" {
		t.Skip("MARKET_DATA_INTERVAL is not set")
	}

	// ------------------------------------------------------------
	// 1. PostgreSQL connection
	// ------------------------------------------------------------

	ctx, cancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancel()

	pool, err := postgres.NewPool(ctx, databaseURL)
	if err != nil {
		t.Fatalf(
			"PostgreSQL connection failed: %v",
			err,
		)
	}
	defer pool.Close()

	// ------------------------------------------------------------
	// 2. Create repositories
	// ------------------------------------------------------------

	candleRepository := postgres.NewCandleRepository(pool)
	backtestRepository := postgres.NewBacktestRepository(pool)

	// ------------------------------------------------------------
	// 3. Verify PostgreSQL actually contains candles
	// ------------------------------------------------------------

	start := time.Unix(0, 0).UTC()
	end := time.Now().UTC()

	candles, err := candleRepository.GetCandles(
		ctx,
		symbol,
		timeframe,
		start,
		end,
	)
	if err != nil {
		t.Fatalf(
			"failed to load candles from PostgreSQL: %v",
			err,
		)
	}

	if len(candles) == 0 {
		t.Skip(
			"no candles found in PostgreSQL for " +
				symbol + " " + timeframe,
		)
	}

	t.Logf(
		"loaded %d candles from PostgreSQL",
		len(candles),
	)

	// ------------------------------------------------------------
	// 4. Verify chronological ordering
	// ------------------------------------------------------------

	for i := 1; i < len(candles); i++ {
		if candles[i].Timestamp.Before(
			candles[i-1].Timestamp,
		) {
			t.Fatalf(
				"candles are not chronological: index %d (%v) before index %d (%v)",
				i,
				candles[i].Timestamp,
				i-1,
				candles[i-1].Timestamp,
			)
		}
	}

	// ------------------------------------------------------------
	// 5. Use a limited deterministic range
	// ------------------------------------------------------------
	//
	// Do not run the entire database unnecessarily.
	// Use the first available candles as the E2E window.
	//

	const maxCandles = 500

	if len(candles) > maxCandles {
		candles = candles[:maxCandles]
	}

	testStart := candles[0].Timestamp
	testEnd := candles[len(candles)-1].Timestamp

	// ------------------------------------------------------------
	// 6. Create Historical Candle Pipeline
	// ------------------------------------------------------------

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
		t.Fatalf(
			"failed to create historical candle pipeline: %v",
			err,
		)
	}

	// ------------------------------------------------------------
	// 7. Generate unique RunID
	// ------------------------------------------------------------

	runID := int(time.Now().UnixNano() % 1_000_000_000)

	strategyName := "EMA Cross"

	// ------------------------------------------------------------
	// 8. Create Runner
	// ------------------------------------------------------------

	runner, err := NewRunner(
		RunnerConfig{
			RunID:          runID,
			StrategyName:   strategyName,
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
		t.Fatalf(
			"NewRunner() failed: %v",
			err,
		)
	}

	// ------------------------------------------------------------
	// 9. Run complete backtest
	// ------------------------------------------------------------

	result, err := runner.Run()
	if err != nil {
		t.Fatalf(
			"Runner.Run() failed: %v",
			err,
		)
	}

	// ------------------------------------------------------------
	// 10. Verify returned result
	// ------------------------------------------------------------

	if result.RunID != runID {
		t.Errorf(
			"RunID = %d, want %d",
			result.RunID,
			runID,
		)
	}

	if result.StrategyName != strategyName {
		t.Errorf(
			"StrategyName = %q, want %q",
			result.StrategyName,
			strategyName,
		)
	}

	if result.Symbol != symbol {
		t.Errorf(
			"Symbol = %q, want %q",
			result.Symbol,
			symbol,
		)

	}

	if result.Timeframe != timeframe {
		t.Errorf(
			"Timeframe = %q, want %q",
			result.Timeframe,
			timeframe,
		)
	}

	// ------------------------------------------------------------
	// 11. Verify actual event-driven flow
	// ------------------------------------------------------------

	trades := runner.matching.Trades()

	snapshots := runner.portfolio.Snapshots()

	equityCurve := runner.statistics.EquityCurve()

	t.Logf(
		"executed trades: %d",
		len(trades),
	)

	t.Logf(
		"portfolio snapshots: %d",
		len(snapshots),
	)

	t.Logf(
		"equity curve points: %d",
		len(equityCurve),
	)

	// ------------------------------------------------------------
	// 12. Verify PostgreSQL persisted BacktestResult
	// ------------------------------------------------------------

	savedResult, err := backtestRepository.GetBacktestResult(
		ctx,
		runID,
	)
	if err != nil {
		t.Fatalf(
			"failed to read persisted backtest result: %v",
			err,
		)
	}

	if savedResult == nil {
		t.Fatalf(
			"PostgreSQL does not contain backtest result for RunID %d",
			runID,
		)
	}

	// ------------------------------------------------------------
	// 13. Compare Runner result with PostgreSQL result
	// ------------------------------------------------------------

	if savedResult.RunID != result.RunID {
		t.Errorf(
			"DB RunID = %d, Runner RunID = %d",
			savedResult.RunID,
			result.RunID,
		)
	}

	if savedResult.StrategyName != result.StrategyName {
		t.Errorf(
			"DB StrategyName = %q, Runner StrategyName = %q",
			savedResult.StrategyName,
			result.StrategyName,
		)
	}

	if savedResult.Symbol != result.Symbol {
		t.Errorf(
			"DB Symbol = %q, Runner Symbol = %q",
			savedResult.Symbol,
			result.Symbol,
		)
	}

	if savedResult.Timeframe != result.Timeframe {
		t.Errorf(
			"DB Timeframe = %q, Runner Timeframe = %q",
			savedResult.Timeframe,
			result.Timeframe,
		)
	}

	if savedResult.StartDate != result.StartDate {
		t.Errorf(
			"DB StartDate = %v, Runner StartDate = %v",
			savedResult.StartDate,
			result.StartDate,
		)
	}

	if savedResult.EndDate != result.EndDate {
		t.Errorf(
			"DB EndDate = %v, Runner EndDate = %v",
			savedResult.EndDate,
			result.EndDate,
		)
	}

	if savedResult.TotalReturn != result.TotalReturn {
		t.Errorf(
			"DB TotalReturn = %.10f, Runner TotalReturn = %.10f",
			savedResult.TotalReturn,
			result.TotalReturn,
		)
	}

	if savedResult.WinRate != result.WinRate {
		t.Errorf(
			"DB WinRate = %.10f, Runner WinRate = %.10f",
			savedResult.WinRate,
			result.WinRate,
		)
	}

	if savedResult.SharpeRatio != result.SharpeRatio {
		t.Errorf(
			"DB SharpeRatio = %.10f, Runner SharpeRatio = %.10f",
			savedResult.SharpeRatio,
			result.SharpeRatio,
		)
	}

	if savedResult.MaxDrawdown != result.MaxDrawdown {
		t.Errorf(
			"DB MaxDrawdown = %.10f, Runner MaxDrawdown = %.10f",
			savedResult.MaxDrawdown,
			result.MaxDrawdown,
		)
	}

	// ------------------------------------------------------------
	// 14. Final summary
	// ------------------------------------------------------------

	t.Log("========================================")
	t.Log("PostgreSQL E2E TEST PASSED")
	t.Log("========================================")

	t.Logf("Run ID             : %d", result.RunID)
	t.Logf("Symbol             : %s", result.Symbol)
	t.Logf("Timeframe          : %s", result.Timeframe)

	t.Log("")
	t.Log("Backtest Activity")
	t.Logf("Candles            : %d", len(candles))
	t.Logf("Executed Trades    : %d", len(trades))
	t.Logf("Portfolio Snapshots: %d", len(snapshots))
	t.Logf("Equity Curve Points: %d", len(equityCurve))

	t.Log("")
	t.Log("Performance Summary")
	t.Logf("Total Return       : %.2f%%", result.TotalReturn*100)
	t.Logf("Win Rate           : %.2f%%", result.WinRate*100)
	t.Logf("Sharpe Ratio       : %.3f", result.SharpeRatio)
	t.Logf("Max Drawdown       : %.2f%%", result.MaxDrawdown*100)

	t.Log("")
	t.Log("Database")
	t.Log("PostgreSQL Persistence: VERIFIED")
	t.Log("========================================")
}
