package portfolio

import (
	"errors"
	"fmt"
	"time"

	"event-driven-backtesting-engine/internal/domain"
	"event-driven-backtesting-engine/internal/events"
)

var (
	ErrInsufficientCash     = errors.New("insufficient cash")
	ErrNoPosition           = errors.New("no position")
	ErrInsufficientPosition = errors.New("insufficient position")
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
	tradeEvent, ok := event.(events.TradeExecutedEvent)
	if !ok {
		return fmt.Errorf("unsupported event %T", event)
	}

	trade := tradeEvent.Trade

	// Capture the existing position before updating it.
	// This is required to calculate realized PnL using
	// the position's average entry price.
	position, hasPosition := e.positions[trade.Symbol]

	// 1. Update realized PnL before position may be deleted.
	if hasPosition {
		ApplyRealizedPnL(
			&e.portfolio,
			position,
			trade,
		)
	}

	// 2. Update Position
	UpdatePosition(
		e.positions,
		e.portfolio.RunID,
		trade,
	)

	// 3. Update Cash
	ApplyCashUpdate(
		&e.portfolio,
		trade,
	)

	// 4. Update Position Value
	UpdatePositionValue(
		&e.portfolio,
		e.positions,
	)

	// 5. Update portfolio-level Unrealized PnL
	UpdatePortfolioUnrealizedPnL(
		&e.portfolio,
		e.positions,
	)

	// 6. Record portfolio state
	e.recordSnapshot(trade.ExecutedTime)

	return nil
}

// UpdateMarketPrices updates the current market price of positions
// and records the latest portfolio state.
func (e *Engine) UpdateMarketPrices(candle domain.Candle) {
	position, ok := e.positions[candle.Symbol]
	if !ok {
		return
	}

	UpdatePrice(&position, candle.Close)
	UpdateMarketValue(&position)
	UpdateUnrealizedPnL(&position)

	e.positions[candle.Symbol] = position

	// Recalculate portfolio valuation.
	UpdatePositionValue(
		&e.portfolio,
		e.positions,
	)

	// Recalculate portfolio-level unrealized PnL.
	UpdatePortfolioUnrealizedPnL(
		&e.portfolio,
		e.positions,
	)

	// Record portfolio state.
	e.recordSnapshot(candle.Timestamp)
}

// recordSnapshot is the single place responsible for creating
// PortfolioSnapshot and publishing PortfolioUpdatedEvent.
func (e *Engine) recordSnapshot(timestamp time.Time) {
	UpdateEquity(&e.portfolio)
	e.portfolio.UpdateTimestamp(timestamp)

	snapshot := NewPortfolioSnapshot(e.portfolio)
	replaced := false
	for index := range e.snapshots {
		if e.snapshots[index].Time.Equal(snapshot.Time) {
			e.snapshots[index] = snapshot
			replaced = true
			break
		}
	}

	if !replaced {
		e.snapshots = append(e.snapshots, snapshot)
	}

	portfolioEvent := events.NewPortfolioUpdatedEvent(
		e.portfolio,
	)

	e.queue.Push(portfolioEvent)
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

	if order.Price <= 0 {
		return nil
	}

	requiredCash := order.Quantity * order.Price

	if e.portfolio.Cash < requiredCash {
		return fmt.Errorf(
			"%w: required=%.2f available=%.2f",
			ErrInsufficientCash,
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
			"%w: symbol=%s",
			ErrNoPosition,
			order.Symbol,
		)
	}

	if position.Quantity < order.Quantity {
		return fmt.Errorf(
			"%w: symbol=%s required=%.8f available=%.8f",
			ErrInsufficientPosition,
			order.Symbol,
			order.Quantity,
			position.Quantity,
		)
	}

	return nil
}
