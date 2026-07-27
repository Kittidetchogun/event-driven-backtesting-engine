package events

import "fmt"

type EventHandler func(Event) error

type EventDispatcher struct {
	handlers map[string][]EventHandler
}

func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{
		handlers: make(map[string][]EventHandler),
	}
}

func (d *EventDispatcher) Register(eventType string, handler EventHandler) {
	if handler == nil {
		return
	}

	d.handlers[eventType] = append(d.handlers[eventType], handler)
}

func (d *EventDispatcher) Dispatch(event Event) error {
	if event == nil {
		return fmt.Errorf("event is nil")
	}

	handlers, ok := d.handlers[event.Type()]
	if !ok {
		return fmt.Errorf("no handler registered for event type %s", event.Type())
	}

	for _, handler := range handlers {
		if err := handler(event); err != nil {
			return err
		}
	}

	return nil
}