package paradigm

import (
	"fmt"
)

type SynthTaskTrackItem struct {
	*UnprocessedTask
	Total      int32
	History    int32
	IsReliable bool
}

func (t *SynthTaskTrackItem) Commit(item CommitSlotItem) error {
	if item.State() != FINALIZE {
		return fmt.Errorf("the commit Slot is not finalized") // 只能提交finalized的，因为已经通过投票了所以不需要check
	}
	//t.records[slot.Slot] = slotRecord
	t.UnprocessedTask.Process(item.Process)
	t.History += item.Process
	//Log("TRACKER", fmt.Sprintf("Task Track %s process %d by node %d, slotHash: %s, total: %d, history: %d, unprocessedSize: %d", item.Sign, item.Process, item.Nid, item.SlotHash(), t.Total, t.History, t.Size))
	return nil
}
func (t *SynthTaskTrackItem) IsFinish() bool {
	return t.Size <= 0
}
func (t *SynthTaskTrackItem) Next() UnprocessedTask {
	return *t.UnprocessedTask
}

// UnprocessedTask 待调度的任务,也就是任务track: 1. 新建的任务; 2. 未完成过期的任务
type UnprocessedTask struct {
	TaskID TaskHash // 任务标识
	// 这里不需要记录scheduleID，记录在Scheduler内部即可
	Size     int32                  // data size
	SlotSize int32                  //指定slot的大小
	Model    SupportModelType       // 模型名称
	Params   map[string]interface{} // 不确定的模型参数
}

func (t *UnprocessedTask) Process(size int32) {

	t.Size -= size
	if t.Size < 0 {
		t.Size = 0
	}

}

type PendingCommitSlotTrack struct {
	*CommitSlotItem
	IsFinalized      int32
	hasVerifiedProof bool
	hasWonVote       bool
}

func (t *PendingCommitSlotTrack) Check() bool {
	return t.hasVerifiedProof && t.hasWonVote
}
func NewPendingCommitSlotTrack(item *CommitSlotItem, isReliable bool) *PendingCommitSlotTrack {
	return &PendingCommitSlotTrack{
		CommitSlotItem:   item,
		hasVerifiedProof: isReliable, // 如果不需要可信证明，那么就是完成了
		hasWonVote:       false,
		IsFinalized:      0,
	}
}
func (t *PendingCommitSlotTrack) ReceiveProof() {
	t.hasVerifiedProof = true
}

func (t *PendingCommitSlotTrack) WonVote() {
	t.hasWonVote = true
}
