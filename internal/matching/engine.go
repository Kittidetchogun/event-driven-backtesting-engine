package matching

import (
	"errors"
	"fmt"
	"math"
	"time"

	"event-driven-backtesting-engine/internal/domain"
	"event-driven-backtesting-engine/internal/events"
)

var (
	ErrInvalidExecutionPrice    = errors.New("invalid execution price")
	ErrInvalidExecutionQuantity = errors.New("invalid execution quantity")
	ErrCashReserverRequired     = errors.New("cash reservation provider is required")
)

// CashReserver is the minimal portfolio capability required by matching.
// The portfolio remains the owner of cash and reservation state.
type CashReserver interface {
	Reserve(orderID domain.OrderID, amount float64) error
	Release(orderID domain.OrderID) error
}

type Engine struct {
	queue    *events.EventQueue
	reserver CashReserver

	trades []domain.Trade

	nextTradeID int

	pendingOrders []domain.Order
}

func NewEngine(queue *events.EventQueue, reserver CashReserver) (*Engine, error) {
	if queue == nil {
		return nil, errors.New("event queue is required")
	}
	if reserver == nil {
		return nil, ErrCashReserverRequired
	}

	return &Engine{
		queue:         queue,
		reserver:      reserver,
		trades:        make([]domain.Trade, 0),
		nextTradeID:   1,
		pendingOrders: make([]domain.Order, 0),
	}, nil
}

func (e *Engine) Trades() []domain.Trade {
	return e.trades
}

// Consume receives OrderCreatedEvent and stores the order
// for execution on the next candle.
func (e *Engine) Consume(event events.Event) error {

	orderEvent, ok := event.(events.OrderCreatedEvent)
	if !ok {
		return fmt.Errorf("unsupported event %T", event)
	}

	order := orderEvent.Order

	// Validate order before accepting it.
	if err := domain.ValidateOrder(order); err != nil {
		if rejectErr := order.Reject(err.Error(), order.CreatedAt); rejectErr != nil {
			return rejectErr
		}
		e.queue.Push(events.NewOrderRejectedEvent(order, err.Error()))
		return nil
	}

	// Do NOT fill here.
	// The order must wait until the next candle.
	e.pendingOrders = append(e.pendingOrders, order)

	return nil
}

// ExecutePending executes all orders that were created
// before the current candle.
//
// Orders are executed at the current candle's Open price.
func (e *Engine) ExecutePending(candle domain.Candle) error {

	if len(e.pendingOrders) == 0 {
		return nil
	}

	remaining := make([]domain.Order, 0, len(e.pendingOrders))

	for index, order := range e.pendingOrders {

		// Safety check: an order created on the same candle
		// must not be executed on that candle.
		if !order.CreatedAt.Before(candle.Timestamp) {
			remaining = append(remaining, order)
			continue
		}

		// Execution price = NEXT candle OPEN.
		executedPrice := candle.Open
		if executedPrice <= 0 || math.IsNaN(executedPrice) || math.IsInf(executedPrice, 0) {
			if err := e.rejectPendingOrder(&order, candle.Timestamp, ErrInvalidExecutionPrice.Error()); err != nil {
				return err
			}
			continue
		}
		if order.Quantity <= 0 || math.IsNaN(order.Quantity) || math.IsInf(order.Quantity, 0) {
			if err := e.rejectPendingOrder(&order, candle.Timestamp, ErrInvalidExecutionQuantity.Error()); err != nil {
				return err
			}
			continue
		}

		transactionCost := CalculateFee(
			executedPrice,
			order.Quantity,
			DefaultCommissionRate,
		)
		totalCost := executedPrice*order.Quantity + transactionCost
		if err := e.reserver.Reserve(order.ID, totalCostForOrder(order, totalCost)); err != nil {
			if rejectErr := e.rejectPendingOrder(&order, candle.Timestamp, err.Error()); rejectErr != nil {
				return rejectErr
			}
			continue
		}

		// Fill only after execution price, quantity and cash have passed validation.
		if err := order.Fill(candle.Timestamp); err != nil {
			releaseErr := e.reserver.Release(order.ID)
			remaining = append(remaining, e.pendingOrders[index+1:]...)
			e.pendingOrders = remaining
			if releaseErr != nil {
				return fmt.Errorf("fill failed: %w; release reservation failed: %v", err, releaseErr)
			}
			return err
		}

		tradeID := e.nextTradeID
		e.nextTradeID++

		trade := domain.NewTrade(
			tradeID,
			order.RunID,
			int(order.ID),
			order.Symbol,
			order.Side,
			order.Quantity,
			executedPrice,
			transactionCost,
			candle.Timestamp,
		)

		e.trades = append(e.trades, trade)

		tradeEvent := events.NewTradeExecutedEvent(trade)

		e.queue.Push(tradeEvent)
	}

	e.pendingOrders = remaining

	return nil
}

func totalCostForOrder(order domain.Order, buyCost float64) float64 {
	if order.Side == domain.BuyOrder {
		return buyCost
	}
	return 0
}

func (e *Engine) rejectPendingOrder(order *domain.Order, rejectedAt time.Time, reason string) error {
	if err := order.Reject(reason, rejectedAt); err != nil {
		return err
	}
	e.queue.Push(events.NewOrderRejectedEvent(*order, reason, rejectedAt))
	return nil
}
