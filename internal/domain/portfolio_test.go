package domain

import "testing"

func TestNewPortfolio(t *testing.T) {

	portfolio := NewPortfolio(1, 100000)

	if portfolio.RunID != 1 {
		t.Fatalf("expected RunID 1, got %d", portfolio.RunID)
	}

	if portfolio.InitCapital != 100000 {
		t.Fatalf("expected initial capital 100000, got %.2f", portfolio.InitCapital)
	}

	if portfolio.Cash != 100000 {
		t.Fatalf("expected cash 100000, got %.2f", portfolio.Cash)
	}

	if portfolio.Equity != 100000 {
		t.Fatalf("expected equity 100000, got %.2f", portfolio.Equity)
	}

	if portfolio.PositionValue != 0 {
		t.Fatalf("expected position value 0, got %.2f", portfolio.PositionValue)
	}

	if portfolio.UnrealizedPnL != 0 {
		t.Fatalf("expected unrealized pnl 0, got %.2f", portfolio.UnrealizedPnL)
	}

	if portfolio.RealizedPnL != 0 {
		t.Fatalf("expected realized pnl 0, got %.2f", portfolio.RealizedPnL)
	}
}

func TestValidatePortfolio_Valid(t *testing.T) {

	portfolio := NewPortfolio(1, 100000)

	if err := ValidatePortfolio(portfolio); err != nil {
		t.Fatalf("expected valid portfolio, got %v", err)
	}
}

func TestValidatePortfolio_InvalidInitialCapital(t *testing.T) {

	portfolio := NewPortfolio(1, -1)

	if err := ValidatePortfolio(portfolio); err != ErrInvalidInitialCapital {
		t.Fatalf("expected %v, got %v", ErrInvalidInitialCapital, err)
	}
}

func TestValidatePortfolio_InvalidCash(t *testing.T) {

	portfolio := NewPortfolio(1, 100000)
	portfolio.Cash = -1

	if err := ValidatePortfolio(portfolio); err != ErrInvalidCash {
		t.Fatalf("expected %v, got %v", ErrInvalidCash, err)
	}
}

func TestValidatePortfolio_InvalidEquity(t *testing.T) {

	portfolio := NewPortfolio(1, 100000)
	portfolio.Equity = -1

	if err := ValidatePortfolio(portfolio); err != ErrInvalidEquity {
		t.Fatalf("expected %v, got %v", ErrInvalidEquity, err)
	}
}

func TestValidatePortfolio_InvalidPositionValue(t *testing.T) {

	portfolio := NewPortfolio(1, 100000)
	portfolio.PositionValue = -1

	if err := ValidatePortfolio(portfolio); err != ErrInvalidPositionValue {
		t.Fatalf("expected %v, got %v", ErrInvalidPositionValue, err)
	}
}

func TestPortfolioReset(t *testing.T) {

	portfolio := NewPortfolio(1, 100000)

	portfolio.Cash = 50000
	portfolio.Equity = 120000
	portfolio.PositionValue = 70000
	portfolio.UnrealizedPnL = 10000
	portfolio.RealizedPnL = 5000

	portfolio.Reset()

	if portfolio.Cash != portfolio.InitCapital {
		t.Fatalf("expected cash %.2f, got %.2f",
			portfolio.InitCapital,
			portfolio.Cash,
		)
	}

	if portfolio.Equity != portfolio.InitCapital {
		t.Fatalf("expected equity %.2f, got %.2f",
			portfolio.InitCapital,
			portfolio.Equity,
		)
	}

	if portfolio.PositionValue != 0 {
		t.Fatalf("expected position value 0")
	}

	if portfolio.UnrealizedPnL != 0 {
		t.Fatalf("expected unrealized pnl 0")
	}

	if portfolio.RealizedPnL != 0 {
		t.Fatalf("expected realized pnl 0")
	}
}