package events

import "event-driven-backtesting-engine/internal/domain"

const OrderCreatedEventType = "OrderCreatedEvent"

type OrderCreatedEvent struct {
	BaseEvent
	Order domain.Order
}

var _ Event = OrderCreatedEvent{}

// NewOrderCreatedEvent creates an event after the Order Manager
// successfully creates a new order.
func NewOrderCreatedEvent(order domain.Order) OrderCreatedEvent {
	return OrderCreatedEvent{
		BaseEvent: NewBaseEvent(
			OrderCreatedEventType,
			order.CreatedAt,
		),
		Order: order,
	}
}
