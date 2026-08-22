package order

import (
	"fmt"

	"event-driven-backtesting-engine/internal/domain"
	"event-driven-backtesting-engine/internal/events"
	"event-driven-backtesting-engine/internal/portfolio"
	"event-driven-backtesting-engine/internal/sizing"
)

type Manager struct {
    portfolio *portfolio.Engine
    queue     *events.EventQueue
    sizer     sizing.Sizer
}

func NewManager(
    portfolio *portfolio.Engine,
    queue *events.EventQueue,
    sizer sizing.Sizer,
) *Manager {
    return &Manager{
        portfolio: portfolio,
        queue:     queue,
        sizer:     sizer,
    }
}

// Consume allows Order Manager to be registered as an Event Consumer.
func (m *Manager) Consume(event events.Event) error {

	signal, ok := event.(events.SignalGeneratedEvent)
	if !ok {
		return fmt.Errorf("unsupported event %T", event)
	}

	p := m.portfolio.Portfolio()

	quantity := m.sizer.Size(
		p.Equity,
		signal.Price,
	)

	order := domain.NewOrder(
		signal.RunID,
		signal.Symbol,
		domain.OrderSide(signal.SignalType),
		quantity,
		signal.Price,
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
