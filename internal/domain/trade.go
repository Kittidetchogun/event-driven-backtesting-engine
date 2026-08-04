package domain

import "time"

type Trade struct {
	TradeID int
	RunID   int
	OrderID int

	Symbol string
	Side   OrderSide

	Quantity        float64
	ExecutedPrice   float64
	TransactionCost float64

	ExecutedTime time.Time
}

func NewTrade(
	tradeID int,
	runID int,
	orderID int,
	symbol string,
	side OrderSide,
	quantity float64,
	executedPrice float64,
	transactionCost float64,
	executedTime time.Time,
) Trade {
	return Trade{
		TradeID:         tradeID,
		RunID:           runID,
		OrderID:         orderID,
		Symbol:          symbol,
		Side:            side,
		Quantity:        quantity,
		ExecutedPrice:   executedPrice,
		TransactionCost: transactionCost,
		ExecutedTime:    executedTime,
	}
}
