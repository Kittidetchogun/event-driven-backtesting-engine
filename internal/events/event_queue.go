package events

import "sync"

type EventQueue struct {
	mu    sync.Mutex
	items []Event
}

func NewEventQueue() *EventQueue {
	return &EventQueue{
		items: make([]Event, 0),
	}
}

func (q *EventQueue) Push(event Event) {
	if event == nil {
		return
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	q.items = append(q.items, event)
}

func (q *EventQueue) Pop() (Event, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) == 0 {
		return nil, false
	}

	event := q.items[0]
	q.items[0] = nil
	q.items = q.items[1:]

	return event, true
}

func (q *EventQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.items)
}

func (q *EventQueue) IsEmpty() bool {
	return q.Len() == 0
}