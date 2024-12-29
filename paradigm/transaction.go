package paradigm

import (
	"github.com/FISCO-BCOS/go-sdk/v3/types"
)

/*** Transaction相关内容 ***/

type TransactionType int

const (
	INIT_TASK_TRANSACTION = iota
	TASK_PROCESS_TRANSACTION
	EPOCH_RECORD_TRANSACTION
)

type Transaction interface {
	Call() string                     // 调用合约哪个函数
	CallData() map[string]interface{} // 调用合约函数时候的参数
	Blob() interface{}                // 提供额外的附属内容
}

// TODO @YZM 这里的CallData可以改成遍历结构体中的所有成员变量，然后结合reflect变成map里的内容,这样新定义的transaction无需写CallData()

// TODO @XQ 这里定义了几个我初步想的具体交易实例，修改相应的合约和sdk client
// 下面的所有的合约函数名、参数等可以根据solidity实况进行调整，如果有大调整比如类型无法支持啥的可以提issues
// 写了一些对前端的想象，可以理解一下，有问题可以提

// InitTaskTransaction 表示RappaMaster新收到了一个合成任务，将任务的一些相关参数传上去，同时在合约里其他一些和任务相关的地方准备新建这个相关的字段
/* 交易类型一: 新建任务交易 */
type InitTaskTransaction struct {
	//Sign       string                 // 合成任务标识
	//Size       int32                  // 合成数据量
	//Model      string                 // 模型,这里可以用某个特殊的函数映射一下也行
	//IsReliable bool                   // 是否需要可信证明，这个一定要上链
	//Params     map[string]interface{} // 合成任务相关参数
	*Task // 这里有很多字段
	// 大概是这样
}

func (t *InitTaskTransaction) Call() string {
	return "InitTask"
}
func (t *InitTaskTransaction) CallData() map[string]interface{} {
	result := make(map[string]interface{})
	result["Sign"] = t.Sign
	result["Size"] = t.Size
	result["Model"] = t.Model
	result["IsReliable"] = t.IsReliable()
	result["Params"] = t.Params // 这里如果solidity不太好操作，就把它用hash转化成一个bytes传上去，后续是否要考虑验证啥的再说
	return result
}
func (t *InitTaskTransaction) Blob() interface{} {
	return t.Task
}
func NewInitTaskTransaction(task *Task) *InitTaskTransaction {
	return &InitTaskTransaction{task}
}

// TaskProcessTransaction 表示在RappaMaster中一个commitSlot通过了验证(1. 存储, 2. 可信证明(可选))
/* 交易类型二: 任务进度更新交易 */
// 这里上链以后要维护的是若干个task结构体，这些结构体的状态会不断更新，其中会有一个slots数组，slots根据下面提交的slot和nid更新字段，可能会比较大，后续用于展示
// 我现在想的我们未来的前端里，溯源界面，会有一个EpochChain，也就是每个epoch完成了哪些任务的slot；会有一个Task的搜索，给出TaskSlotChain，也就是每个task的每个slot在哪些epoch里被完成了多少
type TaskProcessTransaction struct {
	*CommitSlotItem          // 这里有很多的字段
	Proof           Proof    // 这个是可信证明，暂时先不用放到链上，先在链上准备一下相关的字段和函数
	Signatures      [][]byte // 这里后面我们要加入节点的公私钥，每次投票都会附上自己的签名，我打算在这里签名上链验证的，和上面一样，先留下字段
}

func (t *TaskProcessTransaction) Call() string {
	return "Commit" // 简单先写一下，后面具体和合约对齐
}
func (t *TaskProcessTransaction) CallData() map[string]interface{} {
	result := make(map[string]interface{})
	result["Sign"] = t.Sign
	result["Slot"] = t.Slot
	result["Process"] = t.Process
	result["ID"] = t.Nid
	//result["Vote"] = t.Votes
	result["Epoch"] = t.Epoch
	result["Hash"] = t.hash
	result["Commitment"] = t.Commitment
	// 下面的calldata里proof和signatures可以注释掉
	result["Proof"] = t.Proof
	result["Signatures"] = t.Signatures
	return result
}
func (t *TaskProcessTransaction) Blob() interface{} {
	return t.CommitSlotItem // 将整个commitSlot返回
}

// EpochRecordTransaction
/* 交易类型三：epoch记录交易 */
// 这里每个epoch会给出三个内容:
// 1. 节点提交了哪些slot(commit);
// 2. 哪些slot通过了投票(justified);
// 3. 检测出了哪些不合法的slot
// 后面可能还会加上收到了哪些slot的proof等，先按上面的写
type EpochRecordTransaction struct {
	// 这里所有的slot都是用一个独一无二的hash表示的(目前就是个string也就是sign_slot_nid)，因为我们只需要在链上记录真正落盘的slot就行
	// 其它那些可能在未来检测出不合法的，没必要存，那些落盘的最后会由TaskProcessTransaction给出
	*EpochRecord        // 这里有很多字段
	Id            int32 // epoch id
	CommitsHash   []SlotHash
	JustifiedHash []SlotHash
	Invalids      map[SlotHash]InvalidCommitType //这里不同的异常slot有不同的理由
}

func (t *EpochRecordTransaction) Call() string {
	return "EpochUpdate"
}
func (t *EpochRecordTransaction) CallData() map[string]interface{} {
	result := make(map[string]interface{})
	result["id"] = t.Id
	result["Justified"] = t.JustifiedHash
	result["Commits"] = t.CommitsHash
	result["invalids"] = t.Invalids
	return result
}
func (t *EpochRecordTransaction) Blob() interface{} {
	return t.EpochRecord // 返回epochRecord
}

type PackedTransaction struct {
	Tx      Transaction
	Id      int
	Receipt *types.Receipt
}

func (t *PackedTransaction) SetID(id int) {
	t.Id = id
}
func NewPackedTransaction(tx Transaction, receipt *types.Receipt) *PackedTransaction {
	return &PackedTransaction{
		Tx:      tx,
		Id:      -1,
		Receipt: receipt,
	}
}
