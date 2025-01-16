package Event

import (
	"BHLayer2Node/paradigm"
	"time"
)

// Event 事件
type Event struct {
	//EpochEvent chan bool
	channel *paradigm.RappaChannel
}

func (e *Event) Start() {
	epochForward := func() {
		timeStart := time.Now()
		for {
			if time.Since(timeStart) >= 10*time.Second {
				e.channel.EpochEvent <- true
				timeStart = time.Now()
			}
		}
	}
	go epochForward()
}

func NewEvent(channel *paradigm.RappaChannel) *Event {
	return &Event{channel: channel}
}
