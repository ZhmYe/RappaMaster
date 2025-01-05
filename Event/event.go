package Event

import (
	"BHLayer2Node/paradigm"
	"time"
)

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

func NewEvent(channel *paradigm.RappaChannel) *Event {
	return &Event{EpochEvent: channel.EpochEvent}
}
