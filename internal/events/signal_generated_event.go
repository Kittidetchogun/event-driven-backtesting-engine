package events

import (
	"time"

	"event-driven-backtesting-engine/internal/domain"
)

const SignalGeneratedEventType = "SignalGeneratedEvent"

type SignalGeneratedEvent struct {
	BaseEvent

	RunID      int
	Symbol     string
	SignalType domain.OrderSide
	Quantity   float64
	SignalTime time.Time
}

func NewSignalGeneratedEvent(
	runID int,
	symbol string,
	signalType domain.OrderSide,
	quantity float64,
	signalTime time.Time,
) SignalGeneratedEvent {
	return SignalGeneratedEvent{
		BaseEvent:  NewBaseEvent(SignalGeneratedEventType, signalTime),
		RunID:      runID,
		Symbol:     symbol,
		SignalType: signalType,
		Quantity:   quantity,
		SignalTime: signalTime,
	}
}