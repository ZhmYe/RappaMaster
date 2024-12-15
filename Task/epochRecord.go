package Task

import (
	"BHLayer2Node/paradigm"
	"BHLayer2Node/pb/service"
)

type EpochRecord struct {
	commits   []*service.JustifiedSlot
	finalizes []*service.JustifiedSlot
}

func (r *EpochRecord) commit(slot paradigm.CommitSlotItem) {
	check := func() bool {
		// 这里判断slot的合法性 todo
		if slot.State() == paradigm.INVALID || slot.State() == paradigm.ABORT {
			return false
		}
		return slot.Check() // 除了这个可能还有别的逻辑
	}
	if check() {
		r.commits = append(r.commits, slot.JustifiedSlot)
	}
}
func (r *EpochRecord) finalize(slot paradigm.CommitSlotItem) {
	check := func() bool {
		// 这里判断合法性 todo
		if slot.State() != paradigm.FINALIZE {
			return false
		}
		return true
	}
	if check() {
		//slot.SetFinalize() // finalize
		r.finalizes = append(r.finalizes, slot.JustifiedSlot)
	} else {
		slot.SetAbort() // 这里目前没用，甚至不会进入这里 todo
	}
}
func (r *EpochRecord) Refresh() {
	r.commits = make([]*service.JustifiedSlot, 0)
	r.finalizes = make([]*service.JustifiedSlot, 0)
}

func NewEpochRecord() *EpochRecord {
	return &EpochRecord{
		commits:   make([]*service.JustifiedSlot, 0),
		finalizes: make([]*service.JustifiedSlot, 0),
	}
}
