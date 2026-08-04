package domain

import (
	"testing"
	"time"
)

func TestNewTrade(t *testing.T) {

	now := time.Now()

	trade := NewTrade(
		1,
		1,
		10,
		"BTCUSDT",
		BuyOrder,
		1,
		50000,
		10,
		now,
	)

	if trade.TradeID != 1 {
		t.Fatalf("expected TradeID 1 got %d", trade.TradeID)
	}

	if trade.RunID != 1 {
		t.Fatalf("expected RunID 1 got %d", trade.RunID)
	}

	if trade.OrderID != 10 {
		t.Fatalf("expected OrderID 10 got %d", trade.OrderID)
	}

	if trade.Symbol != "BTCUSDT" {
		t.Fatalf("expected symbol BTCUSDT got %s", trade.Symbol)
	}

	if trade.Side != BuyOrder {
		t.Fatalf("expected side BUY got %s", trade.Side)
	}

	if trade.Quantity != 1 {
		t.Fatalf("expected quantity 1 got %f", trade.Quantity)
	}

	if trade.ExecutedPrice != 50000 {
		t.Fatalf("expected executed price 50000 got %f", trade.ExecutedPrice)
	}

	if trade.TransactionCost != 10 {
		t.Fatalf("expected transaction cost 10 got %f", trade.TransactionCost)
	}

	if !trade.ExecutedTime.Equal(now) {
		t.Fatal("expected executed time to match")
	}
}

func TestValidateTrade_Valid(t *testing.T) {

	trade := NewTrade(
		1,
		1,
		1,
		"BTCUSDT",
		BuyOrder,
		1,
		50000,
		10,
		time.Now(),
	)

	if err := ValidateTrade(trade); err != nil {
		t.Fatal(err)
	}
}

func TestValidateTrade_InvalidSymbol(t *testing.T) {

	trade := NewTrade(
		1,
		1,
		1,
		"",
		BuyOrder,
		1,
		50000,
		10,
		time.Now(),
	)

	if err := ValidateTrade(trade); err != ErrInvalidTradeSymbol {
		t.Fatalf("expected ErrInvalidTradeSymbol got %v", err)
	}
}

func TestValidateTrade_InvalidQuantity(t *testing.T) {

	trade := NewTrade(
		1,
		1,
		1,
		"BTCUSDT",
		BuyOrder,
		0,
		50000,
		10,
		time.Now(),
	)

	if err := ValidateTrade(trade); err != ErrInvalidTradeQuantity {
		t.Fatalf("expected ErrInvalidTradeQuantity got %v", err)
	}
}

func TestValidateTrade_InvalidExecutedPrice(t *testing.T) {

	trade := NewTrade(
		1,
		1,
		1,
		"BTCUSDT",
		BuyOrder,
		1,
		0,
		10,
		time.Now(),
	)

	if err := ValidateTrade(trade); err != ErrInvalidExecutedPrice {
		t.Fatalf("expected ErrInvalidExecutedPrice got %v", err)
	}
}

func TestValidateTrade_InvalidSide(t *testing.T) {

	trade := NewTrade(
		1,
		1,
		1,
		"BTCUSDT",
		OrderSide("INVALID"),
		1,
		50000,
		10,
		time.Now(),
	)

	if err := ValidateTrade(trade); err != ErrInvalidTradeSide {
		t.Fatalf("expected ErrInvalidTradeSide got %v", err)
	}
}

func TestValidateTrade_InvalidExecutedTime(t *testing.T) {

	trade := NewTrade(
		1,
		1,
		1,
		"BTCUSDT",
		BuyOrder,
		1,
		50000,
		10,
		time.Time{},
	)

	if err := ValidateTrade(trade); err != ErrInvalidTradeExecutionTime {
		t.Fatalf("expected ErrInvalidTradeExecutionTime got %v", err)
	}
}
