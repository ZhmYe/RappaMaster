package types

import (
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/sha3"
	"time"
)

// SlotHash 调度的标识
type SlotHash []byte

func (sh SlotHash) String() string {
	return hex.EncodeToString(sh)
}

// SlotCommitment slot数据完整性，signature签名针对commitment
type SlotCommitment = []byte

type SlotStatus int

const (
	Committed SlotStatus = iota // slot has been committed by node
	Justified                   // slot has been verified and justified in epoch
	Finalized                   // slot has been finalized when task completes
)

func (s SlotStatus) String() string {
	switch s {
	case Committed:
		return "committed"
	case Justified:
		return "justified"
	case Finalized:
		return "finalized"
	default:
		return "unknown"
	}
}

type ScheduleSlot struct {
	NodeID    int
	Task      string // sign
	Model     SupportModelType
	Size      int64
	timestamp time.Time
}

func (s *ScheduleSlot) SlotHash() SlotHash {
	hasher := sha3.New256()
	hasher.Write([]byte(s.Task))
	hasher.Write([]byte(ModelTypeToString(s.Model)))
	hasher.Write([]byte(fmt.Sprintf("%d", s.Size)))
	hasher.Write([]byte(fmt.Sprintf("%d", s.timestamp.Unix())))
	return hasher.Sum(nil)
}

func (s *ScheduleSlot) String() string {
	return s.SlotHash().String()
}

func NewScheduleSlot(nodeID int, task string, size int64) ScheduleSlot {
	return ScheduleSlot{
		NodeID:    nodeID,
		Task:      task,
		Size:      size,
		timestamp: time.Now(),
	}
}
