package Task

import (
	"BHLayer2Node/paradigm"
)

type EpochRecord struct {
	//commits   []*service.JustifiedSlot
	//finalizes []*service.JustifiedSlot
	//invalids  []*service.JustifiedSlot
	commits   map[paradigm.SlotHash]paradigm.SlotCommitment    // 在这个epoch里commit的slot，目前状态为justified, map的内容为commitment
	finalizes map[paradigm.SlotHash]paradigm.SlotCommitment    // 在这个epoch里已经确认finalized的，节点在收到这个后可以确认落盘
	invalids  map[paradigm.SlotHash]paradigm.InvalidCommitType // 在这个epoch里被检测出的问题slot, 节点可以根据这个删、改
}

func (r *EpochRecord) commit(slot paradigm.CommitSlotItem) {
	check := func() bool {
		// 这里判断slot的合法性 todo
		if slot.State() == paradigm.INVALID {
			return false
		}
		return slot.Check() // 除了这个可能还有别的逻辑
	}
	if check() {
		r.commits[slot.SlotHash()] = slot.Commitment
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
		//r.finalizes = append(r.finalizes, slot.JustifiedSlot)
		r.finalizes[slot.SlotHash()] = slot.Commitment
	} else {
		slot.SetInvalid(paradigm.UNKNOWN) // 这里目前没用，甚至不会进入这里 todo

	}
}
func (r *EpochRecord) abort(slot paradigm.CommitSlotItem, reason paradigm.InvalidCommitType) {
	check := func() bool {
		return true
	}
	if check() {
		slot.SetInvalid(reason)
		r.invalids[slot.SlotHash()] = reason
	} else {
		// TODO
	}

}
func (r *EpochRecord) Refresh() {
	r.commits = make(map[paradigm.SlotHash]paradigm.SlotCommitment)
	r.finalizes = make(map[paradigm.SlotHash]paradigm.SlotCommitment)
	r.invalids = make(map[paradigm.SlotHash]paradigm.InvalidCommitType)
}

func NewEpochRecord() *EpochRecord {
	return &EpochRecord{
		commits:   make(map[paradigm.SlotHash]paradigm.SlotCommitment),
		finalizes: make(map[paradigm.SlotHash]paradigm.SlotCommitment),
		invalids:  make(map[paradigm.SlotHash]paradigm.InvalidCommitType),
	}
}
