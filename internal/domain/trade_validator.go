package domain

import (
	"errors"
)

var (
	ErrInvalidTradeSymbol          = errors.New("trade symbol cannot be empty")
	ErrInvalidTradeQuantity        = errors.New("trade quantity must be greater than zero")
	ErrInvalidExecutedPrice        = errors.New("executed price must be greater than zero")
	ErrInvalidTradeSide            = errors.New("invalid trade side")
	ErrInvalidTradeExecutionTime   = errors.New("executed time cannot be zero")
)

func ValidateTrade(trade Trade) error {

	if trade.Symbol == "" {
		return ErrInvalidTradeSymbol
	}

	if trade.Quantity <= 0 {
		return ErrInvalidTradeQuantity
	}

	if trade.ExecutedPrice <= 0 {
		return ErrInvalidExecutedPrice
	}

	switch trade.Side {
	case BuyOrder, SellOrder:
		// valid
	default:
		return ErrInvalidTradeSide
	}

	if trade.ExecutedTime.IsZero() {
		return ErrInvalidTradeExecutionTime
	}

	return nil
}
