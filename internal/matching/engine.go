package matching

import (
	"fmt"

	"event-driven-backtesting-engine/internal/domain"
	"event-driven-backtesting-engine/internal/events"
)

type Engine struct {
    queue *events.EventQueue

    trades []domain.Trade
}

func NewEngine(queue *events.EventQueue) *Engine {
    return &Engine{
        queue:  queue,
        trades: make([]domain.Trade, 0),
    }
}

func (e *Engine) Trades() []domain.Trade {
    return e.trades
}

// Consume receives OrderCreatedEvent from Event Queue.
func (e *Engine) Consume(event events.Event) error {

	// 1. Event ต้องเป็น OrderCreatedEvent
	orderEvent, ok := event.(events.OrderCreatedEvent)
	if !ok {
		return fmt.Errorf("unsupported event %T", event)
	}

	order := orderEvent.Order

	// 2. Validate Order
	if err := domain.ValidateOrder(order); err != nil {
		return err
	}

	// 3. Execute (Prototype: Market Order Fill ทันที)

	// 4. Fill Order
	if err := order.Fill(order.CreatedAt); err != nil {
		return err
	}

	// 5. Create Trade
	executedPrice := executionPrice(order)

	transactionCost := CalculateFee(
		executedPrice,
		order.Quantity,
		DefaultCommissionRate,
	)

	trade := domain.NewTrade(
		1,    // TODO: Generate Trade ID
		order.RunID,
		int(order.ID),
		order.Symbol,
		order.Side,
		order.Quantity,
		executedPrice,
		transactionCost,
		order.CreatedAt,    // Prototype: Execute immediately at CreatedAt
	)

	e.trades = append(e.trades, trade)

	// 6. Push TradeExecutedEvent
	tradeEvent := events.NewTradeExecutedEvent(trade)

	e.queue.Push(tradeEvent)

	return nil
}
