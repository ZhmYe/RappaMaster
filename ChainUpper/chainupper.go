package ChainUpper

import (
	"RappaMaster/ChainUpper/service"
	"RappaMaster/channel"
	"RappaMaster/fisco-bcos-client/contract/store"
	"RappaMaster/paradigm"
	"RappaMaster/transaction"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"sync"
	"time"

	"context"
	"encoding/hex"
	"log"

	"github.com/FISCO-BCOS/go-sdk/v3/client"
)

// ChainUpper handles the process of up-chain transaction
// we consider 2 conditions to trigger the up-chain process
// after a timeout or the transactionPool is full
type ChainUpper struct {
	ctx             *context.Context
	channel         *channel.RappaChannel
	transactionPool []transaction.Transaction
	mu              sync.Mutex
	queue           chan transaction.Transaction
	instance        *Store.Store
	count           int
	ticker          *time.Ticker
	batchSize       int
	threshold       int
}

func (c *ChainUpper) processAsyncTransaction(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case tx := <-c.channel.UpchainBuffer:
			c.mu.Lock()
			c.transactionPool = append(c.transactionPool, tx)
			c.mu.Unlock()
		}
	}
}
func (c *ChainUpper) upchain(txs []transaction.Transaction) {
	if len(txs) == 0 {
		return
	}
	check := func(tx transaction.Transaction) error {
		calldata := tx.CallData()
		switch tx.(type) {
		case *transaction.InitTaskTransaction:
			// todo
			return nil
		case *transaction.TaskProcessTransaction:
			if calldata["Process"].(int32) < 0 {
				return fmt.Errorf("TaskProcessTransaction Process <0")
			}
			// todo
			return nil
		case *transaction.EpochRecordTransaction:
			// todo
			return nil
		default:
			return nil
		}
	}
	ctrl := make()
	for _, tx := range txs {
		if err := check(tx); err != nil {
			paradigm.Log("ERROR", err.Error())
			continue
		} else {
			c.queue <- tx
		}
	}
}
func (c *ChainUpper) consume(w map[transaction.TransactionType]transaction.PackedParams) {
	client := c.client
	instance := c.instance
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

func (c *ChainUpper) Start(ctx context.Context) {
	go c.handleQuery()
	timeStart := time.Now()
	go func() {
		for {
			if time.Since(timeStart) >= 10*time.Second {
				timeStart = time.Now()
				c.UpChain()
			}
		}
	}()
}
func (c *ChainUpper) UpChain() {
	// 这里简单写一下，应该是用异步上链组件接入这里
	pack := func() []transaction.Transaction {
		c.mu.Lock()
		defer c.mu.Unlock()
		if c.unprocessedIndex == len(c.transactionPool) {
			return []transaction.Transaction{}
		}
		packedTransactions := c.transactionPool[c.unprocessedIndex:]
		c.unprocessedIndex = len(c.transactionPool)
		return packedTransactions
	}
	packedTransactions := pack()
	if len(packedTransactions) > 0 {
		// 将交易打包为链上合约的参数
		for _, tx := range packedTransactions {
			// modify by zhmye
			check := func(tx transaction.Transaction) error {
				calldata := tx.CallData()
				switch tx.(type) {
				// 这里可以写在Transaction的interface里，加一个Check()，然后下面统一tx.Check()
				case *transaction.InitTaskTransaction:
					// todo
					return nil
				case *transaction.TaskProcessTransaction:
					if calldata["Process"].(int32) < 0 {
						return fmt.Errorf("TaskProcessTransaction Process <0")
					}
					// todo
					return nil
				case *transaction.EpochRecordTransaction:
					// todo
					return nil
				default:
					return nil
				}
			}
			if err := check(tx); err != nil {
				//panic(err)
				paradigm.Log("ERROR", err.Error())
				continue
			} else {
				c.queue <- tx
			}
		}
		// LogWriter.Log("CHAINUP", fmt.Sprintf("%d Transactions pushed to queue for async processing", len(packedTransactions)))
		paradigm.Log("CHAINUP", fmt.Sprintf("up %d transactions to blockchain...", len(packedTransactions)))

	} else {
		//paradigm.Log("CHAINUP", "Nothing to up to Blockchain..., len(transactionPool) = 0")
	}
}

func NewChainUpper(channel *channel.RappaChannel, config *paradigm.RappaMasterConfig) (*ChainUpper, error) {
	// 初始化 FISCO-BCOS 客户端
	privateKey, _ := hex.DecodeString(config.PrivateKey)
	client, err := client.DialContext(context.Background(), &client.Config{
		IsSMCrypto:  false,
		GroupID:     config.GroupID,
		PrivateKey:  privateKey,
		Host:        config.FiscoBcosHost,
		Port:        config.FiscoBcosPort,
		TLSCaFile:   config.TLSCaFile,
		TLSCertFile: config.TLSCertFile,
		TLSKeyFile:  config.TLSKeyFile,
	})
	if err != nil {
		e := paradigm.Error(paradigm.RuntimeError, fmt.Sprintf("Failed to initialize FISCO-BCOS client: %v", err))
		//paradigm.Log("ERROR", fmt.Sprintf("failed to initialize FISCO-BCOS client: %v", err))
		return nil, fmt.Errorf(e.Error())
	}

	// 部署或加载合约
	// instance, err := Store.NewStore(common.HexToAddress(config.ContractAddress), client)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to load contract: %v", err)
	// }
	address, receipt, instance, err := Store.DeployStore(client.GetTransactOpts(), client)
	if err != nil {
		log.Fatal(err)
	}
	paradigm.Print("INFO", fmt.Sprintf("Deploy Contract on Blockchain Success, contract adddress: %s, transaction hash: %s", address.Hex(), receipt.TransactionHash))
	//paradigm.Print("INFO", fmt.Sprintf("contract address: %s", address.Hex())) // the address should be saved, will use in next example
	//paradigm.Log("INFO", fmt.Sprintf("transaction hash: %s", receipt.TransactionHash))

	// 初始化队列和 Worker
	queue := make(chan transaction.Transaction, config.QueueBufferSize)
	for i := 0; i < config.WorkerCount; i++ {
		worker := service.NewUpchainWorker(i, config.BatchSize, queue, channel.DevTransactionChannel, instance, client)
		go worker.Process()
		//go service. (i, queue, instance, client)
	}
	paradigm.Log("INFO", "Chainupper initialized successfully, workers waiting for transactions...")

	return &ChainUpper{
		channel: channel,
		//pendingTransactions: channel.PendingTransactions,
		transactionPool:  make([]transaction.Transaction, 0),
		unprocessedIndex: 0,
		mu:               sync.Mutex{},
		queue:            queue,
		client:           client,
		instance:         instance,
	}, nil
}
