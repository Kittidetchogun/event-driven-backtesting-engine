package portfolio

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"event-driven-backtesting-engine/internal/domain"
	"event-driven-backtesting-engine/internal/events"
)

var (
	ErrInsufficientCash      = errors.New("insufficient cash")
	ErrNoPosition            = errors.New("no position")
	ErrInsufficientPosition  = errors.New("insufficient position")
	ErrReservationExists     = errors.New("cash reservation already exists")
	ErrReservationNotFound   = errors.New("cash reservation not found")
	ErrInvalidReservation    = errors.New("invalid cash reservation")
	ErrSettlementMismatch    = errors.New("settlement does not match cash reservation")
	ErrReservedCashInvariant = errors.New("reserved cash invariant violated")
	ErrInvalidMarketPrice    = errors.New("invalid market price")
)

const reservationTolerance = 1e-12

type cashReservation struct {
	amount float64
}

type Engine struct {
	mu           sync.Mutex
	queue        *events.EventQueue
	portfolio    domain.Portfolio
	positions    map[string]domain.Position
	snapshots    []PortfolioSnapshot
	reservedCash float64
	reservations map[domain.OrderID]cashReservation
}

func NewEngine(
	queue *events.EventQueue,
	portfolio domain.Portfolio,
) *Engine {
	return &Engine{
		queue:        queue,
		portfolio:    portfolio,
		positions:    make(map[string]domain.Position),
		snapshots:    make([]PortfolioSnapshot, 0),
		reservations: make(map[domain.OrderID]cashReservation),
	}
}

// Consume receives TradeExecutedEvent from Event Queue.
func (e *Engine) Consume(event events.Event) error {
	tradeEvent, ok := event.(events.TradeExecutedEvent)
	if !ok {
		return fmt.Errorf("unsupported event %T", event)
	}

	trade := tradeEvent.Trade

	e.mu.Lock()
	defer e.mu.Unlock()

	reservation, err := e.validateReservationForTrade(trade)
	if err != nil {
		return err
	}
	nextReservedCash, err := e.nextReservedCash(reservation.amount)
	if err != nil {
		return err
	}

	// Capture the existing position before updating it.
	// This is required to calculate realized PnL using
	// the position's average entry price.
	position, hasPosition := e.positions[trade.Symbol]
	if trade.Side == domain.SellOrder {
		if !hasPosition {
			return fmt.Errorf("%w: symbol=%s", ErrNoPosition, trade.Symbol)
		}
		if position.Quantity < trade.Quantity {
			return fmt.Errorf(
				"%w: symbol=%s required=%.8f available=%.8f",
				ErrInsufficientPosition,
				trade.Symbol,
				trade.Quantity,
				position.Quantity,
			)
		}
	}

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
	applyCashUpdate(
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
	delete(e.reservations, domain.OrderID(trade.OrderID))
	e.reservedCash = nextReservedCash

	return nil
}

// Reserve atomically reduces available cash for a pending execution.
func (e *Engine) Reserve(orderID domain.OrderID, amount float64) error {
	if orderID <= 0 || amount < 0 || math.IsNaN(amount) || math.IsInf(amount, 0) {
		return ErrInvalidReservation
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.reservations[orderID]; exists {
		return ErrReservationExists
	}
	if e.portfolio.Cash-e.reservedCash < amount {
		return fmt.Errorf("%w: required=%.2f available=%.2f", ErrInsufficientCash, amount, e.portfolio.Cash-e.reservedCash)
	}

	e.reservations[orderID] = cashReservation{amount: amount}
	e.reservedCash += amount
	return nil
}

// Release removes an unused reservation exactly once.
func (e *Engine) Release(orderID domain.OrderID) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	reservation, ok := e.reservations[orderID]
	if !ok {
		return ErrReservationNotFound
	}
	nextReservedCash, err := e.nextReservedCash(reservation.amount)
	if err != nil {
		return err
	}
	delete(e.reservations, orderID)
	e.reservedCash = nextReservedCash
	return nil
}

func (e *Engine) validateReservationForTrade(trade domain.Trade) (cashReservation, error) {
	orderID := domain.OrderID(trade.OrderID)
	reservation, ok := e.reservations[orderID]
	if !ok {
		return cashReservation{}, fmt.Errorf("%w: order_id=%d", ErrReservationNotFound, trade.OrderID)
	}
	if orderID <= 0 || trade.Symbol == "" || trade.Quantity <= 0 || math.IsNaN(trade.Quantity) || math.IsInf(trade.Quantity, 0) ||
		trade.ExecutedPrice <= 0 || math.IsNaN(trade.ExecutedPrice) || math.IsInf(trade.ExecutedPrice, 0) ||
		trade.TransactionCost < 0 || math.IsNaN(trade.TransactionCost) || math.IsInf(trade.TransactionCost, 0) {
		return cashReservation{}, fmt.Errorf("%w: invalid trade data for order_id=%d", ErrSettlementMismatch, trade.OrderID)
	}
	expected := 0.0
	if trade.Side == domain.BuyOrder {
		expected = trade.ExecutedPrice*trade.Quantity + trade.TransactionCost
	} else if trade.Side != domain.SellOrder {
		return cashReservation{}, fmt.Errorf("%w: invalid trade side for order_id=%d", ErrSettlementMismatch, trade.OrderID)
	}
	if math.Abs(reservation.amount-expected) > reservationTolerance {
		return cashReservation{}, fmt.Errorf("%w: order_id=%d reserved=%.12f settled=%.12f", ErrSettlementMismatch, trade.OrderID, reservation.amount, expected)
	}
	return reservation, nil
}

func (e *Engine) nextReservedCash(amount float64) (float64, error) {
	next := e.reservedCash - amount
	if next < -reservationTolerance {
		return 0, ErrReservedCashInvariant
	}
	if math.Abs(next) <= reservationTolerance {
		next = 0
	}
	return next, nil
}

func (e *Engine) ReservedCash() float64 {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.reservedCash
}

func (e *Engine) AvailableCash() float64 {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.portfolio.Cash - e.reservedCash
}

// UpdateMarketPrices updates the current market price of positions
// and records the latest portfolio state.
func (e *Engine) UpdateMarketPrices(candle domain.Candle) error {
	if candle.Close <= 0 || math.IsNaN(candle.Close) || math.IsInf(candle.Close, 0) {
		return fmt.Errorf("%w: close=%v", ErrInvalidMarketPrice, candle.Close)
	}

	position, ok := e.positions[candle.Symbol]
	if !ok {
		e.recordSnapshot(candle.Timestamp)
		return nil
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
	return nil
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

func (e *Engine) RecordSnapshot(timestamp time.Time) {
	e.recordSnapshot(timestamp)
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
