package events

import (
	"testing"
	"time"

	"event-driven-backtesting-engine/internal/domain"
)

func TestPrintConsumer_Consume(t *testing.T) {

	consumer := NewPrintConsumer()

	candle := domain.Candle{
		Symbol:    "BTCUSDT",
		Timeframe: "1d",
		Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	event := NewCandleReceivedEvent(candle)

	if err := consumer.Consume(event); err != nil {
		t.Fatalf("Consume() error = %v", err)
	}
}

func TestPrintConsumer_NilEvent(t *testing.T) {

	consumer := NewPrintConsumer()

	if err := consumer.Consume(nil); err == nil {
		t.Fatal("expected error")
	}
}

func TestPrintConsumer_UnsupportedEvent(t *testing.T) {

	consumer := NewPrintConsumer()

	event := NewBaseEvent("UnknownEvent", time.Now())

	if err := consumer.Consume(event); err == nil {
		t.Fatal("expected error")
	}
}