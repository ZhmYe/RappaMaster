package Event

import (
	"RappaMaster/channel"
	"time"
)

// Event 事件
type Event struct {
	//EpochEvent chan bool
	channel *channel.RappaChannel
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

func NewEvent(channel *channel.RappaChannel) *Event {
	return &Event{channel: channel}
}
