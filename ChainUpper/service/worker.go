package service

import (
	"context"
	"fmt"
	"math/big"
	"time"

	Store "BHLayer2Node/ChainUpper/contract/storeData"
	"BHLayer2Node/paradigm"

	"github.com/FISCO-BCOS/go-sdk/v3/client"
	"github.com/FISCO-BCOS/go-sdk/v3/types"
	"github.com/ethereum/go-ethereum/common"
)

// 单笔上链，不做批量 链上存储具体的交易数据
type UpChainWorker struct {
	id                   int
	queue                chan paradigm.Transaction
	devPackedTransaction chan []*paradigm.PackedTransaction
	instance             *Store.StoreData
	client               *client.Client
}

func NewUpchainWorker(
	id int,
	queue chan paradigm.Transaction,
	dev chan []*paradigm.PackedTransaction,
	instance *Store.StoreData,
	client *client.Client,
) *UpChainWorker {
	return &UpChainWorker{
		id:                   id,
		queue:                queue,
		devPackedTransaction: dev,
		instance:             instance,
		client:               client,
	}
}

func (w *UpChainWorker) Process() {
	for tx := range w.queue {
		if tx == nil {
			return
		}
		params, err := tx.StoreParams()
		paradigm.Log("CHAINUP", fmt.Sprintf("Transaction %s:\n %+v\n upchain params: %+v", tx.Call(), tx.CallData(), params))
		if err != nil {
			paradigm.Error(paradigm.UpchainError, "转换上链参数失败: "+err.Error())
			continue
		}
		session := &Store.StoreDataSession{Contract: w.instance, TransactOpts: *w.client.GetTransactOpts()}
		// 收到单笔交易，直接上链
		var receipt *types.Receipt
		switch tx.(type) {
		case *paradigm.InitTaskTransaction:
			_, receipt, err = session.StoreInitTask(
				params[0].([32]byte),
				params[1].(string),
				params[2].(uint32),
				params[3].([32]byte),
				params[4].(bool),
				params[5].([]byte),
			)
			Tx := tx.Blob().(*paradigm.Task)
			Txjson, err := w.GetInitTaskJSON(Tx.Sign)
			if err != nil {
				paradigm.Log("ERROR", fmt.Sprintf("query InitTaskTransaction from blockchain By task %s failed %s", Tx.Sign, err))
			} else {
				paradigm.Log("CHAINUP", fmt.Sprintf("query InitTaskTransaction from blockchain By task %s: %s", Tx.Sign, Txjson))
			}
		case *paradigm.TaskProcessTransaction:
			_, receipt, err = session.StoreTaskProcess(
				params[0].([32]byte),
				params[1].([32]byte),
				params[2].(uint32),
				params[3].(uint32),
				params[4].(uint32),
				params[5].(uint32),
				params[6].(common.Hash),
				params[7].([]byte),
				params[8].([][]byte),
			)
			Tx := tx.Blob().(*paradigm.CommitSlotItem)
			Txjson, err := w.GetTaskProcessJSON(Tx.SlotHash())
			if err != nil {
				paradigm.Log("ERROR", fmt.Sprintf("query TaskProcessTransaction from blockchain By Slot %s failed %s", Tx.SlotHash(), err))
			} else {
				paradigm.Log("CHAINUP", fmt.Sprintf("query TaskProcessTransaction from blockchain By Slot %s: %s", Tx.SlotHash(), Txjson))
			}
		case *paradigm.EpochRecordTransaction:
			_, receipt, err = session.StoreEpochRecord(
				params[0].(*big.Int),
				params[1].([][32]byte),
				params[2].([][32]byte),
				params[3].([][32]byte),
				params[4].([]uint8),
			)
			Tx := tx.Blob().(*paradigm.EpochRecord)
			Txjson, err := w.GetEpochRecordJSON(uint64(Tx.Id))
			if err != nil {
				paradigm.Log("ERROR", fmt.Sprintf("query EpochRecordTransaction from blockchain By ID %d failed %s", Tx.Id, err))
			} else {
				paradigm.Log("CHAINUP", fmt.Sprintf("query EpochRecordTransaction from blockchain By ID %d: %s", Tx.Id, Txjson))
			}
		default:
			panic("Unknown transaction type")
		}
		if err != nil {
			paradigm.Error(paradigm.UpchainError, "上链失败: "+err.Error())
			continue
		}
		w.waitAndDevPush(receipt, tx)
	}
}

// 交易上链后，将回执和交易打包发送给Oracle
func (w *UpChainWorker) waitAndDevPush(receipt *types.Receipt, tx paradigm.Transaction) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rcpt, err := w.client.GetTransactionReceipt(ctx, common.HexToHash(receipt.TransactionHash), true)
	if err != nil {
		paradigm.Error(paradigm.RuntimeError, "GetReceipt error: "+err.Error())
		return
	}
	blockHash, _ := w.client.GetBlockHashByNumber(ctx, int64(rcpt.BlockNumber))
	devTxs := []*paradigm.PackedTransaction{
		paradigm.NewPackedTransaction(tx, rcpt, blockHash.Hex()),
	}
	w.devPackedTransaction <- devTxs
}

// =============================2025年5月29日之前：批量上链kv对 packedParams将交易打包，但实际上现在的batchsize只是1，不需要批量上链功能

// TODO @XQ 这里看着应该是和合约相关，要把合约更新，我这边改成了结构体,下面一些todo看一下如果改不了就按照现在的一些格式写
// 下面的参数我从现在的写法来看，好像是go自动生成的代码里需要把合约的参数全部列出来传到函数里，那这样似乎不同交易的batch比较麻烦？
// 下面目前是把不同交易的参数分开了,不同的交易分开batch?

// // UpChainWorker modify by zhmye
// type UpChainWorker struct {
// 	id                   int
// 	queue                chan paradigm.Transaction
// 	devPackedTransaction chan []*paradigm.PackedTransaction // add by zhmye 这里是用来给dev的，所有已经上链的交易都要给
// 	instance             *Store.Store
// 	client               *client.Client
// 	batchSize            int
// 	//signs                                                  [][32]byte
// 	//slotsBigInt, processesBigInt, nidsBigInt, epochsBigInt []*big.Int
// 	params map[paradigm.TransactionType]paradigm.PackedParams // 这里记录各种类型的交易参数 add by zhmye
// 	count  int                                                // 交易计数
// }

// func (w *UpChainWorker) Process() {
// 	for {
// 		select {
// 		case tx := <-w.queue: // 尝试从通道中接收数据
// 			if tx != nil { // 判断是否接收到有效值
// 				// log.Printf("Worker %d Received result: %v", id, result)
// 				//paradigm.Log("CHAINUP", fmt.Sprintf("Worker %d Received Transaction: %v", w.id, tx))

// 				switch tx.(type) {
// 				case *paradigm.InitTaskTransaction:
// 					w.params[paradigm.INIT_TASK_TRANSACTION].UpdateFromTransaction(tx)
// 				case *paradigm.TaskProcessTransaction:
// 					w.params[paradigm.TASK_PROCESS_TRANSACTION].UpdateFromTransaction(tx)
// 				case *paradigm.EpochRecordTransaction:
// 					w.params[paradigm.EPOCH_RECORD_TRANSACTION].UpdateFromTransaction(tx)
// 				default:
// 					panic("Invalid Transaction Type!!!")
// 				}
// 				w.count++
// 				if w.count >= w.batchSize {
// 					// 每当收集到batchSize个transaction的信息时，调用批量上链函数
// 					w.consumer()
// 					w.count = 0

// 				}
// 			} else {
// 				paradigm.Log("ERROR", fmt.Sprintf("Upchain channel closed, received nil value"))
// 				return
// 			}
// 		}
// 	}
// }

// // consumer 将当前收集的交易参数转换为 KV 对后批量调用上链函数
// func (w *UpChainWorker) consumer() {
// 	client := w.client
// 	instance := w.instance
// 	// 对每种交易类型，都调用 ConvertParamsToKVPairs 得到 KV 对，然后调用合约函数 setItems
// 	for tType, packedParam := range w.params {
// 		if packedParam.IsEmpty() {
// 			// txs := packedParam.GetParams()
// 			// LogWriter.Log("DEBUG", fmt.Sprintf("TX_%d waiting for upchain: %s", tType, txs))
// 			continue
// 		}
// 		// key, value := packedParam.ConvertParamsToKVPairs()
// 		keys, values := packedParam.ParamsToKVPairs()
// 		// LogWriter.Log("DEBUG", fmt.Sprintf("TX_%d convert to:[key]%s [value]%s", tType, keys, values))
// 		storeSession := &Store.StoreSession{Contract: instance, CallOpts: *client.GetCallOpts(), TransactOpts: *client.GetTransactOpts()}
// 		// _, receipt, err := storeSession.SetItem(key, value)
// 		_, receipt, err := storeSession.SetItems(keys, values)
// 		if err != nil {
// 			paradigm.Error(paradigm.RuntimeError, fmt.Sprintf("Worker %d Failed to call SetItems for type %v: %v", w.id, tType, err))
// 		}
// 		// 获得有merkleProof的receipt
// 		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 		defer cancel()
// 		_receipt, err := w.client.GetTransactionReceipt(ctx, common.HexToHash(receipt.TransactionHash), true)
// 		if err != nil {
// 			paradigm.Error(paradigm.RuntimeError, fmt.Sprintf("Failed to getReceipt with merkleProof for type %v: %v", tType, err))
// 		} else {
// 			// LogWriter.Log("DEBUG", fmt.Sprintf("Receipt with merkleProof: %s", _receipt)) //debug
// 			// block, _ := w.client.GetBlockByNumber(ctx, int64(_receipt.BlockNumber), false, false)
// 			blockHash, _ := w.client.GetBlockHashByNumber(ctx, int64(_receipt.BlockNumber))
// 			ptxs := packedParam.BuildDevTransactions([]*types.Receipt{_receipt}, blockHash.Hex())
// 			w.devPackedTransaction <- ptxs // 传递到dev
// 		}

// 		// test: 验证能否从receipt中解析上链数据
// 		fmt.Println("Origin upchain Data:")
// 		fmt.Println("key: ", "0x"+hex.EncodeToString(keys[0][:]))
// 		fmt.Println("value: ", "0x"+hex.EncodeToString(values[0][:]))
// 		w.ParseReceipt(_receipt)

// 	}
// 	w.params = paradigm.NewParamsMap()
// }

// func NewUpchainWorker(id int, batchSize int, queue chan paradigm.Transaction, dev chan []*paradigm.PackedTransaction, instance *Store.Store, client *client.Client) *UpChainWorker {
// 	return &UpChainWorker{
// 		id:                   id,
// 		queue:                queue,
// 		devPackedTransaction: dev,
// 		instance:             instance,
// 		client:               client,
// 		batchSize:            batchSize,
// 		params:               paradigm.NewParamsMap(),
// 		count:                0,
// 	}
// }

// =====================================

// func (w *UpChainWorker) consumer() {
// 	client := w.client
// 	instance := w.instance
// 	// TODO @XQ 这里我不清楚是异步上链组件之前是在合约里把一大批交易压成一笔交易来上链？那一笔交易执行的不会很慢吗...
// 	// 我之前以为是区块链的交易处理能力有上限tps，因此每次打包一定数量的交易并行上链等待处理
// 	// 下面先按照原来的写
// 	// todo 这里目前只有taskProcessTransaction的，因为其他几个还没写
// 	for tType, packedParam := range w.params {
// 		// 这里三种不同的交易可以并行，用waitgroup + go routine todo,把下面的函数写成func(),然后下面go func(), wg.wait()
// 		if tType == paradigm.INIT_TASK_TRANSACTION {
// 			continue // todo
// 		}
// 		if tType == paradigm.EPOCH_RECORD_TRANSACTION {
// 			continue // todo
// 		}
// 		if tType == paradigm.TASK_PROCESS_TRANSACTION {
// 			storeSession := &Store.StoreSession{Contract: instance, CallOpts: *client.GetCallOpts(), TransactOpts: *client.GetTransactOpts()}
// 			param := packedParam.GetParams()
// 			if len(param) != packedParam.ParamsLen() {
// 				panic("Param Length Error...Please check the code in paradigm!!!")
// 			}
// 			// todo 下面看起来好像只有一笔交易（同上），我原本的设想是每个commitSlot一个交易hash，当然现在这样也是每个一个Hash只不过重合了
// 			// 如果sdk就是这样子设计的，那就这样也行

// 			// 下面的receipt是交易上链后收到的回执，需要用到便于追溯
// 			_, receipt, err := slotCommitSession.CommitSlotsBatch(
// 				param[0].([][32]byte),
// 				param[1].([]*big.Int),
// 				param[2].([]*big.Int),
// 				param[3].([]*big.Int),
// 				param[4].([]*big.Int),
// 			)
// 			if err != nil {
// 				LogWriter.Log("ERROR", fmt.Sprintf("Failed to call CommitSlotsBatch: %v", err))
// 			} else {
// 				// 下面就是把所有的交易都依次推到dev的channel里
// 				ptxs := packedParam.BuildDevTransactions([]*types.Receipt{receipt})
// 				w.devPackedTransaction <- ptxs // 传递到dev
// 				continue
// 			}

// 		}

// 	}
// 	w.params = paradigm.NewParamsMap()

// }

//func Worker(id int, queue chan map[string]interface{}, instance *SlotCommit.SlotCommit, client *client.Client) {
//	// 每次收集 batchsize 个transaction再一起上链，但是实际产生的数据很少 设置为1，收到后直接上链
//	batchSize := 1
//	// 上链参数
//	var signs [][32]byte
//	var slotsBigInt, processesBigInt, nidsBigInt, epochsBigInt []*big.Int
//
//	for {
//		select {
//		case result := <-queue: // 尝试从通道中接收数据
//			if result != nil { // 判断是否接收到有效值
//				// log.Printf("Worker %d Received result: %v", id, result)
//				LogWriter.Log("CHAINUP", fmt.Sprintf("Worker %d Received Transaction: %v", id, result))
//
//				sign := result["Sign"].(string)
//				slot := result["Slot"].(int32)
//				process := result["Process"].(int32)
//				nid := result["ID"].(int32)
//				epoch := result["Epoch"].(int)
//
//				// 转换 sign 为 bytes32
//				signBytes32 := toBytes32(sign)
//
//				signs = append(signs, signBytes32)
//				slotsBigInt = append(slotsBigInt, new(big.Int).SetUint64(uint64(slot)))
//				processesBigInt = append(processesBigInt, new(big.Int).SetUint64(uint64(process)))
//				nidsBigInt = append(nidsBigInt, new(big.Int).SetUint64(uint64(nid)))
//				epochsBigInt = append(epochsBigInt, new(big.Int).SetUint64(uint64(epoch)))
//
//				// 每当收集到batchSize个transaction的信息时，调用批量上链函数
//				if len(signs) >= batchSize {
//					consumerImpl(signs, slotsBigInt, processesBigInt, nidsBigInt, epochsBigInt, instance, client)
//					LogWriter.Log("CHAINUP", fmt.Sprintf("Worker %d completed Batch transactions Upchain. Count: %d\n", id, len(signs)))
//					// 检查上链的数据是否成功 这边signs应该是有重复的，后续可以去重再查询
//					// getTransactionImpl(signs, instance, client)
//					// 清空数组
//					signs = nil
//					slotsBigInt = nil
//					processesBigInt = nil
//					nidsBigInt = nil
//					epochsBigInt = nil
//
//				}
//				// LogWriter.Log("CHAINUP", fmt.Sprintf("Worker %d: finished Transaction: %v\n", id, result))
//			} else {
//				LogWriter.Log("ERROR", fmt.Sprintf("Upchain channel closed, received nil value"))
//				return
//			}
//		}
//	}
//}

//// 调用合约函数进行批量数据上链
//func consumerImpl(signs [][32]byte, slotsBigInt []*big.Int, processesBigInt []*big.Int, nidsBigInt []*big.Int, epochsBigInt []*big.Int, instance *SlotCommit.SlotCommit, client *client.Client) {
//	slotCommitSession := &SlotCommit.SlotCommitSession{Contract: instance, CallOpts: *client.GetCallOpts(), TransactOpts: *client.GetTransactOpts()}
//
//	_, _, err := slotCommitSession.CommitSlotsBatch(
//		signs,
//		slotsBigInt,
//		processesBigInt,
//		nidsBigInt,
//		epochsBigInt,
//	)
//	if err != nil {
//		LogWriter.Log("ERROR", fmt.Sprintf("Failed to call CommitSlotsBatch: %v", err))
//	}
//}

//func getTransactionImpl(signs [][32]byte, instance *SlotCommit.SlotCommit, client *client.Client) {
//	slotCommitSession := &SlotCommit.SlotCommitSession{Contract: instance, CallOpts: *client.GetCallOpts(), TransactOpts: *client.GetTransactOpts()}
//	for id, sign := range signs {
//		ret, err := slotCommitSession.GetSlotCommits(sign)
//		if err != nil {
//			LogWriter.Log("ERROR", fmt.Sprintf("Failed to call getSlotBySign %d %v: %v", id, sign, err))
//		}
//		LogWriter.Log("CHAINUP", fmt.Sprintf("Sign %d: %v, get chain transaction: %v", id, sign, ret))
//	}
//
//}
