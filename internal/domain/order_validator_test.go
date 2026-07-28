package domain

import (
	"errors"
	"testing"
	"time"
)

func validOrder() Order {
	return NewOrder(
		1,
		"BTCUSDT",
		BuyOrder,
		1,
		100000,
		time.Now(),
	)
}

func TestValidateOrder_Valid(t *testing.T) {
	order := validOrder()

	if err := ValidateOrder(order); err != nil {
		t.Fatalf("expected valid order, got %v", err)
	}
}

func TestValidateOrder_InvalidSymbol(t *testing.T) {
	order := validOrder()
	order.Symbol = ""

	err := ValidateOrder(order)

	if !errors.Is(err, ErrInvalidSymbol) {
		t.Fatalf("expected ErrInvalidSymbol, got %v", err)
	}
}

func TestValidateOrder_InvalidQuantity(t *testing.T) {
	order := validOrder()
	order.Quantity = 0

	err := ValidateOrder(order)

	if !errors.Is(err, ErrInvalidQuantity) {
		t.Fatalf("expected ErrInvalidQuantity, got %v", err)
	}
}

func TestValidateOrder_InvalidSide(t *testing.T) {
	order := validOrder()
	order.Side = OrderSide("INVALID")

	err := ValidateOrder(order)

	if !errors.Is(err, ErrInvalidSide) {
		t.Fatalf("expected ErrInvalidSide, got %v", err)
	}
}

func TestValidateOrder_InvalidOrderType(t *testing.T) {
	order := validOrder()
	order.Type = OrderType("LIMIT")

	err := ValidateOrder(order)

	if !errors.Is(err, ErrInvalidOrderType) {
		t.Fatalf("expected ErrInvalidOrderType, got %v", err)
	}
}
