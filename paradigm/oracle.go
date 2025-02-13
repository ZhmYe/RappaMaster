package paradigm

import (
	"BHLayer2Node/LogWriter"
	"fmt"

	"github.com/FISCO-BCOS/go-sdk/v3/types"
)

// =============== 以下是Reference部分==============

type RFType int

const (
	InitTaskTx = iota
	EpochTx
	SlotTX
)

// DevReference 指代一个txMap得到的结果
type DevReference struct {
	TxHash      string
	TxReceipt   types.Receipt // 以上是交易信息
	TxBlockHash string
	Rf          RFType // 类型
	// 如果是InitTask，那么就是一个交易->TaskID，没有额外信息
	// 如果是EpochTx，那么就是一个交易->EpochID，没有额外信息
	// 如是果SlotTx, 那么需要包含两类信息
	// 1. Slot所在epoch; 2. Slot所在Task
	TaskID  TaskHash
	EpochID int32
	//ScheduleID ScheduleHash
}

// CommitRecord 每个commitRecord对应一个完成finalize的commitSlotItem，对应一笔TaskProcessTransaction
type CommitRecord struct {
	*CommitSlotItem
	TxReceipt *types.Receipt // 交易回执
	TxID      int            // 这个交易的id
}

func (r *CommitRecord) Print() {
	fmt.Printf("		CommitSlot Hash: %s\n", r.hash)
	fmt.Printf("		Commit Size: %d\n", r.Process)
	fmt.Printf("		Commit Commitment: %s\n", r.Commitment)
	fmt.Printf("		Commit Epoch: %d\n", r.Epoch)
	fmt.Printf("		Commit Sign: %s\n", r.Sign)
	fmt.Printf("		Commit Slot: %d\n", r.Slot)
	fmt.Printf("		Commit State: %d\n", r.State())
}
func NewCommitRecord(ptx *PackedTransaction) *CommitRecord {
	switch ptx.Tx.(type) {
	case *TaskProcessTransaction:
		return &CommitRecord{
			CommitSlotItem: ptx.Tx.Blob().(*CommitSlotItem),
			TxReceipt:      ptx.Receipt,
			TxID:           ptx.Id,
		}
	default:
		panic("A CommitRecord should be init from an TaskProcessTransaction!!!")
	}
}

// =================== 以下是epoch部分=========================

type DevEpoch struct {
	EpochID     int32
	Process     int32
	Commits     []*Slot
	Justifieds  []*Slot
	Finalizes   []*Slot
	Invalids    []*Slot
	InitTasks   []*Task
	TxReceipt   *types.Receipt // 交易上链后会有一个对应的receipt
	TxID        int            // 交易ID，用于在Dev中定位交易
	TxBlockHash string
}

//func NewDevEpoch(ptx *PackedTransaction) *DevEpoch {
//	switch ptx.Tx.(type) {
//	case *EpochRecordTransaction:
//		return &DevEpoch{
//			EpochRecord: ptx.Tx.Blob().(*EpochRecord),
//			TxReceipt:   ptx.Receipt,
//			TxID:        ptx.Id,
//		}
//	default:
//		panic("A DevEpoch should be init from an EpochRecordTransaction!!!")
//	}
//}

/*** EpochRecord 用于记录一个epoch内情况，由TaskManager更新***/

type EpochRecord struct {
	//commits   []*service.JustifiedSlot
	//finalizes []*service.JustifiedSlot
	//invalids  []*service.JustifiedSlot
	Id         int                         // Epoch id
	Commits    map[SlotHash]SlotCommitment // 在这个epoch里commit的slot，目前状态为undetermined, map的内容为commitment
	Justifieds map[SlotHash]SlotCommitment
	Finalizes  map[SlotHash]SlotCommitment    // 在这个epoch里已经确认finalized的，节点在收到这个后可以确认落盘
	Invalids   map[SlotHash]InvalidCommitType // 在这个epoch里被检测出的问题slot, 节点可以根据这个删、改
	Tasks      map[string]int32               // 新收到的任务sign, 对应的数据量
	Process    int32                          // 一共处理了多少
}

func (r *EpochRecord) UpdateTask(task *Task) {
	if _, exist := r.Tasks[task.Sign]; exist {
		panic("Repeat Epoch Sign!!!")
	}
	r.Tasks[task.Sign] = task.Size
}
func (r *EpochRecord) Commit(slot *CommitSlotItem) {
	check := func() bool {
		// 这里判断slot的合法性 todo
		if slot.State() == INVALID {
			return false
		}
		return slot.Check() // 除了这个可能还有别的逻辑
	}
	if check() {
		r.Commits[slot.SlotHash()] = slot.Commitment
	}
}
func (r *EpochRecord) Justified(slot *CommitSlotItem) {
	check := func() bool {
		if slot.State() != JUSTIFIED {
			return false
		}
		return true
	}
	if check() {
		//fmt.Println(slot.State(), len(r.Justifieds))
		r.Justifieds[slot.SlotHash()] = slot.Commitment
	} else {
		slot.SetInvalid(UNKNOWN) // TODO
	}
}
func (r *EpochRecord) Finalize(slot *CommitSlotItem) {
	check := func() bool {
		// 这里判断合法性 todo
		if slot.State() != FINALIZE {
			return false
		}
		return true
	}
	if check() {
		//slot.SetFinalize() // finalize
		//r.finalizes = append(r.finalizes, slot.JustifiedSlot)
		r.Finalizes[slot.SlotHash()] = slot.Commitment
		r.Process += slot.Process
	} else {
		slot.SetInvalid(UNKNOWN) // 这里目前没用，甚至不会进入这里 todo

	}
}
func (r *EpochRecord) Abort(slot *CommitSlotItem, reason InvalidCommitType) {
	check := func() bool {
		return true
	}
	if check() {
		slot.SetInvalid(reason)
		r.Invalids[slot.SlotHash()] = reason
	} else {
		// TODO
	}

}
func (r *EpochRecord) Refresh() {
	r.Id++
	r.Process = 0
	r.Commits = make(map[SlotHash]SlotCommitment)
	r.Justifieds = make(map[SlotHash]SlotCommitment)
	r.Finalizes = make(map[SlotHash]SlotCommitment)
	r.Invalids = make(map[SlotHash]InvalidCommitType)
}
func (r *EpochRecord) Echo() {
	LogWriter.Log("EPOCH", fmt.Sprintf("Epoch %d Record: ", r.Id))
	LogWriter.Log("EPOCH", fmt.Sprintf("	Commits: %v", r.Commits))
	LogWriter.Log("EPOCH", fmt.Sprintf("	Justifieds: %v", r.Justifieds))
	LogWriter.Log("EPOCH", fmt.Sprintf("	Finalizeds: %v", r.Finalizes))
	LogWriter.Log("EPOCH", fmt.Sprintf("	Invalids: %v", r.Invalids))
}
func NewEpochRecord() *EpochRecord {
	return &EpochRecord{
		Id:         0,
		Commits:    make(map[SlotHash]SlotCommitment),
		Finalizes:  make(map[SlotHash]SlotCommitment),
		Justifieds: make(map[SlotHash]SlotCommitment),
		Invalids:   make(map[SlotHash]InvalidCommitType),
		Tasks:      make(map[string]int32),
	}
}
