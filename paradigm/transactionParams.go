package paradigm

import (
	"BHLayer2Node/utils"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/FISCO-BCOS/go-sdk/v3/types"
)

type PackedParams interface {
	UpdateFromTransaction(tx Transaction) // 根据交易得到参数
	GetParams() []interface{}
	ParamsLen() int
	IsEmpty() bool
	ConvertParamsToKVPairs() ([32]byte, []byte) // setItem
	ParamsToKVPairs() ([][32]byte, [][]byte)    // batch setItem
	BuildDevTransactions(receipts []*types.Receipt, blockHash string) []*PackedTransaction
}

// TODO 这里每多一个交易类型就要加上对应的参数

type InitTaskTransactionParams struct {
	txs []*InitTaskTransaction
	// TODO @XQ
	signs  [][32]byte
	sizes  []*big.Int
	models [][32]byte
	// isReliables []bool
	// paramsList  []map[string]interface{}
	isReliables [][32]byte
	paramsList  [][32]byte
}

func (p *InitTaskTransactionParams) UpdateFromTransaction(tx Transaction) {
	switch tx.(type) {
	case *InitTaskTransaction:
		p.txs = append(p.txs, tx.(*InitTaskTransaction))
		calldata := tx.CallData()
		p.signs = append(p.signs, utils.StringToBytes32(calldata["Sign"].(string)))
		p.sizes = append(p.sizes, new(big.Int).SetUint64(uint64(calldata["Size"].(int32))))
		p.models = append(p.models, utils.StringToBytes32(ModelTypeToString(calldata["Model"].(SupportModelType))))
		// p.isReliables = append(p.isReliables, calldata["IsReliable"].(bool))
		// p.paramsList = append(p.paramsList, calldata["Params"].(map[string]interface{}))
		isRel := calldata["IsReliable"].(bool)
		reliableStr := "false"
		if isRel {
			reliableStr = "true"
		}
		p.isReliables = append(p.isReliables, utils.StringToBytes32(reliableStr))
		p.paramsList = append(p.paramsList, utils.SerializeParams(calldata["Params"].(map[string]interface{})))

	default:
		e := Error(RuntimeError, "TaskProcessTransactionParams should be updated from TaskProcessTransaction!!!")
		panic(e.Error())

	}
}

//	func (p *InitTaskTransactionParams) getSigns() [][32]byte {
//		return [][32]byte{}
//	}
func (p *InitTaskTransactionParams) GetParams() []interface{} {
	// return []interface{}{}
	return []interface{}{p.signs, p.sizes, p.models, p.isReliables, p.paramsList}
}
func (p *InitTaskTransactionParams) ParamsLen() int {
	return 5
}
func (p *InitTaskTransactionParams) IsEmpty() bool {
	return len(p.txs) == 0
}

// todo: 有的交易缺乏一些参数，导致参数列表长度对不齐 暂时全部打包成一笔交易
func (p *InitTaskTransactionParams) ConvertParamsToKVPairs() ([32]byte, []byte) {
	// 获取所有参数数据
	params := p.GetParams()
	// 序列化为 JSON 字符串
	value, err := json.Marshal(params)
	if err != nil {
		e := Error(RuntimeError, fmt.Sprintf("JSON marshal error: %v", err))
		panic(e.Error())
	}
	// 获取当前时间戳，并生成 key 字符串
	timestamp := time.Now().Unix()
	keyStr := fmt.Sprintf("INIT_TASK_BATCH_%d", timestamp)
	// 将 key 字符串转换为 [32]byte
	key := utils.StringToBytes32(keyStr)
	return key, value
}
func (p *InitTaskTransactionParams) ParamsToKVPairs() ([][32]byte, [][]byte) {
	n := len(p.signs)
	keys := make([][32]byte, n)
	values := make([][]byte, n)
	for i := 0; i < n; i++ {
		keys[i] = p.signs[i] // 使用任务sign作为key
		var composite []byte
		// 直接拼接各个[32]byte字段
		composite = append(composite, p.signs[i][:]...)
		composite = append(composite, utils.BigIntToBytes32(p.sizes[i])...)
		composite = append(composite, p.models[i][:]...)
		composite = append(composite, p.isReliables[i][:]...)
		composite = append(composite, p.paramsList[i][:]...)
		values[i] = composite
	}
	return keys, values
}
func (p *InitTaskTransactionParams) BuildDevTransactions(receipts []*types.Receipt, blockHash string) []*PackedTransaction {
	receipt := receipts[0]
	result := make([]*PackedTransaction, 0)
	for _, tx := range p.txs {
		result = append(result, NewPackedTransaction(tx, receipt, blockHash))
	}
	return result
}
func NewInitTaskTransactionParams() *InitTaskTransactionParams {
	return &InitTaskTransactionParams{
		txs:         make([]*InitTaskTransaction, 0),
		signs:       make([][32]byte, 0),
		sizes:       make([]*big.Int, 0),
		models:      make([][32]byte, 0),
		isReliables: make([][32]byte, 0),
		paramsList:  make([][32]byte, 0),
	}
}

type TaskProcessTransactionParams struct {
	// TODO @XQ
	signs, hashs                                           [][32]byte
	slotsBigInt, processesBigInt, nidsBigInt, epochsBigInt []*big.Int
	txs                                                    []*TaskProcessTransaction
}

func (p *TaskProcessTransactionParams) UpdateFromTransaction(tx Transaction) {
	switch tx.(type) {
	case *TaskProcessTransaction:
		calldata := tx.CallData()
		p.txs = append(p.txs, tx.(*TaskProcessTransaction))
		p.signs = append(p.signs, utils.StringToBytes32(calldata["Sign"].(string)))
		p.hashs = append(p.hashs, utils.StringToBytes32(calldata["Hash"].(string)))
		p.slotsBigInt = append(p.slotsBigInt, new(big.Int).SetUint64(uint64(calldata["Slot"].(int32))))
		p.processesBigInt = append(p.processesBigInt, new(big.Int).SetUint64(uint64(calldata["Process"].(int32))))
		// p.nidsBigInt = append(p.nidsBigInt, new(big.Int).SetUint64(uint64(calldata["ID"].(int32))))
		var idValue int32 = 0
		if calldata["ID"] != nil {
			idValue = calldata["ID"].(int32)
		}
		p.nidsBigInt = append(p.nidsBigInt, new(big.Int).SetUint64(uint64(idValue)))
		p.epochsBigInt = append(p.epochsBigInt, new(big.Int).SetUint64(uint64(calldata["Epoch"].(int32))))
	default:
		e := Error(RuntimeError, "TaskProcessTransactionParams should be updated from TaskProcessTransaction!!!")
		panic(e.Error())

	}
}
func (p *TaskProcessTransactionParams) GetParams() []interface{} {
	return []interface{}{p.signs, p.hashs, p.slotsBigInt, p.processesBigInt, p.nidsBigInt, p.epochsBigInt}
}
func (p *TaskProcessTransactionParams) ParamsLen() int {
	return 6
}
func (p *TaskProcessTransactionParams) IsEmpty() bool {
	return len(p.txs) == 0
}

func (p *TaskProcessTransactionParams) ConvertParamsToKVPairs() ([32]byte, []byte) {
	// 获取所有参数数据
	params := p.GetParams()
	// 序列化为 JSON 字符串
	value, err := json.Marshal(params)
	if err != nil {
		e := Error(RuntimeError, fmt.Sprintf("JSON marshal error: %v", err))
		panic(e.Error())
	}
	// 获取当前时间戳，并生成 key 字符串
	timestamp := time.Now().Unix()
	keyStr := fmt.Sprintf("PROCESS_TASK_BATCH_%d", timestamp)
	// 将 key 字符串转换为 [32]byte
	key := utils.StringToBytes32(keyStr)
	return key, value
}
func (p *TaskProcessTransactionParams) ParamsToKVPairs() ([][32]byte, [][]byte) {
	n := len(p.signs)
	keys := make([][32]byte, n)
	values := make([][]byte, n)
	for i := 0; i < n; i++ {
		keys[i] = p.hashs[i] // 使用hash作为key
		var composite []byte
		// 直接拼接各个[32]byte字段
		composite = append(composite, p.signs[i][:]...)
		composite = append(composite, p.hashs[i][:]...)
		composite = append(composite, utils.BigIntToBytes32(p.slotsBigInt[i])...)
		composite = append(composite, utils.BigIntToBytes32(p.processesBigInt[i])...)
		composite = append(composite, utils.BigIntToBytes32(p.nidsBigInt[i])...)
		composite = append(composite, utils.BigIntToBytes32(p.epochsBigInt[i])...)
		values[i] = composite
	}
	return keys, values
}

func (p *TaskProcessTransactionParams) BuildDevTransactions(receipts []*types.Receipt, blockHash string) []*PackedTransaction {
	// todo 这里暂时先为多个receipt做好准备
	// 要判断receipt和transaction长度 todo

	// 现在只有一个receipt
	receipt := receipts[0]
	result := make([]*PackedTransaction, 0)
	for _, tx := range p.txs {
		result = append(result, NewPackedTransaction(tx, receipt, blockHash))
	}
	return result
}
func NewTaskProcessTransactionParams() *TaskProcessTransactionParams {
	return &TaskProcessTransactionParams{
		signs:           make([][32]byte, 0),
		hashs:           make([][32]byte, 0),
		slotsBigInt:     make([]*big.Int, 0),
		processesBigInt: make([]*big.Int, 0),
		nidsBigInt:      make([]*big.Int, 0),
		epochsBigInt:    make([]*big.Int, 0),
		txs:             make([]*TaskProcessTransaction, 0),
	}
}

type EpochRecordTransactionParams struct {
	// TODO @XQ
	txs        []*EpochRecordTransaction
	ids        []*big.Int
	justifieds [][][32]byte
	commits    [][][32]byte
	invalids   []map[string]interface{}
}

func (p *EpochRecordTransactionParams) UpdateFromTransaction(tx Transaction) {
	switch tx.(type) {
	case *EpochRecordTransaction:
		p.txs = append(p.txs, tx.(*EpochRecordTransaction))
		calldata := tx.CallData()
		p.ids = append(p.ids, new(big.Int).SetUint64(uint64(calldata["id"].(int32))))
		// 处理Justified (带类型安全检查和默认值)
		var justifiedsArr [][32]byte
		if justified, ok := calldata["Justified"].([]SlotHash); ok {
			for _, s := range justified {
				justifiedsArr = append(justifiedsArr, utils.StringToBytes32(s))
			}
		}
		p.justifieds = append(p.justifieds, justifiedsArr)
		// 处理Commits (同上)
		var commitArr [][32]byte
		if commits, ok := calldata["Commits"].([]SlotHash); ok {
			for _, s := range commits {
				commitArr = append(commitArr, utils.StringToBytes32(s))
			}
		}
		p.commits = append(p.commits, commitArr)
		// 异常slot处理
		var invalidMap = make(map[string]interface{})
		for k, v := range calldata["invalids"].(map[SlotHash]InvalidCommitType) {
			invalidMap[k] = v
		}
		p.invalids = append(p.invalids, invalidMap)

	default:
		e := Error(RuntimeError, "TaskProcessTransactionParams should be updated from TaskProcessTransaction!!!")
		panic(e.Error())

	}
}
func (p *EpochRecordTransactionParams) GetParams() []interface{} {
	return []interface{}{p.ids, p.justifieds, p.commits, p.invalids}
}
func (p *EpochRecordTransactionParams) ParamsLen() int {
	return 4
}
func (p *EpochRecordTransactionParams) IsEmpty() bool {
	return len(p.txs) == 0
}
func (p *EpochRecordTransactionParams) ConvertParamsToKVPairs() ([32]byte, []byte) {
	params := p.GetParams()
	// 序列化为 JSON 字符串
	value, err := json.Marshal(params)
	if err != nil {
		panic(fmt.Sprintf("JSON marshal error: %v", err))
	}
	// 获取当前时间戳，并生成 key 字符串
	timestamp := time.Now().Unix()
	keyStr := fmt.Sprintf("EPOCH_RECORD_BATCH_%d", timestamp)
	// 将 key 字符串转换为 [32]byte
	key := utils.StringToBytes32(keyStr)
	return key, value
}
func (p *EpochRecordTransactionParams) ParamsToKVPairs() ([][32]byte, [][]byte) {
	n := len(p.ids)
	// LogWriter.Log("DEBUG", fmt.Sprintf("EPOCH PARAMS: %s", p.GetParams()...))
	// 会出现某个epoch没有任何新提交的情况，但是epoch仍需要上链
	keys := make([][32]byte, n)
	values := make([][]byte, n)
	for i := 0; i < n; i++ {
		keys[i] = utils.StringToBytes32(fmt.Sprintf("EPOCH_%d", p.ids[i])) // 使用epochid作为key
		var composite []byte
		// 直接拼接各个[32]byte字段
		composite = append(composite, utils.FlattenByte32Slice(p.justifieds[i])...)
		composite = append(composite, utils.FlattenByte32Slice(p.commits[i])...)
		invalids := utils.SerializeParams(p.invalids[i])
		composite = append(composite, invalids[:]...)
		values[i] = composite
	}
	return keys, values
}
func (p *EpochRecordTransactionParams) BuildDevTransactions(receipts []*types.Receipt, blockHash string) []*PackedTransaction {
	receipt := receipts[0]
	result := make([]*PackedTransaction, 0)
	for _, tx := range p.txs {
		result = append(result, NewPackedTransaction(tx, receipt, blockHash))
	}
	return result
}
func NewEpochRecordTransactionParams() *EpochRecordTransactionParams {
	return &EpochRecordTransactionParams{
		txs:        make([]*EpochRecordTransaction, 0),
		ids:        make([]*big.Int, 0),
		justifieds: make([][][32]byte, 0),
		commits:    make([][][32]byte, 0),
		invalids:   make([]map[string]interface{}, 0),
	}
}

func NewParamsMap() map[TransactionType]PackedParams {
	params := make(map[TransactionType]PackedParams)
	params[INIT_TASK_TRANSACTION] = NewInitTaskTransactionParams()
	params[TASK_PROCESS_TRANSACTION] = NewTaskProcessTransactionParams()
	params[EPOCH_RECORD_TRANSACTION] = NewEpochRecordTransactionParams()
	return params
}
