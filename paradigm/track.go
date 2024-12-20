package paradigm

//
//type TaskSlotRecord struct {
//	sign     string
//	slot     int
//	schedule TaskSchedule
//}

type PendingCommitSlotTrack struct {
	*CommitSlotItem
	hasVerifiedProof bool
	hasWonVote       bool
}

func (t *PendingCommitSlotTrack) Check() bool {
	return t.hasWonVote && t.hasVerifiedProof
}
func NewPendingCommitSlotTrack(item *CommitSlotItem, needProof bool) *PendingCommitSlotTrack {
	return &PendingCommitSlotTrack{
		CommitSlotItem:   item,
		hasVerifiedProof: needProof, // 如果不需要可信证明，那么就是完成了
		hasWonVote:       false,
	}
}
func (t *PendingCommitSlotTrack) ReceiveProof() {
	t.hasVerifiedProof = true
}
func (t *PendingCommitSlotTrack) WonVote() {
	t.hasWonVote = true
}
