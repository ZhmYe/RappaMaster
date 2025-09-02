package service

import (
	"RappaMaster/fisco-bcos-client/contract/store"
	"RappaMaster/paradigm"
	"RappaMaster/transaction"
	"context"
	"fmt"
	"time"

	"github.com/FISCO-BCOS/go-sdk/v3/client"
	"github.com/FISCO-BCOS/go-sdk/v3/types"
	"github.com/ethereum/go-ethereum/common"
)

// UpChainWorker modify by zhmye
type UpChainWorker struct {
	id                   int
	queue                chan transaction.Transaction
	devPackedTransaction chan []*transaction.PackedTransaction // add by zhmye 这里是用来给dev的，所有已经上链的交易都要给
	instance             *Store.Store
	client               *client.Client
	batchSize            int
	params               map[transaction.TransactionType]transaction.PackedParams // 这里记录各种类型的交易参数 add by zhmye
	count                int                                                      // 交易计数
}

func (w *UpChainWorker) Process() {
	for {
		select {
		case tx := <-w.queue: // 尝试从通道中接收数据
			if tx != nil { // 判断是否接收到有效值
				// log.Printf("Worker %d Received result: %v", id, result)
				//paradigm.Log("CHAINUP", fmt.Sprintf("Worker %d Received Transaction: %v", w.id, tx))

				switch tx.(type) {
				case *transaction.InitTaskTransaction:
					w.params[transaction.INIT_TASK_TRANSACTION].UpdateFromTransaction(tx)
				case *transaction.TaskProcessTransaction:
					w.params[transaction.TASK_PROCESS_TRANSACTION].UpdateFromTransaction(tx)
				case *transaction.EpochRecordTransaction:
					w.params[transaction.EPOCH_RECORD_TRANSACTION].UpdateFromTransaction(tx)
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
				paradigm.Log("ERROR", fmt.Sprintf("Upchain channel closed, received nil value"))
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
			paradigm.Error(paradigm.RuntimeError, fmt.Sprintf("Worker %d Failed to call SetItems for type %v: %v", w.id, tType, err))
		}
		// 获得有merkleProof的receipt
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_receipt, err := w.client.GetTransactionReceipt(ctx, common.HexToHash(receipt.TransactionHash), true)
		if err != nil {
			paradigm.Error(paradigm.RuntimeError, fmt.Sprintf("Failed to getReceipt with merkleProof for type %v: %v", tType, err))
		} else {
			// LogWriter.Log("DEBUG", fmt.Sprintf("Receipt with merkleProof: %s", _receipt)) //debug
			// block, _ := w.client.GetBlockByNumber(ctx, int64(_receipt.BlockNumber), false, false)
			blockHash, _ := w.client.GetBlockHashByNumber(ctx, int64(_receipt.BlockNumber))
			ptxs := packedParam.BuildDevTransactions([]*types.Receipt{_receipt}, blockHash.Hex())
			w.devPackedTransaction <- ptxs // 传递到dev
		}

	}
	w.params = transaction.NewParamsMap()
}

func NewUpchainWorker(id int, batchSize int, queue chan transaction.Transaction, dev chan []*transaction.PackedTransaction, instance *Store.Store, client *client.Client) *UpChainWorker {
	return &UpChainWorker{
		id:                   id,
		queue:                queue,
		devPackedTransaction: dev,
		instance:             instance,
		client:               client,
		batchSize:            batchSize,
		params:               transaction.NewParamsMap(),
		count:                0,
	}
}
