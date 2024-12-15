package paradigm

import "BHLayer2Node/pb/service"

// CommitSlotItem 节点完成任务后提交
// 2024-12-15 13:04 这里假定正确节点运行的逻辑是，将n个数据块全部分发（出现超时会重新分配直到完成）
// 要考虑恶意节点故意说自己完成了，也就是节点commit的内容需要确认是否正确才能上链
// 这样上链之前可以确认的是，确实可以恢复，但没有办法说明恢复后的数据确实是对的，恶意节点可以伪造一个文件，然后分块
// 这里就需要节点提供证明，可验证计算的证明？或者是质量的证明？
// todo 这里要调研一下，已有的系统是怎么解决文件造假的问题的,zkp太慢了，这部分的容错还有待考虑

type CommitState int

const (
	INVALID   CommitState = iota // 不合法的提交,这里指commitment不通过 todo
	ABORT                        // abort，说明原先节点提交后，投票失败
	JUSTIFIED                    // 就是默认的状态，收到commit后就设置为justified
	FINALIZE                     // 说明确认了，可以上链
)

type CommitSlotItem struct {
	*service.JustifiedSlot
	state CommitState
	//epoch int // 在哪个epoch被提交
	//Commitment SimpleCommitment // 这里简单做一下 todo
	//votes map[int]string // 投票，暂时先不要这个，后续要加上 todo
}

func (c *CommitSlotItem) Record() ScheduleItem {
	return ScheduleItem{
		Size: c.Process,
		NID:  int(c.Nid),
	}
}
func (c *CommitSlotItem) Check() bool {
	// todo
	check := func() bool {
		// 这里就是简单的看一下commitment是否正确
		return true
	}
	if !check() {
		c.SetInvalid()
		return false
	}
	return true
}
func (c *CommitSlotItem) State() CommitState {
	return c.state
}
func (c *CommitSlotItem) SetAbort() {
	c.state = ABORT
}
func (c *CommitSlotItem) SetInvalid() {
	c.state = INVALID
}
func (c *CommitSlotItem) SetDefault() {
	c.state = JUSTIFIED
}
func (c *CommitSlotItem) SetFinalize() {
	c.state = FINALIZE
}
func (c *CommitSlotItem) SetEpoch(e int32) {
	c.Epoch = e
}

// NewCommitSlotItem 默认的commitSlot，状态为justified
func NewCommitSlotItem(slot *service.JustifiedSlot) CommitSlotItem {
	return CommitSlotItem{
		JustifiedSlot: slot,
		state:         JUSTIFIED,
	}
}
