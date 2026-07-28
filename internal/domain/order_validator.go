package domain

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidSymbol    = errors.New("symbol is required")
	ErrInvalidQuantity  = errors.New("quantity must be greater than zero")
	ErrInvalidSide      = errors.New("invalid order side")
	ErrInvalidOrderType = errors.New("invalid order type")
)

// ValidateOrder validates an order before it is submitted
// to the Order Manager / Matching Engine.
func ValidateOrder(order Order) error {
	if order.Symbol == "" {
		return ErrInvalidSymbol
	}

	if order.Quantity <= 0 {
		return ErrInvalidQuantity
	}

	switch order.Side {
	case BuyOrder, SellOrder:
		// valid
	default:
		return fmt.Errorf("%w: %q", ErrInvalidSide, order.Side)
	}

	switch order.Type {
	case MarketOrder:
		// valid
	default:
		return fmt.Errorf("%w: %q", ErrInvalidOrderType, order.Type)
	}

	return nil
}
