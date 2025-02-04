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
	ConvertParamsToKVPairs() ([32]byte, []byte) // 将参数转化为32字节的kv对数组
	BuildDevTransactions(receipts []*types.Receipt) []*PackedTransaction
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
		panic("TaskProcessTransactionParams should be updated from TaskProcessTransaction!!!")

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
		panic(fmt.Sprintf("JSON marshal error: %v", err))
	}
	// 获取当前时间戳，并生成 key 字符串
	timestamp := time.Now().Unix()
	keyStr := fmt.Sprintf("INIT_TASK_BATCH_%d", timestamp)
	// 将 key 字符串转换为 [32]byte
	key := utils.StringToBytes32(keyStr)
	return key, value
}
func (p *InitTaskTransactionParams) BuildDevTransactions(receipts []*types.Receipt) []*PackedTransaction {
	receipt := receipts[0]
	result := make([]*PackedTransaction, 0)
	for _, tx := range p.txs {
		result = append(result, NewPackedTransaction(tx, receipt))
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
		p.hashs = append(p.signs, utils.StringToBytes32(calldata["Hash"].(string)))
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
		panic("TaskProcessTransactionParams should be updated from TaskProcessTransaction!!!")

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
		panic(fmt.Sprintf("JSON marshal error: %v", err))
	}
	// 获取当前时间戳，并生成 key 字符串
	timestamp := time.Now().Unix()
	keyStr := fmt.Sprintf("PROCESS_TASK_BATCH_%d", timestamp)
	// 将 key 字符串转换为 [32]byte
	key := utils.StringToBytes32(keyStr)
	return key, value
}

func (p *TaskProcessTransactionParams) BuildDevTransactions(receipts []*types.Receipt) []*PackedTransaction {
	// todo 这里暂时先为多个receipt做好准备
	// 要判断receipt和transaction长度 todo

	// 现在只有一个receipt
	receipt := receipts[0]
	result := make([]*PackedTransaction, 0)
	for _, tx := range p.txs {
		result = append(result, NewPackedTransaction(tx, receipt))
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

		var justifiedsArr [][32]byte
		for _, s := range calldata["Justified"].([]SlotHash) {
			justifiedsArr = append(justifiedsArr, utils.StringToBytes32(s))
		}
		p.commits = append(p.justifieds, justifiedsArr)

		var commitArr [][32]byte
		for _, s := range calldata["Commits"].([]SlotHash) {
			commitArr = append(commitArr, utils.StringToBytes32(s))
		}
		p.commits = append(p.commits, commitArr)
		// 异常slot处理
		// var invalidMap = make(map[[32]byte]interface{})
		// for k, v := range calldata["invalids"].(map[SlotHash]InvalidCommitType) {
		// 	invalidMap[utils.StringToBytes32(k)] = v
		// }
		// p.invalids = append(p.invalids, invalidMap)
		var invalidMap = make(map[string]interface{})
		for k, v := range calldata["invalids"].(map[SlotHash]InvalidCommitType) {
			invalidMap[k] = v
		}
		p.invalids = append(p.invalids, invalidMap)

	default:
		panic("TaskProcessTransactionParams should be updated from TaskProcessTransaction!!!")

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

// 使用 epoch id 作为 key，其他属性整合为 value，格式： "justified1,justified2,...|commit1,commit2,...|invalidKey1:invalidValue1,invalidKey2:invalidValue2,..."
// func (p *EpochRecordTransactionParams) ConvertParamsToKVPairs() ([][32]byte, [][32]byte) {
// 	n := len(p.ids)
// 	keys := make([][32]byte, n)
// 	values := make([][32]byte, n)
// 	for i := 0; i < n; i++ {
// 		// key：将 id 转换为字符串后转换为 [32]byte
// 		keys[i] = utils.StringToBytes32(p.ids[i].String())

// 		// 处理 justified 数组：如果 p.justifieds 未包含当前索引，则使用空数组
// 		var justArr [][32]byte
// 		if i < len(p.justifieds) {
// 			justArr = p.justifieds[i]
// 		} else {
// 			justArr = make([][32]byte, 0)
// 		}
// 		justStr := ""
// 		for j, b := range justArr {
// 			if j > 0 {
// 				justStr += ","
// 			}
// 			justStr += string(b[:])
// 		}

// 		// 处理 commits 数组：同上，确保不会越界
// 		var commitArr [][32]byte
// 		if i < len(p.commits) {
// 			commitArr = p.commits[i]
// 		} else {
// 			commitArr = make([][32]byte, 0)
// 		}
// 		commitStr := ""
// 		for j, b := range commitArr {
// 			if j > 0 {
// 				commitStr += ","
// 			}
// 			commitStr += string(b[:])
// 		}

// 		// 处理 invalids map：如果不存在，则使用空 map
// 		var invMap map[[32]byte]interface{}
// 		if i < len(p.invalids) {
// 			invMap = p.invalids[i]
// 		} else {
// 			invMap = make(map[[32]byte]interface{})
// 		}
// 		invalidStr := ""
// 		for k, v := range invMap {
// 			invalidStr += fmt.Sprintf("%s:%v,", string(k[:]), v)
// 		}

//			// 将各部分拼接，采用 "|" 分隔
//			valStr := fmt.Sprintf("%s|%s|%s", justStr, commitStr, invalidStr)
//			values[i] = utils.StringToBytes32(valStr)
//		}
//		return keys, values
//	}
func (p *EpochRecordTransactionParams) BuildDevTransactions(receipts []*types.Receipt) []*PackedTransaction {
	receipt := receipts[0]
	result := make([]*PackedTransaction, 0)
	for _, tx := range p.txs {
		result = append(result, NewPackedTransaction(tx, receipt))
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
