package domain

import (
	"testing"
	"time"
)

func TestDummyPortfolioChecker_CanBuy(t *testing.T) {
	checker := DummyPortfolioChecker{}

	order := NewOrder(
		1,
		"BTCUSDT",
		BuyOrder,
		1,
		100000,
		time.Now(),
	)

	if err := checker.CanBuy(order); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestDummyPortfolioChecker_CanSell(t *testing.T) {
	checker := DummyPortfolioChecker{}

	order := NewOrder(
		1,
		"BTCUSDT",
		SellOrder,
		1,
		100000,
		time.Now(),
	)

	if err := checker.CanSell(order); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}
