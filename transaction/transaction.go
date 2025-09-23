package transaction

import (
	"RappaMaster/types"
	"RappaMaster/utils"
)

type TransactionType int

const (
	INIT_TASK_TRANSACTION = iota
	TASK_PROCESS_TRANSACTION
	EPOCH_RECORD_TRANSACTION
)

type Transaction interface {
	Call() string
	CallData() ([][32]byte, [][32]byte)
}

// InitTaskTransaction a transaction to create a task, get the txHash
type InitTaskTransaction struct {
	tasks []types.Task
}

func (t *InitTaskTransaction) Call() string {
	return "InitTask"
}
func (t *InitTaskTransaction) CallData() ([][32]byte, [][32]byte) {
	keys := make([][32]byte, len(t.tasks))
	values := make([][32]byte, 1)
	for i := 0; i < len(t.tasks); i++ {
		keys[i] = utils.StringToBytes32(t.tasks[i].Name())
		values[i] = utils.StringToBytes32(t.tasks[i].Sign()) // todo not a good implementation
	}
	return keys, values
}

func NewInitTaskTransaction(tasks []types.Task) *InitTaskTransaction {
	return &InitTaskTransaction{tasks}
}

//
//// TaskProcessTransaction
//type TaskProcessTransaction struct {
//	*paradigm.CommitSlotItem                           // 这里有很多的字段
//	Proof                    paradigm.Proof            // 这个是可信证明，暂时先不用放到链上，先在链上准备一下相关的字段和函数
//	Signatures               [][]byte                  // 这里后面我们要加入节点的公私钥，每次投票都会附上自己的签名，我打算在这里签名上链验证的，和上面一样，先留下字段
//	Model                    paradigm.SupportModelType // TODO 这里的目的是暂时拿到任务类型给monitor，先这样写
//}
//
//func (t *TaskProcessTransaction) Call() string {
//	return "Commit" // 简单先写一下，后面具体和合约对齐
//}
//func (t *TaskProcessTransaction) CallData() map[string]interface{} {
//	result := make(map[string]interface{})
//	result["Sign"] = t.Sign
//	result["Slot"] = t.Slot
//	result["Process"] = t.Process
//	result["ID"] = t.Nid
//	//result["Vote"] = t.Votes
//	result["Epoch"] = t.Epoch
//	result["Hash"] = t.hash
//	result["Commitment"] = t.Commitment
//	result["Proof"] = t.Proof
//	result["Signatures"] = t.Signatures
//	return result
//}
//func (t *TaskProcessTransaction) Blob() interface{} {
//	return t.CommitSlotItem // 将整个commitSlot返回
//}
//
//// EpochRecordTransaction
///* 交易类型三：epoch记录交易 */
//// 这里每个epoch会给出三个内容:
//// 1. 节点提交了哪些slot(commit);
//// 2. 哪些slot通过了投票(justified);
//// 3. 检测出了哪些不合法的slot
//// 后面可能还会加上收到了哪些slot的proof等，先按上面的写
//type EpochRecordTransaction struct {
//	// 这里所有的slot都是用一个独一无二的hash表示的(目前就是个string也就是sign_slot_nid)，因为我们只需要在链上记录真正落盘的slot就行
//	// 其它那些可能在未来检测出不合法的，没必要存，那些落盘的最后会由TaskProcessTransaction给出
//	*paradigm.EpochRecord       // 这里有很多字段
//	Id                    int32 // epoch id
//	CommitsHash           []paradigm.SlotHash
//	JustifiedHash         []paradigm.SlotHash
//	Invalids              map[paradigm.SlotHash]paradigm.InvalidCommitType //这里不同的异常slot有不同的理由
//}
//
//func (t *EpochRecordTransaction) Call() string {
//	return "EpochUpdate"
//}
//func (t *EpochRecordTransaction) CallData() map[string]interface{} {
//	result := make(map[string]interface{})
//	result["id"] = t.Id
//	result["Justified"] = t.JustifiedHash
//	result["Commits"] = t.CommitsHash
//	result["invalids"] = t.Invalids
//	return result
//}
//func (t *EpochRecordTransaction) Blob() interface{} {
//	return t.EpochRecord // 返回epochRecord
//}
//
//type PackedTransaction struct {
//	Tx          Transaction
//	Id          int
//	Receipt     *types.Receipt
//	BlockHash   string
//	UpchainTime time.Time
//}
//
//func (t *PackedTransaction) SetID(id int) {
//	t.Id = id
//}
//
//func (t *PackedTransaction) SetUpchainTime(time time.Time) {
//	t.UpchainTime = time
//}
//
//func (t *PackedTransaction) SetBlockInfo(block *types.Block) {
//	t.BlockHash = block.Hash
//	t.UpchainTime = paradigm.TimestampConvert(block.Timestamp)
//}
//
//func NewPackedTransaction(tx Transaction, receipt *types.Receipt, blockHash string) *PackedTransaction {
//	return &PackedTransaction{
//		Tx:        tx,
//		Id:        -1,
//		Receipt:   receipt,
//		BlockHash: blockHash,
//		// MerkleRoot: merkleRoot,
//	}
//}
