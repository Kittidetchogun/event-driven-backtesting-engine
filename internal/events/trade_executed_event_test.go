package events

import (
    "testing"
    "time"

    "event-driven-backtesting-engine/internal/domain"
)

func TestNewTradeExecutedEvent(t *testing.T) {

    now := time.Now()

    trade := domain.NewTrade(
        1,
        1,
        1,
        "BTCUSDT",
        domain.BuyOrder,
        1,
        50000,
        10,
        now,
    )

    event := NewTradeExecutedEvent(trade)

    if event.Type() != TradeExecutedEventType {
        t.Fatalf("expected %s got %s",
            TradeExecutedEventType,
            event.Type(),
        )
    }

    if event.Trade.TradeID != trade.TradeID {
        t.Fatal("trade mismatch")
    }

    if !event.Timestamp().Equal(now) {
        t.Fatal("timestamp mismatch")
    }
}
