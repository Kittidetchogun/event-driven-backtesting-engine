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
	snapshots []PortfolioSnapshot
}

func NewEngine(
	queue *events.EventQueue,
	portfolio domain.Portfolio,
) *Engine {

	return &Engine{
		queue:     queue,
		portfolio: portfolio,
		positions: make(map[string]domain.Position),
		snapshots: make([]PortfolioSnapshot, 0),
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

	// 2. Update Position
	UpdatePosition(
		e.positions,
		e.portfolio.RunID,
		trade,
	)

	ApplyCashUpdate(
		&e.portfolio,
		trade,
	)

	// 3. Update Position Value
	UpdatePositionValue(
		&e.portfolio,
		e.positions,
	)

	// 4. Update Equity
	UpdateEquity(&e.portfolio)
	e.portfolio.UpdateTimestamp(trade.ExecutedTime)

	snapshot := NewPortfolioSnapshot(e.portfolio)

	e.snapshots = append(
		e.snapshots,
		snapshot,
	)

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

func (e *Engine) Snapshots() []PortfolioSnapshot {
	return e.snapshots
}
