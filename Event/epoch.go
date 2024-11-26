package Event

import "time"

// Event 事件
type Event struct {
	EpochEvent chan bool
}

func (e *Event) Start() {
	epochForward := func() {
		for {
			time.Sleep(10 * time.Second)
			e.EpochEvent <- true
		}
	}
	go epochForward()
}

func NewEvent(epoch chan bool) *Event {
	return &Event{EpochEvent: epoch}
}
