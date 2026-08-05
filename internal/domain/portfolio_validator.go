package domain

import "errors"

var (
	ErrInvalidInitialCapital = errors.New("initial capital must be greater than or equal to zero")
	ErrInvalidCash           = errors.New("cash cannot be negative")
	ErrInvalidEquity         = errors.New("equity cannot be negative")
	ErrInvalidPositionValue  = errors.New("position value cannot be negative")
)

// ValidatePortfolio validates a portfolio before it is used.
func ValidatePortfolio(p Portfolio) error {

	if p.InitCapital < 0 {
		return ErrInvalidInitialCapital
	}

	if p.Cash < 0 {
		return ErrInvalidCash
	}

	if p.Equity < 0 {
		return ErrInvalidEquity
	}

	if p.PositionValue < 0 {
		return ErrInvalidPositionValue
	}

	return nil
}
