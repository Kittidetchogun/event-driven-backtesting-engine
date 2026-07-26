package events

import (
	"testing"
	"time"
)

func TestBaseEvent(t *testing.T) {
	timestamp := time.Date(2026, 7, 27, 12, 0, 0, 0, time.UTC)
	event := NewBaseEvent("test.event", timestamp)

	if event.Type() != "test.event" {
		t.Fatalf("Type() = %q, want %q", event.Type(), "test.event")
	}

	if !event.Timestamp().Equal(timestamp) {
		t.Fatalf("Timestamp() = %s, want %s", event.Timestamp(), timestamp)
	}
}
