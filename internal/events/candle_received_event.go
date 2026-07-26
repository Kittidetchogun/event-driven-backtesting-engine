package events

import "event-driven-backtesting-engine/internal/domain"

const CandleReceivedEventType = "CandleReceivedEvent"

type CandleReceivedEvent struct {
	BaseEvent
	Candle domain.Candle
}

func NewCandleReceivedEvent(candle domain.Candle) CandleReceivedEvent {
	return CandleReceivedEvent{
		BaseEvent: NewBaseEvent(CandleReceivedEventType, candle.Timestamp),
		Candle:    candle,
	}
}
