package domain

import "time"

type OrderID int64

type OrderSide string

const (
	BuyOrder  OrderSide = "BUY"
	SellOrder OrderSide = "SELL"
)

type OrderType string

const (
	MarketOrder OrderType = "MARKET"
)

type OrderStatus string

const (
	PendingOrder   OrderStatus = "PENDING"
	FilledOrder    OrderStatus = "FILLED"
	CancelledOrder OrderStatus = "CANCELLED"
)

type Order struct {
	ID OrderID
	RunID int
	Symbol string
	Side OrderSide
	Type OrderType
	Quantity float64
	Price float64
	Status OrderStatus
	CreatedAt time.Time
	FilledAt *time.Time
}

// NewOrder creates a new pending market order.
func NewOrder(
	runID int,
	symbol string,
	side OrderSide,
	quantity float64,
	price float64,
	createdAt time.Time,
) Order {
	return Order{
		RunID:     runID,
		Symbol:    symbol,
		Side:      side,
		Type:      MarketOrder,
		Quantity:  quantity,
		Price:     price,
		Status:    PendingOrder,
		CreatedAt: createdAt,
	}
}
