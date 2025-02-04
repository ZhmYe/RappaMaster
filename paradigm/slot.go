package paradigm

type SlotStatus int

const (
	Finished SlotStatus = iota
	Processing
	Failed
)

// Slot 一个具体的节点合成任务实例
type Slot struct {
	SlotID       SlotHash
	ScheduleSize int32           // 调度的数量，以KB为单位
	Status       SlotStatus      // 完成状态
	err          string          // 错误信息
	CommitSlot   *CommitSlotItem // 提交上来的commitSlot
}

func (s *Slot) SetError(errorMessage string) {
	s.err = errorMessage
	s.Status = Failed
}

// Commit 将节点提交的结果commit，这里不做完整性等校验，在外面校验完才放到这里
func (s *Slot) Commit(commitSlot *CommitSlotItem) {
	s.CommitSlot = commitSlot
	s.Status = Finished // 这里不区分是否全部做完，不允许多次提交 todo
}

func NewSlot(slotID SlotHash, schedule int32) *Slot {
	return &Slot{
		SlotID:       slotID,
		ScheduleSize: schedule,
		Status:       Processing,
		err:          "",
		CommitSlot:   nil,
	}
}
