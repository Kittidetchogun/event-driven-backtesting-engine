package events

import (
	"testing"
	"time"

	"event-driven-backtesting-engine/internal/domain"
)

func TestNewSignalGeneratedEvent(t *testing.T) {
	signalTime := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)

	event := NewSignalGeneratedEvent(
		1,
		"BTCUSDT",
		domain.BuyOrder,
		0.5,
		signalTime,
	)

	if event.Type() != SignalGeneratedEventType {
		t.Fatalf("Type() = %q, want %q",
			event.Type(),
			SignalGeneratedEventType,
		)
	}

	if !event.Timestamp().Equal(signalTime) {
		t.Fatalf("Timestamp() = %v, want %v",
			event.Timestamp(),
			signalTime,
		)
	}

	if event.RunID != 1 {
		t.Fatalf("RunID = %d, want %d",
			event.RunID,
			1,
		)
	}

	if event.Symbol != "BTCUSDT" {
		t.Fatalf("Symbol = %q, want %q",
			event.Symbol,
			"BTCUSDT",
		)
	}

	if event.SignalType != domain.BuyOrder {
		t.Fatalf("SignalType = %q, want %q",
			event.SignalType,
			domain.BuyOrder,
		)
	}

	if event.Quantity != 0.5 {
		t.Fatalf("Quantity = %f, want %f",
			event.Quantity,
			0.5,
		)
	}

	if !event.SignalTime.Equal(signalTime) {
		t.Fatalf("SignalTime = %v, want %v",
			event.SignalTime,
			signalTime,
		)
	}
}