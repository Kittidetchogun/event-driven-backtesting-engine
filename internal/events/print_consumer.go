package events

import (
	"fmt"
)

// Compile-time interface check
var _ Consumer = (*PrintConsumer)(nil)

// PrintConsumer is a simple consumer used to verify
// the event flow during development.
type PrintConsumer struct{}

// NewPrintConsumer creates a new PrintConsumer.
func NewPrintConsumer() *PrintConsumer {
	return &PrintConsumer{}
}

// Consume receives an event and prints the candle timestamp.
func (c *PrintConsumer) Consume(event Event) error {
	if event == nil {
		return fmt.Errorf("event is nil")
	}

	candleEvent, ok := event.(CandleReceivedEvent)
	if !ok {
		return fmt.Errorf("unsupported event type: %s", event.Type())
	}

	fmt.Printf(
		"Candle Received | Symbol=%s | Time=%s\n",
		candleEvent.Candle.Symbol,
		candleEvent.Candle.Timestamp.Format("2006-01-02 15:04:05"),
	)

	return nil
}