package paradigm

import (
	"BHLayer2Node/pb/service"
	"fmt"
)

// CommitSlotItem 节点完成任务后提交
// 2024-12-15 13:04 这里假定正确节点运行的逻辑是，将n个数据块全部分发（出现超时会重新分配直到完成）
// 要考虑恶意节点故意说自己完成了，也就是节点commit的内容需要确认是否正确才能上链
// 这样上链之前可以确认的是，确实可以恢复，但没有办法说明恢复后的数据确实是对的，恶意节点可以伪造一个文件，然后分块
// 这里就需要节点提供证明，可验证计算的证明？或者是质量的证明？
// todo 这里要调研一下，已有的系统是怎么解决文件造假的问题的,zkp太慢了，这部分的容错还有待考虑

type CommitState int

const (
	INVALID   CommitState = iota // 不合法的提交,这里指commitment不通过 todo
	JUSTIFIED                    // 就是默认的状态，收到commit后就设置为justified
	FINALIZE                     // 说明确认了，可以上链 TODO @YZM： 这里目前是只要通过投票就统一标记为FINALIZE，因为这说明commitment是被所有节点认可，至于最后是否能够上链，取决于zkp
)

type CommitSlotItem struct {
	*service.JustifiedSlot
	state       CommitState
	hash        SlotHash
	InvalidType InvalidCommitType
	//epoch int // 在哪个epoch被提交
	//Commitment SimpleCommitment // 这里简单做一下 todo
	//votes map[int]string // 投票，暂时先不要这个，后续要加上 todo
}

type InvalidCommitType = int32

const (
	INVALID_SLOT       InvalidCommitType = iota // 不合法（负数或过大）的slot
	EXPIRE_SLOT                                 // 过期的slot
	INVALID_COMMITMENT                          // 承诺不通过
	VERIFIED_FAILED                             // 异常的存储状态
	// TODO

	UNKNOWN
	NONE
)

func (c *CommitSlotItem) Record() ScheduleItem {
	return ScheduleItem{
		Size:       c.Process,
		NID:        int(c.Nid),
		Commitment: c.Commitment,
		Hash:       c.hash,
	}
}
func (c *CommitSlotItem) Check() bool {
	// todo
	check := func() bool {
		// 这里就是简单的看一下commitment是否正确
		return true
	}
	if !check() {
		c.SetInvalid(INVALID_COMMITMENT)
		return false
	}
	return true
}
func (c *CommitSlotItem) State() CommitState {
	return c.state
}

//func (c *CommitSlotItem) SetAbort() {
//	c.state = ABORT
//}

func (c *CommitSlotItem) SetInvalid(t InvalidCommitType) {
	c.state = INVALID
	c.InvalidType = t
}
func (c *CommitSlotItem) SetDefault() {
	c.state = JUSTIFIED
	c.InvalidType = NONE
}
func (c *CommitSlotItem) SetFinalize() {
	c.SetDefault()
	c.state = FINALIZE
}
func (c *CommitSlotItem) SetEpoch(e int32) {
	c.Epoch = e
	c.JustifiedSlot.Epoch = e
}
func (c *CommitSlotItem) computeHash() {
	// todo 这里的哈希可以修改
	generateHash := func(sign string, slot int, node int) SlotHash {
		return fmt.Sprintf("%s_%d_%d", sign, slot, node)
	}
	c.hash = generateHash(c.Sign, int(c.Slot), int(c.Nid))

}
func (c *CommitSlotItem) SlotHash() SlotHash {
	return c.hash
}

//func (c *CommitSlotItem) HasProof() bool {
//	return c.proof
//}

// NewCommitSlotItem 默认的commitSlot，状态为justified
func NewCommitSlotItem(slot *service.JustifiedSlot) CommitSlotItem {
	s := CommitSlotItem{
		JustifiedSlot: slot,
		state:         JUSTIFIED,
		InvalidType:   NONE,
		//proof:         false,
	}
	s.computeHash()
	return s
}
func NewFakeCommitSlotItem(hash SlotHash) CommitSlotItem {
	s := CommitSlotItem{
		JustifiedSlot: nil,
		state:         FINALIZE,
		hash:          hash,
		InvalidType:   NONE,
	}
	return s
}
