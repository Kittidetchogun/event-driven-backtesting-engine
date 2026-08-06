package portfolio

import (
	"testing"
	"time"

	"event-driven-backtesting-engine/internal/domain"
)

func TestNewPortfolioSnapshot(t *testing.T) {

	portfolio := domain.NewPortfolio(
		1,
		100000,
	)

	now := time.Now()

	portfolio.Cash = 95000
	portfolio.Equity = 110000
	portfolio.PositionValue = 15000
	portfolio.UnrealizedPnL = 8000
	portfolio.RealizedPnL = 2000
	portfolio.UpdatedAt = now

	snapshot := NewPortfolioSnapshot(portfolio)

	if snapshot.Time != now {
		t.Fatalf(
			"expected time %v, got %v",
			now,
			snapshot.Time,
		)
	}

	if snapshot.Cash != portfolio.Cash {
		t.Fatalf(
			"expected cash %.2f, got %.2f",
			portfolio.Cash,
			snapshot.Cash,
		)
	}

	if snapshot.Equity != portfolio.Equity {
		t.Fatalf(
			"expected equity %.2f, got %.2f",
			portfolio.Equity,
			snapshot.Equity,
		)
	}

	if snapshot.PositionValue != portfolio.PositionValue {
		t.Fatalf(
			"expected position value %.2f, got %.2f",
			portfolio.PositionValue,
			snapshot.PositionValue,
		)
	}

	if snapshot.UnrealizedPnL != portfolio.UnrealizedPnL {
		t.Fatalf(
			"expected unrealized pnl %.2f, got %.2f",
			portfolio.UnrealizedPnL,
			snapshot.UnrealizedPnL,
		)
	}

	if snapshot.RealizedPnL != portfolio.RealizedPnL {
		t.Fatalf(
			"expected realized pnl %.2f, got %.2f",
			portfolio.RealizedPnL,
			snapshot.RealizedPnL,
		)
	}
}
