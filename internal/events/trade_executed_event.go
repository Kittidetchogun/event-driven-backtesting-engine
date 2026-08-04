package events

import (
    "event-driven-backtesting-engine/internal/domain"
)

const TradeExecutedEventType = "TradeExecutedEvent"

type TradeExecutedEvent struct {
    BaseEvent

    Trade domain.Trade
}

func NewTradeExecutedEvent(trade domain.Trade) TradeExecutedEvent {
    return TradeExecutedEvent{
        BaseEvent: NewBaseEvent(
            TradeExecutedEventType,
            trade.ExecutedTime,
        ),
        Trade: trade,
    }
}
