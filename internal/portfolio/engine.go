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

func (e *Engine) CanBuy(order domain.Order) error {
	if order.Side != domain.BuyOrder {
		return fmt.Errorf("order is not a buy order")
	}

	if order.Quantity <= 0 {
		return fmt.Errorf("order quantity must be greater than zero")
	}

	// Market order ยังไม่มีราคาตอน Order Manager ทำงาน
	// จึงยังไม่สามารถตรวจ cash requirement ได้
	if order.Price <= 0 {
		return nil
	}

	requiredCash := order.Quantity * order.Price

	if e.portfolio.Cash < requiredCash {
		return fmt.Errorf(
			"insufficient cash: required=%.2f available=%.2f",
			requiredCash,
			e.portfolio.Cash,
		)
	}

	return nil
}

func (e *Engine) CanSell(order domain.Order) error {
	if order.Side != domain.SellOrder {
		return fmt.Errorf("order is not a sell order")
	}

	if order.Quantity <= 0 {
		return fmt.Errorf("order quantity must be greater than zero")
	}

	position, ok := e.positions[order.Symbol]
	if !ok {
		return fmt.Errorf(
			"no position for symbol %s",
			order.Symbol,
		)
	}

	if position.Quantity < order.Quantity {
		return fmt.Errorf(
			"insufficient position: symbol=%s required=%.8f available=%.8f",
			order.Symbol,
			order.Quantity,
			position.Quantity,
		)
	}

	return nil
}
