package Event

import "time"

// Event 事件
type Event struct {
	EpochEvent chan bool
}

func (e *Event) Start() {
	epochForward := func() {
		timeStart := time.Now()
		for {
			if time.Since(timeStart) >= 10*time.Second {
				e.EpochEvent <- true
				timeStart = time.Now()
			}
		}
	}
	go epochForward()
}

func NewEvent(epoch chan bool) *Event {
	return &Event{EpochEvent: epoch}
}
