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
	INVALID      CommitState = iota // 不合法的提交,这里指commitment不通过 todo
	UNDETERMINED                    // 节点提交后就是undetermined状态
	JUSTIFIED                       // 投票完成后设置为justified
	FINALIZE                        // 说明确认了，可以上链
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

// 用于定义无效提交的错误及错误消息
type InvalidCommitError struct {
	Error        InvalidCommitType
	ErrorMessage string
}

const (
	INVALID_SLOT       InvalidCommitType = iota // 不合法（负数或过大）的slot
	EXPIRE_SLOT                                 // 过期的slot
	INVALID_COMMITMENT                          // 承诺不通过
	VERIFIED_FAILED                             // 异常的存储状态
	DOWN_FAILED                                 //宕机无法恢复
	UNKNOWN
	NONE
)

func InvalidCommitTypeToString(i InvalidCommitType) string {
	switch i {
	case INVALID_SLOT:
		return "INVALID_SLOT"
	case EXPIRE_SLOT:
		return "EXPIRE_SLOT"
	case INVALID_COMMITMENT:
		return "INVALID_COMMITMENT"
	case VERIFIED_FAILED:
		return "VERIFIED_FAILED"
	case DOWN_FAILED:
		return "DOWN_FAILED"
	case UNKNOWN:
		return "UNKNOWN"
	case NONE:
		return "NONE"
	default:
		return "UNDEFINED"
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
	c.state = UNDETERMINED
	c.InvalidType = NONE
}
func (c *CommitSlotItem) SetFinalize() {
	c.SetDefault()
	c.state = FINALIZE
}
func (c *CommitSlotItem) SetEpoch(e int32) {
	c.Epoch = e
	//c.JustifiedSlot.Epoch = e
}

func (c *CommitSlotItem) SetHash(hash SlotHash) {
	c.hash = hash
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

// NewCommitSlotItem 默认的commitSlot，状态为undetermined
func NewCommitSlotItem(slot *service.JustifiedSlot) CommitSlotItem {
	s := CommitSlotItem{
		JustifiedSlot: slot,
		state:         UNDETERMINED,
		InvalidType:   NONE,
		//proof:         false,
	}
	//s.hash = s.Hash // todo 这里简单这样写一下
	//s.computeHash()
	return s
}
func NewFakeCommitSlotItem(hash SlotHash) CommitSlotItem {
	s := CommitSlotItem{
		JustifiedSlot: nil,
		state:         JUSTIFIED,
		hash:          hash,
		InvalidType:   NONE,
	}
	return s
}

func NewInvalidCommitError(t InvalidCommitType, errorMessage string) InvalidCommitError {
	return InvalidCommitError{
		Error:        t,
		ErrorMessage: errorMessage,
	}
}
