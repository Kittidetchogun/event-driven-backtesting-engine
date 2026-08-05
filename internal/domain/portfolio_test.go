package domain

import "testing"

const initialCapital = 100000.0

func TestNewPortfolio(t *testing.T) {

	portfolio := NewPortfolio(1, initialCapital)

	if portfolio.RunID != 1 {
		t.Fatalf("expected RunID 1, got %d", portfolio.RunID)
	}

	if portfolio.InitCapital != initialCapital {
		t.Fatalf(
			"expected initial capital %.2f, got %.2f",
			initialCapital,
			portfolio.InitCapital,
		)
	}

	if portfolio.Cash != initialCapital {
		t.Fatalf(
			"expected cash %.2f, got %.2f",
			initialCapital,
			portfolio.Cash,
		)
	}

	if portfolio.Equity != initialCapital {
		t.Fatalf(
			"expected equity %.2f, got %.2f",
			initialCapital,
			portfolio.Equity,
		)
	}

	if portfolio.PositionValue != 0 {
		t.Fatalf(
			"expected position value 0, got %.2f",
			portfolio.PositionValue,
		)
	}

	if portfolio.UnrealizedPnL != 0 {
		t.Fatalf(
			"expected unrealized pnl 0, got %.2f",
			portfolio.UnrealizedPnL,
		)
	}

	if portfolio.RealizedPnL != 0 {
		t.Fatalf(
			"expected realized pnl 0, got %.2f",
			portfolio.RealizedPnL,
		)
	}

	if portfolio.UpdatedAt.IsZero() {
		t.Fatal("expected UpdatedAt to be set")
	}
}

func TestValidatePortfolio_Valid(t *testing.T) {

	portfolio := NewPortfolio(1, initialCapital)

	if err := ValidatePortfolio(portfolio); err != nil {
		t.Fatalf("expected valid portfolio, got %v", err)
	}
}

func TestValidatePortfolio_InvalidInitialCapital(t *testing.T) {

	portfolio := NewPortfolio(1, initialCapital)
	portfolio.InitCapital = -1

	if err := ValidatePortfolio(portfolio); err != ErrInvalidInitialCapital {
		t.Fatalf("expected %v, got %v", ErrInvalidInitialCapital, err)
	}
}

func TestValidatePortfolio_InvalidCash(t *testing.T) {

	portfolio := NewPortfolio(1, initialCapital)
	portfolio.Cash = -1

	if err := ValidatePortfolio(portfolio); err != ErrInvalidCash {
		t.Fatalf("expected %v, got %v", ErrInvalidCash, err)
	}
}

func TestValidatePortfolio_InvalidEquity(t *testing.T) {

	portfolio := NewPortfolio(1, initialCapital)
	portfolio.Equity = -1

	if err := ValidatePortfolio(portfolio); err != ErrInvalidEquity {
		t.Fatalf("expected %v, got %v", ErrInvalidEquity, err)
	}
}

func TestValidatePortfolio_InvalidPositionValue(t *testing.T) {

	portfolio := NewPortfolio(1, initialCapital)
	portfolio.PositionValue = -1

	if err := ValidatePortfolio(portfolio); err != ErrInvalidPositionValue {
		t.Fatalf("expected %v, got %v", ErrInvalidPositionValue, err)
	}
}

func TestPortfolioReset(t *testing.T) {

	portfolio := NewPortfolio(1, initialCapital)

	portfolio.Cash = 50000
	portfolio.Equity = 120000
	portfolio.PositionValue = 70000
	portfolio.UnrealizedPnL = 10000
	portfolio.RealizedPnL = 5000

	portfolio.Reset()

	if portfolio.Cash != portfolio.InitCapital {
		t.Fatalf(
			"expected cash %.2f, got %.2f",
			portfolio.InitCapital,
			portfolio.Cash,
		)
	}

	if portfolio.Equity != portfolio.InitCapital {
		t.Fatalf(
			"expected equity %.2f, got %.2f",
			portfolio.InitCapital,
			portfolio.Equity,
		)
	}

	if portfolio.PositionValue != 0 {
		t.Fatalf(
			"expected position value 0, got %.2f",
			portfolio.PositionValue,
		)
	}

	if portfolio.UnrealizedPnL != 0 {
		t.Fatalf(
			"expected unrealized pnl 0, got %.2f",
			portfolio.UnrealizedPnL,
		)
	}

	if portfolio.RealizedPnL != 0 {
		t.Fatalf(
			"expected realized pnl 0, got %.2f",
			portfolio.RealizedPnL,
		)
	}
}
