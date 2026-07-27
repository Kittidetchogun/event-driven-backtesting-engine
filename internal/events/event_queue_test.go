package events

import (
	"testing"
	"time"
)

func TestEventQueue_PushPopFIFO(t *testing.T) {
	q := NewEventQueue()

	t1 := time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 7, 1, 9, 5, 0, 0, time.UTC)
	t3 := time.Date(2026, 7, 1, 9, 10, 0, 0, time.UTC)

	e1 := NewBaseEvent("Event1", t1)
	e2 := NewBaseEvent("Event2", t2)
	e3 := NewBaseEvent("Event3", t3)

	q.Push(e1)
	q.Push(e2)
	q.Push(e3)

	if got := q.Len(); got != 3 {
		t.Fatalf("Len() = %d, want %d", got, 3)
	}

	event, ok := q.Pop()
	if !ok {
		t.Fatal("Pop() returned ok=false, want true")
	}
	if event.Type() != "Event1" {
		t.Fatalf("first Pop() Type() = %q, want %q", event.Type(), "Event1")
	}
	if !event.Timestamp().Equal(t1) {
		t.Fatalf("first Pop() Timestamp() = %v, want %v", event.Timestamp(), t1)
	}

	event, ok = q.Pop()
	if !ok {
		t.Fatal("second Pop() returned ok=false, want true")
	}
	if event.Type() != "Event2" {
		t.Fatalf("second Pop() Type() = %q, want %q", event.Type(), "Event2")
	}
	if !event.Timestamp().Equal(t2) {
		t.Fatalf("second Pop() Timestamp() = %v, want %v", event.Timestamp(), t2)
	}

	event, ok = q.Pop()
	if !ok {
		t.Fatal("third Pop() returned ok=false, want true")
	}
	if event.Type() != "Event3" {
		t.Fatalf("third Pop() Type() = %q, want %q", event.Type(), "Event3")
	}
	if !event.Timestamp().Equal(t3) {
		t.Fatalf("third Pop() Timestamp() = %v, want %v", event.Timestamp(), t3)
	}

	if !q.IsEmpty() {
		t.Fatal("queue should be empty after popping all events")
	}
	if got := q.Len(); got != 0 {
		t.Fatalf("Len() after popping all events = %d, want 0", got)
	}
}

func TestEventQueue_LenAndIsEmpty(t *testing.T) {
	q := NewEventQueue()

	if !q.IsEmpty() {
		t.Fatal("new queue should be empty")
	}
	if got := q.Len(); got != 0 {
		t.Fatalf("new queue Len() = %d, want 0", got)
	}

	q.Push(NewBaseEvent("Event1", time.Now().UTC()))

	if q.IsEmpty() {
		t.Fatal("queue should not be empty after Push()")
	}
	if got := q.Len(); got != 1 {
		t.Fatalf("Len() after one Push() = %d, want 1", got)
	}

	_, ok := q.Pop()
	if !ok {
		t.Fatal("Pop() returned ok=false, want true")
	}

	if !q.IsEmpty() {
		t.Fatal("queue should be empty after Pop()")
	}
	if got := q.Len(); got != 0 {
		t.Fatalf("Len() after Pop() = %d, want 0", got)
	}
}

func TestEventQueue_PopEmptyReturnsFalse(t *testing.T) {
	q := NewEventQueue()

	event, ok := q.Pop()
	if ok {
		t.Fatal("Pop() on empty queue returned ok=true, want false")
	}
	if event != nil {
		t.Fatalf("Pop() on empty queue returned event=%v, want nil", event)
	}
}

func TestEventQueue_PushNilIgnored(t *testing.T) {
	q := NewEventQueue()

	q.Push(nil)

	if !q.IsEmpty() {
		t.Fatal("queue should remain empty after pushing nil")
	}
	if got := q.Len(); got != 0 {
		t.Fatalf("Len() after pushing nil = %d, want 0", got)
	}
}