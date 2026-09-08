package events

import (
	"event-driven-backtesting-engine/internal/domain"
)

const OrderRejectedEventType = "OrderRejectedEvent"

type OrderRejectedEvent struct {
	BaseEvent
	Order domain.Order
	Reason string
}

var _ Event = OrderRejectedEvent{}

func NewOrderRejectedEvent(
	order domain.Order,
	reason string,
) OrderRejectedEvent {
	return OrderRejectedEvent{
		BaseEvent: NewBaseEvent(
			OrderRejectedEventType,
			order.CreatedAt,
		),
		Order:  order,
		Reason: reason,
	}
}
