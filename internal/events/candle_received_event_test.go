package events

import (
	"testing"
	"time"

	"event-driven-backtesting-engine/internal/domain"
)

var _ Event = CandleReceivedEvent{}

func TestNewCandleReceivedEvent(t *testing.T) {
	candle := domain.Candle{
		Timestamp: time.Date(2026, 7, 27, 12, 0, 0, 0, time.UTC),
		Symbol:    "BTCUSDT",
		Timeframe: "1d",
		Open:      100,
		High:      110,
		Low:       90,
		Close:     105,
		Volume:    42,
	}

	event := NewCandleReceivedEvent(candle)

	if event.Type() != CandleReceivedEventType {
		t.Fatalf("Type() = %q, want %q", event.Type(), CandleReceivedEventType)
	}

	if !event.Timestamp().Equal(candle.Timestamp) {
		t.Fatalf("Timestamp() = %s, want %s", event.Timestamp(), candle.Timestamp)
	}

	if event.Candle != candle {
		t.Fatalf("Candle = %+v, want %+v", event.Candle, candle)
	}
}
