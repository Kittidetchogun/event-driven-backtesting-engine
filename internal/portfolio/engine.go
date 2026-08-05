package portfolio

import (
	"fmt"

	"event-driven-backtesting-engine/internal/domain"
	"event-driven-backtesting-engine/internal/events"
)

type Engine struct {
	queue     *events.EventQueue
	portfolio domain.Portfolio
	positions map[string]domain.Position
}

func NewEngine(
	queue *events.EventQueue,
	portfolio domain.Portfolio,
) *Engine {

	return &Engine{
		queue:     queue,
		portfolio: portfolio,
		positions: make(map[string]domain.Position),
	}
}

// Consume receives TradeExecutedEvent from Event Queue.
func (e *Engine) Consume(event events.Event) error {

	// 1. Event ต้องเป็น TradeExecutedEvent
	tradeEvent, ok := event.(events.TradeExecutedEvent)
	if !ok {
		return fmt.Errorf("unsupported event %T", event)
	}

	trade := tradeEvent.Trade

	// 2. Update Cash
	switch trade.Side {

	case domain.BuyOrder:
		e.portfolio.Cash -=
			trade.ExecutedPrice*trade.Quantity +
				trade.TransactionCost

	case domain.SellOrder:
		e.portfolio.Cash +=
			trade.ExecutedPrice*trade.Quantity -
				trade.TransactionCost
	}

	// 3. Update Position
	position, exists := e.positions[trade.Symbol]

	if !exists {
		position = domain.NewPosition(
			1, // TODO: Generate PositionID
			e.portfolio.RunID,
			trade.Symbol,
			trade.Side,
			0,
			0,
			trade.ExecutedPrice,
		)
	}

	switch trade.Side {

	case domain.BuyOrder:

		totalCost :=
			position.AveragePrice*position.Quantity +
				trade.ExecutedPrice*trade.Quantity

		position.Quantity += trade.Quantity

		if position.Quantity > 0 {
			position.AveragePrice =
				totalCost / position.Quantity
		}

	case domain.SellOrder:

		position.Quantity -= trade.Quantity

		if position.Quantity < 0 {
			position.Quantity = 0
		}
	}

	position.UpdatePrice(trade.ExecutedPrice)

	e.positions[trade.Symbol] = position

	// 4. Update Equity
	e.portfolio.PositionValue = position.CurrentValue
	e.portfolio.UpdateEquity()
	e.portfolio.UpdateTimestamp(trade.ExecutedTime)

	// 5. Push PortfolioUpdatedEvent
	portfolioEvent :=
		events.NewPortfolioUpdatedEvent(e.portfolio)

	e.queue.Push(portfolioEvent)

	return nil
}

func (e *Engine) Portfolio() domain.Portfolio {
	return e.portfolio
}

func (e *Engine) Positions() map[string]domain.Position {
	return e.positions
}
