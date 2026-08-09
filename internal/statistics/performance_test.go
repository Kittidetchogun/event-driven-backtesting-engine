package statistics

import (
	"math"
	"testing"
	"time"

	"event-driven-backtesting-engine/internal/domain"
	"event-driven-backtesting-engine/internal/events"
)

func TestEnginePerformance(t *testing.T) {
	engine := NewEngine()

	now := time.Unix(1000, 0)

	// Portfolio snapshots
	portfolio1 := domain.NewPortfolio(1, 100)
	portfolio1.Equity = 100
	portfolio1.Cash = 100
	portfolio1.UpdatedAt = now

	portfolio2 := domain.NewPortfolio(1, 120)
	portfolio2.Equity = 120
	portfolio2.Cash = 120
	portfolio2.UpdatedAt = now.Add(time.Minute)

	portfolio3 := domain.NewPortfolio(1, 108)
	portfolio3.Equity = 108
	portfolio3.Cash = 108
	portfolio3.UpdatedAt = now.Add(2 * time.Minute)

	if err := engine.Consume(
		events.NewPortfolioUpdatedEvent(portfolio1),
	); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := engine.Consume(
		events.NewPortfolioUpdatedEvent(portfolio2),
	); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := engine.Consume(
		events.NewPortfolioUpdatedEvent(portfolio3),
	); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Buy
	buy := domain.NewTrade(
		1,
		1,
		11,
		"BTCUSDT",
		domain.BuyOrder,
		1,
		100,
		1,
		now,
	)

	// Sell
	sell := domain.NewTrade(
		2,
		1,
		12,
		"BTCUSDT",
		domain.SellOrder,
		1,
		110,
		1,
		now.Add(time.Minute),
	)

	if err := engine.Consume(
		events.NewTradeExecutedEvent(buy),
	); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := engine.Consume(
		events.NewTradeExecutedEvent(sell),
	); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	performance := engine.Performance()

	// Trade profit:
	// (110 - 100) * 1 - 1 - 1 = 8
	if math.Abs(performance.NetProfit-8) > 1e-9 {
		t.Fatalf(
			"expected net profit 8, got %.2f",
			performance.NetProfit,
		)
	}

	if performance.TotalTrades != 1 {
		t.Fatalf(
			"expected 1 total trade, got %d",
			performance.TotalTrades,
		)
	}

	if performance.WinningTrades != 1 {
		t.Fatalf(
			"expected 1 winning trade, got %d",
			performance.WinningTrades,
		)
	}

	if performance.LosingTrades != 0 {
		t.Fatalf(
			"expected 0 losing trades, got %d",
			performance.LosingTrades,
		)
	}

	if performance.WinRate != 1 {
		t.Fatalf(
			"expected win rate 1, got %.2f",
			performance.WinRate,
		)
	}

	if math.Abs(performance.AverageWinningTrade-8) > 1e-9 {
		t.Fatalf(
			"expected average winning trade 8, got %.2f",
			performance.AverageWinningTrade,
		)
	}
}

func TestEnginePerformance_Empty(t *testing.T) {
	engine := NewEngine()

	performance := engine.Performance()

	if performance.NetProfit != 0 {
		t.Fatalf("expected net profit 0, got %.2f", performance.NetProfit)
	}

	if performance.WinRate != 0 {
		t.Fatalf("expected win rate 0, got %.2f", performance.WinRate)
	}

	if performance.SharpeRatio != 0 {
		t.Fatalf(
			"expected sharpe ratio 0, got %.2f",
			performance.SharpeRatio,
		)
	}

	if performance.MaxDrawdown != 0 {
		t.Fatalf(
			"expected max drawdown 0, got %.2f",
			performance.MaxDrawdown,
		)
	}

	if performance.TotalTrades != 0 {
		t.Fatalf(
			"expected total trades 0, got %d",
			performance.TotalTrades,
		)
	}
}
