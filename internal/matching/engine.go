package matching

import (
	"fmt"

	"event-driven-backtesting-engine/internal/domain"
	"event-driven-backtesting-engine/internal/events"
)

type Engine struct {
	queue *events.EventQueue

	trades []domain.Trade

	nextTradeID int

	pendingOrders []domain.Order
}

func NewEngine(queue *events.EventQueue) *Engine {
	return &Engine{
		queue:         queue,
		trades:        make([]domain.Trade, 0),
		nextTradeID:   1,
		pendingOrders: make([]domain.Order, 0),
	}
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
		return err
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

	remaining := make([]domain.Order, 0)

	for _, order := range e.pendingOrders {

		// Safety check: an order created on the same candle
		// must not be executed on that candle.
		if !order.CreatedAt.Before(candle.Timestamp) {
			remaining = append(remaining, order)
			continue
		}

		// Fill at the current candle timestamp.
		if err := order.Fill(candle.Timestamp); err != nil {
			return err
		}

		// Execution price = NEXT candle OPEN.
		executedPrice := candle.Open

		transactionCost := CalculateFee(
			executedPrice,
			order.Quantity,
			DefaultCommissionRate,
		)

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
