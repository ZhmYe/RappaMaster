package paradigm

import (
	"BHLayer2Node/utils"
	"github.com/FISCO-BCOS/go-sdk/v3/types"
	"math/big"
)

type PackedParams interface {
	UpdateFromTransaction(tx Transaction) // 根据交易得到参数
	GetParams() []interface{}
	ParamsLen() int
	BuildDevTransactions(receipts []*types.Receipt) []*PackedTransaction
}

// TODO 这里每多一个交易类型就要加上对应的参数

type InitTaskTransactionParams struct {
	txs []*InitTaskTransaction
	// TODO @XQ
}

func (p *InitTaskTransactionParams) UpdateFromTransaction(tx Transaction) {
	switch tx.(type) {
	case *InitTaskTransaction:
		p.txs = append(p.txs, tx.(*InitTaskTransaction))
	default:
		panic("TaskProcessTransactionParams should be updated from TaskProcessTransaction!!!")

	}
}

//	func (p *InitTaskTransactionParams) getSigns() [][32]byte {
//		return [][32]byte{}
//	}
func (p *InitTaskTransactionParams) GetParams() []interface{} {
	return []interface{}{}
}
func (p *InitTaskTransactionParams) ParamsLen() int {
	return -1
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
		txs: make([]*InitTaskTransaction, 0),
	}
}

type TaskProcessTransactionParams struct {
	// TODO @XQ
	signs                                                  [][32]byte
	slotsBigInt, processesBigInt, nidsBigInt, epochsBigInt []*big.Int
	txs                                                    []*TaskProcessTransaction
}

func (p *TaskProcessTransactionParams) UpdateFromTransaction(tx Transaction) {
	switch tx.(type) {
	case *TaskProcessTransaction:
		calldata := tx.CallData()
		p.txs = append(p.txs, tx.(*TaskProcessTransaction))
		p.signs = append(p.signs, utils.StringToBytes32(calldata["Sign"].(string)))
		p.slotsBigInt = append(p.slotsBigInt, new(big.Int).SetUint64(uint64(calldata["Slot"].(int32))))
		p.processesBigInt = append(p.processesBigInt, new(big.Int).SetUint64(uint64(calldata["Process"].(int32))))
		p.nidsBigInt = append(p.nidsBigInt, new(big.Int).SetUint64(uint64(calldata["ID"].(int32))))
		p.epochsBigInt = append(p.epochsBigInt, new(big.Int).SetUint64(uint64(calldata["Epoch"].(int32))))
	default:
		panic("TaskProcessTransactionParams should be updated from TaskProcessTransaction!!!")

	}
}
func (p *TaskProcessTransactionParams) GetParams() []interface{} {
	return []interface{}{p.signs, p.slotsBigInt, p.processesBigInt, p.nidsBigInt, p.epochsBigInt}
}
func (p *TaskProcessTransactionParams) ParamsLen() int {
	return 5
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
		slotsBigInt:     make([]*big.Int, 0),
		processesBigInt: make([]*big.Int, 0),
		nidsBigInt:      make([]*big.Int, 0),
		epochsBigInt:    make([]*big.Int, 0),
		txs:             make([]*TaskProcessTransaction, 0),
	}
}

type EpochRecordTransactionParams struct {
	txs []*EpochRecordTransaction
	// TODO @XQ
}

func (p *EpochRecordTransactionParams) UpdateFromTransaction(tx Transaction) {
	switch tx.(type) {
	case *EpochRecordTransaction:
		p.txs = append(p.txs, tx.(*EpochRecordTransaction))
	default:
		panic("TaskProcessTransactionParams should be updated from TaskProcessTransaction!!!")

	}
}
func (p *EpochRecordTransactionParams) GetParams() []interface{} {
	return []interface{}{}
}
func (p *EpochRecordTransactionParams) ParamsLen() int {
	return -1
}
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
		txs: make([]*EpochRecordTransaction, 0),
	}
}

func NewParamsMap() map[TransactionType]PackedParams {
	params := make(map[TransactionType]PackedParams)
	params[INIT_TASK_TRANSACTION] = NewInitTaskTransactionParams()
	params[TASK_PROCESS_TRANSACTION] = NewTaskProcessTransactionParams()
	params[EPOCH_RECORD_TRANSACTION] = NewEpochRecordTransactionParams()
	return params
}
