package domain

import "errors"

var (
	ErrInvalidPositionSymbol       = errors.New("symbol cannot be empty")
	ErrInvalidPositionQuantity     = errors.New("quantity must be greater than zero")
	ErrInvalidAveragePrice         = errors.New("average price must be greater than zero")
	ErrInvalidCurrentPrice         = errors.New("current price cannot be negative")
	ErrInvalidCurrentValue         = errors.New("current value cannot be negative")
)

// ValidatePosition validates a position before it is used.
func ValidatePosition(position Position) error {

	if position.Symbol == "" {
		return ErrInvalidPositionSymbol
	}

	if position.Quantity <= 0 {
		return ErrInvalidPositionQuantity
	}

	if position.AveragePrice <= 0 {
		return ErrInvalidAveragePrice
	}

	if position.CurrentPrice < 0 {
		return ErrInvalidCurrentPrice
	}

	if position.CurrentValue < 0 {
		return ErrInvalidCurrentValue
	}

	return nil
}
