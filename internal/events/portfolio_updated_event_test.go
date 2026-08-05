package events

import (
	"testing"

	"event-driven-backtesting-engine/internal/domain"
)

func TestNewPortfolioUpdatedEvent(t *testing.T) {

	portfolio := domain.NewPortfolio(
		1,
		100000,
	)

	event := NewPortfolioUpdatedEvent(portfolio)

	if event.Type() != PortfolioUpdatedEventType {
		t.Fatalf(
			"expected %s, got %s",
			PortfolioUpdatedEventType,
			event.Type(),
		)
	}

	if event.Portfolio.RunID != portfolio.RunID {
		t.Fatalf(
			"expected RunID %d, got %d",
			portfolio.RunID,
			event.Portfolio.RunID,
		)
	}

	if event.Portfolio.Cash != portfolio.Cash {
		t.Fatalf(
			"expected Cash %.2f, got %.2f",
			portfolio.Cash,
			event.Portfolio.Cash,
		)
	}

	if !event.Timestamp().Equal(portfolio.UpdatedAt) {
		t.Fatalf(
			"expected timestamp %v, got %v",
			portfolio.UpdatedAt,
			event.Timestamp(),
		)
	}

	if event.Portfolio.Equity != portfolio.Equity {
		t.Fatalf("expected Equity %.2f, got %.2f",
			portfolio.Equity,
			event.Portfolio.Equity,
		)
	}

	if event.Portfolio.PositionValue != portfolio.PositionValue {
		t.Fatal("unexpected position value")
	}
}
