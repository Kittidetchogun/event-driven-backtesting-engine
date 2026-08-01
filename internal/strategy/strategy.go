package strategy

import (
    "event-driven-backtesting-engine/internal/domain"
    "event-driven-backtesting-engine/internal/events"
)

type Strategy interface {
    Name() string
    OnData(candle domain.Candle)
    Initialize() error
    Reset()
    SetDispatcher(dispatcher *events.EventDispatcher)
}
