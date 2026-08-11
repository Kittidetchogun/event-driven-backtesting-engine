package events

import (
	"testing"
	"time"
)

func TestDispatcher_RegisterAndDispatch(t *testing.T) {

	dispatcher := NewEventDispatcher()

	called := false

	dispatcher.Register("TestEvent", func(e Event) error {
		called = true
		return nil
	})

	event := NewBaseEvent("TestEvent", time.Now())

	if err := dispatcher.Dispatch(event); err != nil {
		t.Fatalf("Dispatch() error = %v", err)
	}

	if !called {
		t.Fatal("handler was not called")
	}
}

func TestDispatcher_DispatchUnknownEvent(t *testing.T) {

	dispatcher := NewEventDispatcher()

	event := NewBaseEvent("UnknownEvent", time.Now())

	if err := dispatcher.Dispatch(event); err == nil {
		t.Fatal("expected error for unknown event")
	}
}

func TestDispatcher_RegisterMultipleHandlers(t *testing.T) {

	dispatcher := NewEventDispatcher()

	count := 0

	dispatcher.Register("TestEvent", func(e Event) error {
		count++
		return nil
	})

	dispatcher.Register("TestEvent", func(e Event) error {
		count++
		return nil
	})

	event := NewBaseEvent("TestEvent", time.Now())

	if err := dispatcher.Dispatch(event); err != nil {
		t.Fatalf("Dispatch() error = %v", err)
	}

	if count != 2 {
		t.Fatalf("handler count = %d, want 2", count)
	}
}

func TestDispatcher_NilEvent(t *testing.T) {

	dispatcher := NewEventDispatcher()

	if err := dispatcher.Dispatch(nil); err == nil {
		t.Fatal("expected error for nil event")
	}
}

func TestEventDispatcher_DispatchMultipleHandlers(t *testing.T) {
	dispatcher := NewEventDispatcher()

	count := 0

	handler1 := func(event Event) error {
		count++
		return nil
	}

	handler2 := func(event Event) error {
		count++
		return nil
	}

	dispatcher.Register(
		BacktestCompletedEventType,
		handler1,
	)

	dispatcher.Register(
		BacktestCompletedEventType,
		handler2,
	)

	event := NewBacktestCompletedEvent(
		1,
		"EMA Cross",
		"BTCUSDT",
		"1d",
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
		0.10,
		0.60,
		1.20,
		-0.05,
	)

	if err := dispatcher.Dispatch(event); err != nil {
		t.Fatalf("Dispatch() error = %v", err)
	}

	if count != 2 {
		t.Fatalf("handlers called = %d, want 2", count)
	}
}
