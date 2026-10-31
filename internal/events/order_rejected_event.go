package events

import (
	"time"

	"event-driven-backtesting-engine/internal/domain"
)

const OrderRejectedEventType = "OrderRejectedEvent"

type OrderRejectedEvent struct {
	BaseEvent
	Order  domain.Order
	Reason string
}

var _ Event = OrderRejectedEvent{}

func NewOrderRejectedEvent(
	order domain.Order,
	reason string,
	rejectedAt ...time.Time,
) OrderRejectedEvent {
	timestamp := order.CreatedAt
	if order.RejectedAt != nil {
		timestamp = *order.RejectedAt
	}
	if len(rejectedAt) > 0 {
		timestamp = rejectedAt[0]
	}

	return OrderRejectedEvent{
		BaseEvent: NewBaseEvent(OrderRejectedEventType, timestamp),
		Order:     order,
		Reason:    reason,
	}
}
