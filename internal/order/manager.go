package order

import (
	"fmt"

	"event-driven-backtesting-engine/internal/domain"
	"event-driven-backtesting-engine/internal/events"
)

type Manager struct {
	portfolio domain.PortfolioChecker
	queue      *events.EventQueue
}

func NewManager(
	portfolio domain.PortfolioChecker,
	queue *events.EventQueue,
) *Manager {
	return &Manager{
		portfolio: portfolio,
		queue:      queue,
	}
}

// Consume allows Order Manager to be registered as an Event Consumer.
func (m *Manager) Consume(event events.Event) error {

	signal, ok := event.(events.SignalGeneratedEvent)
	if !ok {
		return fmt.Errorf("unsupported event %T", event)
	}

	order := domain.NewOrder(
		signal.RunID,
		signal.Symbol,
		domain.OrderSide(signal.SignalType),
		signal.Quantity,
		0, // Market order, Matching Engine will determine execution price.
		signal.SignalTime,
	)

	if err := domain.ValidateOrder(order); err != nil {
		return err
	}

	switch order.Side {

	case domain.BuyOrder:
		if err := m.portfolio.CanBuy(order); err != nil {
			return err
		}

	case domain.SellOrder:
		if err := m.portfolio.CanSell(order); err != nil {
			return err
		}
	}

	orderEvent := events.NewOrderCreatedEvent(order)

	m.queue.Push(orderEvent)

	return nil
}
