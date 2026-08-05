package domain

import "testing"

func TestNewPosition(t *testing.T) {

	position := NewPosition(
		1,
		1,
		"BTCUSDT",
		BuyOrder,
		2,
		50000,
		51000,
	)

	if position.PositionID != 1 {
		t.Fatalf("expected PositionID 1, got %d", position.PositionID)
	}

	if position.PortfolioID != 1 {
		t.Fatalf("expected PortfolioID 1, got %d", position.PortfolioID)
	}

	if position.Symbol != "BTCUSDT" {
		t.Fatalf("expected BTCUSDT, got %s", position.Symbol)
	}

	if position.Side != BuyOrder {
		t.Fatalf("expected BUY, got %s", position.Side)
	}

	if position.Quantity != 2 {
		t.Fatalf("expected quantity 2, got %.2f", position.Quantity)
	}

	if position.AveragePrice != 50000 {
		t.Fatalf("expected average price 50000, got %.2f", position.AveragePrice)
	}

	if position.CurrentPrice != 51000 {
		t.Fatalf("expected current price 51000, got %.2f", position.CurrentPrice)
	}

	if position.CurrentValue != 102000 {
		t.Fatalf("expected current value 102000, got %.2f", position.CurrentValue)
	}

	if position.UnrealizedPnL != 2000 {
		t.Fatalf("expected unrealized pnl 2000, got %.2f", position.UnrealizedPnL)
	}

	if position.RealizedPnL != 0 {
		t.Fatalf("expected realized pnl 0, got %.2f", position.RealizedPnL)
	}
}

func TestValidatePosition_Valid(t *testing.T) {

	position := NewPosition(
		1,
		1,
		"BTCUSDT",
		BuyOrder,
		1,
		50000,
		51000,
	)

	if err := ValidatePosition(position); err != nil {
		t.Fatalf("expected valid position, got %v", err)
	}
}

func TestValidatePosition_InvalidSymbol(t *testing.T) {

	position := NewPosition(
		1,
		1,
		"",
		BuyOrder,
		1,
		50000,
		51000,
	)

	if err := ValidatePosition(position); err != ErrInvalidPositionSymbol {
		t.Fatalf("expected %v, got %v", ErrInvalidPositionSymbol, err)
	}
}

func TestValidatePosition_InvalidQuantity(t *testing.T) {

	position := NewPosition(
		1,
		1,
		"BTCUSDT",
		BuyOrder,
		0,
		50000,
		51000,
	)

	if err := ValidatePosition(position); err != ErrInvalidPositionQuantity {
		t.Fatalf("expected %v, got %v", ErrInvalidPositionQuantity, err)
	}
}

func TestValidatePosition_InvalidAveragePrice(t *testing.T) {

	position := NewPosition(
		1,
		1,
		"BTCUSDT",
		BuyOrder,
		1,
		0,
		51000,
	)

	if err := ValidatePosition(position); err != ErrInvalidAveragePrice {
		t.Fatalf("expected %v, got %v", ErrInvalidAveragePrice, err)
	}
}

func TestValidatePosition_InvalidCurrentPrice(t *testing.T) {

	position := NewPosition(
		1,
		1,
		"BTCUSDT",
		BuyOrder,
		1,
		50000,
		-1,
	)

	if err := ValidatePosition(position); err != ErrInvalidCurrentPrice {
		t.Fatalf("expected %v, got %v", ErrInvalidCurrentPrice, err)
	}
}

func TestValidatePosition_InvalidCurrentValue(t *testing.T) {

	position := NewPosition(
		1,
		1,
		"BTCUSDT",
		BuyOrder,
		1,
		50000,
		51000,
	)

	position.CurrentValue = -1

	if err := ValidatePosition(position); err != ErrInvalidCurrentValue {
		t.Fatalf("expected %v, got %v", ErrInvalidCurrentValue, err)
	}
}

func TestPositionReset(t *testing.T) {

	position := NewPosition(
		1,
		1,
		"BTCUSDT",
		BuyOrder,
		2,
		50000,
		51000,
	)

	position.RealizedPnL = 1000

	position.Reset()

	if position.Quantity != 0 {
		t.Fatalf("expected quantity 0")
	}

	if position.AveragePrice != 0 {
		t.Fatalf("expected average price 0")
	}

	if position.CurrentPrice != 0 {
		t.Fatalf("expected current price 0")
	}

	if position.CurrentValue != 0 {
		t.Fatalf("expected current value 0")
	}

	if position.UnrealizedPnL != 0 {
		t.Fatalf("expected unrealized pnl 0")
	}

	if position.RealizedPnL != 0 {
		t.Fatalf("expected realized pnl 0")
	}
}

func TestPositionUpdatePrice(t *testing.T) {

	position := NewPosition(
		1,
		1,
		"BTCUSDT",
		BuyOrder,
		2,
		50000,
		50000,
	)

	position.UpdatePrice(52000)

	if position.CurrentPrice != 52000 {
		t.Fatalf("expected current price 52000")
	}

	if position.CurrentValue != 104000 {
		t.Fatalf("expected current value 104000, got %.2f", position.CurrentValue)
	}

	if position.UnrealizedPnL != 4000 {
		t.Fatalf("expected unrealized pnl 4000, got %.2f", position.UnrealizedPnL)
	}
}
