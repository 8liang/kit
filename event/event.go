package event

type Type int

type Event struct {
	eventType Type
	payload   any
}

func NewEvent(eventType Type, payload ...any) *Event {
	e := &Event{
		eventType: eventType,
	}
	if len(payload) > 0 {
		e.payload = payload[0]
	}
	return e
}

func (e *Event) Type() Type {
	return e.eventType
}

func (e *Event) Payload() any {
	return e.payload
}
