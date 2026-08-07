package analyst

import (
	"math"
	"testing"

	"event-driven-backtesting-engine/internal/domain"
	"event-driven-backtesting-engine/internal/events"
	"event-driven-backtesting-engine/internal/statistics"
)

func newPerformanceWithEquities(equities ...float64) *Performance {
	engine := statistics.NewEngine()

	for _, equity := range equities {
		portfolio := domain.NewPortfolio(1, equity)
		portfolio.Equity = equity
		portfolio.Cash = equity

		if err := engine.Consume(events.NewPortfolioUpdatedEvent(portfolio)); err != nil {
			panic(err)
		}
	}

	return &Performance{stat: engine}
}

func TestPerformanceReturns(t *testing.T) {
	p := newPerformanceWithEquities(100, 120, 108)

	got := p.returns()
	expected := []float64{0.2, -0.1}

	if len(got) != len(expected) {
		t.Fatalf("expected %d returns, got %d", len(expected), len(got))
	}

	for i := range expected {
		if math.Abs(got[i]-expected[i]) > 1e-9 {
			t.Fatalf("expected return %.9f at %d, got %.9f", expected[i], i, got[i])
		}
	}
}

func TestPerformanceSharpeRatio(t *testing.T) {
	p := newPerformanceWithEquities(100, 120, 108)

	got := p.sharpeRatio()
	expected := 0.23570226039551584

	if math.Abs(got-expected) > 1e-12 {
		t.Fatalf("expected sharpe ratio %.15f, got %.15f", expected, got)
	}
}

func TestPerformanceMaxDrawdown(t *testing.T) {
	p := newPerformanceWithEquities(100, 120, 90, 150)

	got := p.maxDrawdown()
	expected := 0.25

	if math.Abs(got-expected) > 1e-12 {
		t.Fatalf("expected max drawdown %.2f, got %.15f", expected, got)
	}
}

func TestPerformanceEmptySnapshots(t *testing.T) {
	p := &Performance{stat: statistics.NewEngine()}

	if got := p.returns(); len(got) != 0 {
		t.Fatalf("expected no returns, got %d", len(got))
	}

	if got := p.sharpeRatio(); got != 0 {
		t.Fatalf("expected sharpe ratio 0, got %.2f", got)
	}

	if got := p.maxDrawdown(); got != 0 {
		t.Fatalf("expected max drawdown 0, got %.2f", got)
	}
}
