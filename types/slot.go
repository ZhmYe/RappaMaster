package types

import (
	"RappaMaster/paradigm"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/sha3"
	"time"
)

type SlotStatus int

const (
	Commit SlotStatus = iota // is commit
	Justified
	Finalized
)

type ScheduleSlot struct {
	NodeID    int
	Task      string // sign
	Model     paradigm.SupportModelType
	Size      int64
	timestamp time.Time
}

func (s *ScheduleSlot) SlotHash() string {
	hasher := sha3.New256()
	hasher.Write([]byte(s.Task))
	hasher.Write([]byte(paradigm.ModelTypeToString(s.Model)))
	hasher.Write([]byte(fmt.Sprintf("%d", s.Size)))
	hasher.Write([]byte(fmt.Sprintf("%d", time.Now().Unix())))
	return hex.EncodeToString(hasher.Sum(nil))
}

func NewScheduleSlot(nodeID int, task string, size int64) ScheduleSlot {
	return ScheduleSlot{
		NodeID:    nodeID,
		Task:      task,
		Size:      size,
		timestamp: time.Now(),
	}
}
