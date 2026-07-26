package events

import "time"

// Event is the minimal contract for every event flowing through the
// event-driven backtesting engine.
type Event interface {
	Type() string
	Timestamp() time.Time
}

// BaseEvent stores the common metadata shared by all engine events.
// Concrete events should embed BaseEvent and add their domain-specific payload.
type BaseEvent struct {
	eventType string
	timestamp time.Time
}

var _ Event = BaseEvent{}

// NewBaseEvent creates event metadata with a stable event type and timestamp.
func NewBaseEvent(eventType string, timestamp time.Time) BaseEvent {
	return BaseEvent{
		eventType: eventType,
		timestamp: timestamp,
	}
}

// Type returns the event category used by queues and consumers to route events.
func (e BaseEvent) Type() string {
	return e.eventType
}

// Timestamp returns the logical time associated with the event.
func (e BaseEvent) Timestamp() time.Time {
	return e.timestamp
}
