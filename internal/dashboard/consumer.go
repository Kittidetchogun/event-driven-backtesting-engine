package dashboard

import (
	"fmt"

	"event-driven-backtesting-engine/internal/events"
)

type Consumer struct {
	results []events.BacktestCompletedEvent
}

func NewConsumer() *Consumer {
	return &Consumer{
		results: make([]events.BacktestCompletedEvent, 0),
	}
}

// Consume receives BacktestCompletedEvent.
func (c *Consumer) Consume(event events.Event) error {
	completedEvent, ok := event.(events.BacktestCompletedEvent)
	if !ok {
		return fmt.Errorf(
			"unsupported event %T",
			event,
		)
	}

	c.results = append(
		c.results,
		completedEvent,
	)

	return nil
}

// Results returns all completed backtest results.
func (c *Consumer) Results() []events.BacktestCompletedEvent {
	return c.results
}
