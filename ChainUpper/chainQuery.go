package ChainUpper

//
//import (
//	"context"
//	"encoding/hex"
//	"fmt"
//	"time"
//
//	Store "BHLayer2Node/ChainUpper/contract/store"
//	"BHLayer2Node/Config"
//	"BHLayer2Node/LogWriter"
//	"BHLayer2Node/paradigm"
//
//	"github.com/FISCO-BCOS/go-sdk/v3/client"
//	"github.com/ethereum/go-ethereum/common"
//)
//
//// ChainQuery 用于初始化链上查询服务，并持久运行 QueryWorker
//type ChainQuery struct {
//	channel  *paradigm.RappaChannel
//	queue    chan paradigm.QueryRequest // 用于将查询请求发送给queryWorker
//	client   *client.Client
//	instance *Store.Store
//}
//
//// Start 方法启动 ChainQuery 服务，持续读取全局 RequestChannel 中的查询请求，转发到内部队列；
//// 同时启动一个模拟查询请求的 goroutine（可用于调试测试）。
//func (cq *ChainQuery) Start() {
//	// 启动一个 goroutine 从全局 RequestChannel 中读取查询请求，并转发到内部队列（cq.queue）
//	go func() {
//		for {
//			req, ok := <-cq.channel.QueryChannel
//			if !ok {
//				LogWriter.Log("ERROR", "Global RequestChannel closed")
//				return
//			}
//			// 转发请求到内部队列
//			cq.queue <- req
//			LogWriter.Log("QUERY", fmt.Sprintf("ChainQuery: Forwarded query request (ID: %s) to internal queue", req.RequestID))
//		}
//	}()
//
//	// 启动一个模拟查询请求的 goroutine（测试用）
//	go cq.simulateQueries()
//}
//
//// simulateQueries 模拟生成查询请求
//func (cq *ChainQuery) simulateQueries() {
//	for {
//		// 1. EpochNumQuery：查询当前区块链的区块数量
//		req1 := paradigm.QueryRequest{
//			QueryType: paradigm.EpochNumQuery,
//			Params:    map[string]interface{}{}, // 此查询类型无需参数
//			RequestID: fmt.Sprintf("sim-%d", time.Now().Unix()),
//			Timestamp: time.Now().Unix(),
//		}
//		cq.channel.QueryChannel <- req1 // 发送到全局 RequestChannel
//		LogWriter.Log("QUERY", fmt.Sprintf("SimulateQueries: Sent simulated EpochNumQuery request (ID: %s)", req1.RequestID))
//		res1 := <-cq.channel.ResponseChannel
//		LogWriter.Log("QUERY", fmt.Sprintf("SimulateQueries: Received query result: %+v", res1))
//
//		// 2. TxNumQuery：查询当前最新区块的交易数量
//		req2 := paradigm.QueryRequest{
//			QueryType: paradigm.TxNumQuery,
//			Params:    map[string]interface{}{},
//			RequestID: fmt.Sprintf("sim-txnum-%d", time.Now().Unix()),
//			Timestamp: time.Now().Unix(),
//		}
//		cq.channel.QueryChannel <- req2
//		LogWriter.Log("QUERY", fmt.Sprintf("SimulateQueries: Sent simulated TxNumQuery request (ID: %s)", req2.RequestID))
//		res2 := <-cq.channel.ResponseChannel
//		LogWriter.Log("QUERY", fmt.Sprintf("SimulateQueries: Received TxNumQuery result: %+v", res2))
//
//		// // 3. BlockInfoQuery：查询区块详情
//		// req3 := paradigm.QueryRequest{
//		// 	QueryType: paradigm.BlockInfoQuery,
//		// 	Params: map[string]interface{}{
//		// 		"blockHash": "0x74aa056884cfd5ad10b30798d8fc6c28e0a71d43bfc7c3a10de8f3808328480b", // 请替换为存在的 blockHash
//		// 	},
//		// 	RequestID: fmt.Sprintf("sim-blockinfo-%d", time.Now().Unix()),
//		// 	Timestamp: time.Now().Unix(),
//		// }
//		// cq.channel.QueryChannel <- req3
//		// LogWriter.Log("QUERY", fmt.Sprintf("SimulateQueries: Sent simulated BlockInfoQuery request (ID: %s)", req3.RequestID))
//		// res3 := <-cq.channel.ResponseChannel
//		// LogWriter.Log("QUERY", fmt.Sprintf("SimulateQueries: Received BlockInfoQuery result: %+v", res3))
//
//		// // 4. TxInfoQuery：查询交易详情（此处使用一个示例 txHash）
//		// req4 := paradigm.QueryRequest{
//		// 	QueryType: paradigm.TxInfoQuery,
//		// 	Params: map[string]interface{}{
//		// 		"txHash": "0x7d4d7a40fd6dbddf33362ea71e02470deec73e7e6f47b31069d6d8d0d4801260", // 替换为存在的 txHash
//		// 	},
//		// 	RequestID: fmt.Sprintf("sim-txinfo-%d", time.Now().Unix()),
//		// 	Timestamp: time.Now().Unix(),
//		// }
//		// cq.channel.QueryChannel <- req4
//		// LogWriter.Log("QUERY", fmt.Sprintf("SimulateQueries: Sent simulated TxInfoQuery request (ID: %s)", req4.RequestID))
//		// res4 := <-cq.channel.ResponseChannel
//		// LogWriter.Log("QUERY", fmt.Sprintf("SimulateQueries: Received TxInfoQuery result: %+v", res4))
//
//		time.Sleep(5 * time.Second)
//	}
//
//}
//
//// NewChainQuery 根据配置初始化 FISCO-BCOS 客户端、加载已部署的合约实例，并构造 QueryWorker
//func NewChainQuery(channel *paradigm.RappaChannel, config *Config.BHLayer2NodeConfig) (*ChainQuery, error) {
//	privateKey, err := hex.DecodeString(config.PrivateKey)
//	if err != nil {
//		LogWriter.Log("ERROR", fmt.Sprintf("Failed to decode private key: %v", err))
//		return nil, err
//	}
//	client, err := client.DialContext(context.Background(), &client.Config{
//		IsSMCrypto:  false,
//		GroupID:     config.GroupID,
//		PrivateKey:  privateKey,
//		Host:        config.FiscoBcosHost,
//		Port:        config.FiscoBcosPort,
//		TLSCaFile:   config.TLSCaFile,
//		TLSCertFile: config.TLSCertFile,
//		TLSKeyFile:  config.TLSKeyFile,
//	})
//	if err != nil {
//		LogWriter.Log("ERROR", fmt.Sprintf("Failed to initialize FISCO-BCOS client: %v", err))
//		return nil, err
//	}
//
//	// 使用配置中的合约地址加载 Store 合约实例
//	addr := common.HexToAddress(config.ContractAddress)
//	instance, err := Store.NewStore(addr, client)
//	if err != nil {
//		LogWriter.Log("ERROR", fmt.Sprintf("Failed to load store contract: %v", err))
//		return nil, err
//	}
//	// LogWriter.Log("INFO", fmt.Sprintf("Loaded store contract at address: %s", config.ContractAddress))
//
//	queue := make(chan paradigm.QueryRequest, config.QueueBufferSize)
//	// 启动多个queryWorker处理查询请求
//	for i := 0; i < config.WorkerCount; i++ {
//		queryWorker := NewQueryWorker(i, queue, channel.ResponseChannel, client, instance)
//		go queryWorker.Process()
//	}
//
//	return &ChainQuery{
//		channel:  channel,
//		queue:    queue,
//		client:   client,
//		instance: instance,
//	}, nil
//}
