package service

import (
	Store "BHLayer2Node/ChainUpper/contract/store"
	"BHLayer2Node/LogWriter"
	"BHLayer2Node/paradigm"
	"fmt"

	"github.com/FISCO-BCOS/go-sdk/v3/client"
	"github.com/FISCO-BCOS/go-sdk/v3/types"
)

// TODO @XQ 这里看着应该是和合约相关，要把合约更新，我这边改成了结构体,下面一些todo看一下如果改不了就按照现在的一些格式写
// 下面的参数我从现在的写法来看，好像是go自动生成的代码里需要把合约的参数全部列出来传到函数里，那这样似乎不同交易的batch比较麻烦？
// 下面目前是把不同交易的参数分开了,不同的交易分开batch?

// UpChainWorker modify by zhmye
type UpChainWorker struct {
	id                   int
	queue                chan paradigm.Transaction
	devPackedTransaction chan []*paradigm.PackedTransaction // add by zhmye 这里是用来给dev的，所有已经上链的交易都要给
	instance             *Store.Store
	client               *client.Client
	batchSize            int
	//signs                                                  [][32]byte
	//slotsBigInt, processesBigInt, nidsBigInt, epochsBigInt []*big.Int
	params map[paradigm.TransactionType]paradigm.PackedParams // 这里记录各种类型的交易参数 add by zhmye
	count  int                                                // 交易计数
}

func (w *UpChainWorker) Process() {
	for {
		select {
		case tx := <-w.queue: // 尝试从通道中接收数据
			if tx != nil { // 判断是否接收到有效值
				// log.Printf("Worker %d Received result: %v", id, result)
				LogWriter.Log("CHAINUP", fmt.Sprintf("Worker %d Received Transaction: %v", w.id, tx))

				switch tx.(type) {
				case *paradigm.InitTaskTransaction:
					w.params[paradigm.INIT_TASK_TRANSACTION].UpdateFromTransaction(tx)
				case *paradigm.TaskProcessTransaction:
					w.params[paradigm.TASK_PROCESS_TRANSACTION].UpdateFromTransaction(tx)
				case *paradigm.EpochRecordTransaction:
					w.params[paradigm.EPOCH_RECORD_TRANSACTION].UpdateFromTransaction(tx)
				default:
					panic("Invalid Transaction Type!!!")
				}
				w.count++
				if w.count >= w.batchSize {
					// 每当收集到batchSize个transaction的信息时，调用批量上链函数
					w.consumer()
					w.count = 0

				}
			} else {
				LogWriter.Log("ERROR", fmt.Sprintf("Upchain channel closed, received nil value"))
				return
			}
		}
	}
}

// consumer 将当前收集的交易参数转换为 KV 对后批量调用上链函数
func (w *UpChainWorker) consumer() {
	client := w.client
	instance := w.instance
	// 对每种交易类型，都调用 ConvertParamsToKVPairs 得到 KV 对，然后调用合约函数 setItems
	for tType, packedParam := range w.params {
		if packedParam.IsEmpty() {
			// txs := packedParam.GetParams()
			// LogWriter.Log("DEBUG", fmt.Sprintf("TX_%d waiting for upchain: %s", tType, txs))
			continue
		}
		// key, value := packedParam.ConvertParamsToKVPairs()
		keys, values := packedParam.ParamsToKVPairs()
		// LogWriter.Log("DEBUG", fmt.Sprintf("TX_%d convert to:[key]%s [value]%s", tType, keys, values))
		storeSession := &Store.StoreSession{Contract: instance, CallOpts: *client.GetCallOpts(), TransactOpts: *client.GetTransactOpts()}
		// _, receipt, err := storeSession.SetItem(key, value)
		_, receipt, err := storeSession.SetItems(keys, values)
		if err != nil {
			LogWriter.Log("ERROR", fmt.Sprintf("Worker %d Failed to call SetItems for type %v: %v", w.id, tType, err))
		} else {
			LogWriter.Log("CHAINUP", fmt.Sprintf("Worker %d: transactions up-chained successfully for TX_%v.", w.id, tType))
			ptxs := packedParam.BuildDevTransactions([]*types.Receipt{receipt})
			w.devPackedTransaction <- ptxs // 传递到dev
		}

	}
	w.params = paradigm.NewParamsMap()
}

func NewUpchainWorker(id int, batchSize int, queue chan paradigm.Transaction, dev chan []*paradigm.PackedTransaction, instance *Store.Store, client *client.Client) *UpChainWorker {
	return &UpChainWorker{
		id:                   id,
		queue:                queue,
		devPackedTransaction: dev,
		instance:             instance,
		client:               client,
		batchSize:            batchSize,
		params:               paradigm.NewParamsMap(),
		count:                0,
	}
}

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
