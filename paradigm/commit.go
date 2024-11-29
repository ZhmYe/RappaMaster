package paradigm

// CommitSlotItem 节点完成任务后提交
type CommitSlotItem struct {
	Nid     int
	Process int
	Sign    string
	Slot    int
	//Commitment SimpleCommitment // 这里简单做一下
}

func (c *CommitSlotItem) Record() ScheduleItem {
	return ScheduleItem{
		Size: c.Process,
		NID:  c.Nid,
	}
}
