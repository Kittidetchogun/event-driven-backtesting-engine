package statistics

import (
	"errors"
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

	performance, err := engine.Performance()
	if err != nil {
		t.Fatal(err)
	}

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

	performance, err := engine.Performance()
	if err != nil {
		t.Fatal(err)
	}

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

func TestSharpeRatioInsufficientReturns(t *testing.T) {
	if got := SharpeRatio(nil); got != 0 || math.IsNaN(got) || math.IsInf(got, 0) {
		t.Fatalf("SharpeRatio(nil) = %v, want finite zero", got)
	}
	if got := SharpeRatio([]float64{100, 110}); got != 0 || math.IsNaN(got) || math.IsInf(got, 0) {
		t.Fatalf("SharpeRatio(one return) = %v, want finite zero", got)
	}
}

func TestSharpeRatioValidSampleCalculationUnchanged(t *testing.T) {
	got := SharpeRatio([]float64{100, 110, 104.5, 109.725})
	want := Mean([]float64{0.10, -0.05, 0.05}) / StdDev([]float64{0.10, -0.05, 0.05})
	if math.Abs(got-want) > 1e-12 {
		t.Fatalf("SharpeRatio() = %.15f, want %.15f", got, want)
	}
}

func TestEnginePerformanceSeparatesBreakevenTrades(t *testing.T) {
	engine := NewEngine()
	now := time.Unix(1000, 0)
	for index, sellPrice := range []float64{120, 110, 95, 85, 100} {
		orderID := index*2 + 1
		buy := domain.NewTrade(orderID, 1, orderID, "BTCUSDT", domain.BuyOrder, 1, 100, 0, now)
		sell := domain.NewTrade(orderID+1, 1, orderID+1, "BTCUSDT", domain.SellOrder, 1, sellPrice, 0, now.Add(time.Minute))
		if err := engine.Consume(events.NewTradeExecutedEvent(buy)); err != nil {
			t.Fatal(err)
		}
		if err := engine.Consume(events.NewTradeExecutedEvent(sell)); err != nil {
			t.Fatal(err)
		}
	}

	performance, err := engine.Performance()
	if err != nil {
		t.Fatal(err)
	}
	if performance.TotalTrades != 5 || performance.WinningTrades != 2 || performance.LosingTrades != 2 || performance.BreakevenTrades != 1 {
		t.Fatalf("performance counts = total %d wins %d losses %d breakeven %d, want 5/2/2/1", performance.TotalTrades, performance.WinningTrades, performance.LosingTrades, performance.BreakevenTrades)
	}
}

func TestPerformanceRejectsIncompleteEquityCurve(t *testing.T) {
	for _, invalid := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		engine := NewEngine()
		for index, equity := range []float64{100, 110, 120, 130} {
			portfolio := domain.NewPortfolio(1, 1000)
			portfolio.Equity = equity
			portfolio.Cash = equity
			portfolio.UpdatedAt = time.Unix(int64(index), 0)
			if err := engine.Consume(events.NewPortfolioUpdatedEvent(portfolio)); err != nil {
				t.Fatal(err)
			}
		}
		engine.Snapshots()[3].Equity = invalid
		if _, err := engine.Performance(); !errors.Is(err, ErrInvalidEquityCurve) {
			t.Fatalf("Performance() error = %v, want ErrInvalidEquityCurve for %v", err, invalid)
		}
	}
}

func TestPerformanceRejectsZeroPreviousEquityAfterValidReturns(t *testing.T) {
	engine := NewEngine()
	for index, equity := range []float64{100, 110, 0, 120} {
		portfolio := domain.NewPortfolio(1, 1000)
		portfolio.Equity = equity
		portfolio.Cash = equity
		portfolio.UpdatedAt = time.Unix(int64(index), 0)
		if err := engine.Consume(events.NewPortfolioUpdatedEvent(portfolio)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := engine.Performance(); !errors.Is(err, ErrInvalidEquityCurve) {
		t.Fatalf("Performance() error = %v, want ErrInvalidEquityCurve", err)
	}
}

func TestPerformanceAcceptsNegativeEquity(t *testing.T) {
	engine := NewEngine()
	for index, equity := range []float64{100, -50, -25} {
		portfolio := domain.NewPortfolio(1, 1000)
		portfolio.Equity = equity
		portfolio.Cash = equity
		portfolio.UpdatedAt = time.Unix(int64(index), 0)
		if err := engine.Consume(events.NewPortfolioUpdatedEvent(portfolio)); err != nil {
			t.Fatal(err)
		}
	}
	performance, err := engine.Performance()
	if err != nil {
		t.Fatal(err)
	}
	if math.IsNaN(performance.SharpeRatio) || math.IsInf(performance.SharpeRatio, 0) || performance.MaxDrawdown != 1 {
		t.Fatalf("performance = %+v, want finite Sharpe and drawdown 1", performance)
	}
}
