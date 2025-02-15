package ChainUpper

//
//import (
//	"context"
//	"encoding/json"
//	"fmt"
//	"time"
//
//	Store "BHLayer2Node/ChainUpper/contract/store"
//	"BHLayer2Node/LogWriter"
//	"BHLayer2Node/paradigm"
//
//	"github.com/FISCO-BCOS/go-sdk/v3/client"
//	"github.com/FISCO-BCOS/go-sdk/v3/types"
//	"github.com/ethereum/go-ethereum/common"
//)
//
//// QueryWorker 结构体新增处理器映射
//type QueryWorker struct {
//	id int
//
//	client   *client.Client
//	instance *Store.Store
//	handlers map[paradigm.DevQueryType]QueryHandler // 新增处理器映射
//}
//
//// 初始化处理器映射
//func NewQueryWorker(
//	id int,
//	queue chan paradigm.QueryRequest,
//	ResponseChannel chan paradigm.QueryResponse,
//	client *client.Client,
//	instance *Store.Store,
//) *QueryWorker {
//	qw := &QueryWorker{
//		id:              id,
//		queue:           queue,
//		ResponseChannel: ResponseChannel,
//		client:          client,
//		instance:        instance,
//		handlers:        make(map[paradigm.DevQueryType]QueryHandler),
//	}
//
//	// 注册处理器
//	qw.RegisterHandler(paradigm.EpochNumQuery, &EpochNumHandler{})
//	qw.RegisterHandler(paradigm.TxNumQuery, &TxNumHandler{})
//	qw.RegisterHandler(paradigm.BlockInfoQuery, &BlockInfoHandler{})
//	qw.RegisterHandler(paradigm.TxInfoQuery, &TxInfoHandler{})
//
//	return qw
//}
//
//// 注册处理器方法
//func (qw *QueryWorker) RegisterHandler(queryType paradigm.DevQueryType, handler QueryHandler) {
//	qw.handlers[queryType] = handler
//}
//
//func (qw *QueryWorker) Process() {
//	for req := range qw.queue {
//		go func(req paradigm.QueryRequest) {
//			qType := req.QueryType
//			handler, exists := qw.handlers[qType] // 根据查询类型调用对应的处理器
//
//			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//			defer cancel()
//
//			var res paradigm.QueryResponse
//			if !exists {
//				res = *paradigm.NewErrorResponse(-1, "Unsupported query type", int(qType))
//			} else {
//				res = handler.Handle(ctx, qw, req)
//			}
//
//			qw.ResponseChannel <- res
//			LogWriter.Log("QUERY", fmt.Sprintf("QueryWorker %d processed request: %+v", qw.id, res))
//		}(req)
//	}
//}
//
//// 具体处理器实现
//// EpochNumHandler 处理区块总数查询
//type EpochNumHandler struct{}
//
//func (h *EpochNumHandler) Handle(ctx context.Context, qw *QueryWorker, req paradigm.QueryRequest) paradigm.QueryResponse {
//	blockNum, err := qw.client.GetBlockNumber(ctx)
//	if err != nil {
//		return *paradigm.NewErrorResponse(-1, fmt.Sprintf("Failed to get block number: %v", err), int(req.QueryType))
//	}
//	return *paradigm.NewSuccessResponse(blockNum, int(req.QueryType))
//}
//
//// TxNumHandler 总交易数量查询
//type TxNumHandler struct{}
//
//func (h *TxNumHandler) Handle(ctx context.Context, qw *QueryWorker, req paradigm.QueryRequest) paradigm.QueryResponse {
//	TransactionCount, err := qw.client.GetTotalTransactionCount(ctx)
//	if err != nil {
//		return *paradigm.NewErrorResponse(-1, fmt.Sprintf("Failed to get Transaction number: %v", err), int(req.QueryType))
//	}
//	return *paradigm.NewSuccessResponse(TransactionCount.TxSum, int(req.QueryType))
//}
//
//// BlockInfoHandler 处理区块信息查询
//type BlockInfoHandler struct{}
//
//// TxDetail 定义交易详情结构
//type TxDetail struct {
//	TxHash   string `json:"txHash"`
//	Contract string `json:"contract"`
//	Method   string `json:"method"`
//}
//
//func (h *BlockInfoHandler) Handle(ctx context.Context, qw *QueryWorker, req paradigm.QueryRequest) paradigm.QueryResponse {
//	blockHash, ok := req.Params["blockHash"].(string)
//	if !ok || blockHash == "" {
//		return *paradigm.NewErrorResponse(-1, "blockHash parameter missing or invalid", int(req.QueryType))
//	}
//
//	hash := common.HexToHash(blockHash)
//	block, err := qw.client.GetBlockByHash(ctx, hash, false, false)
//	if err != nil {
//		return *paradigm.NewErrorResponse(-1, fmt.Sprintf("Failed to get block: %v", err), int(req.QueryType))
//	}
//	var parentHash string
//	if len(block.ParentInfo) > 0 {
//		// 取第一个父区块的哈希（通常只有一个父区块）
//		parentHash = block.ParentInfo[0].GetBlockHash()
//	} else {
//		parentHash = "genesis" // 创世区块没有父哈希
//	}
//
//	var txDetails []TxDetail
//	for _, tx := range block.Transactions {
//		// 首先尝试将 tx 转换为 *types.Transaction
//		if txObj, ok := tx.(*types.Transaction); ok {
//			txDetails = append(txDetails, TxDetail{
//				TxHash:   txObj.DataHash.String(),
//				Contract: txObj.To().String(),
//				Method:   "Unknown", // todo: 无法直接获取调用接口，此处用占位符
//			})
//		} else {
//			// 如果转换失败，尝试转换为 map[string]interface{}
//			if txMap, ok := tx.(map[string]interface{}); ok {
//				txHash, _ := txMap["DataHash"].(string)
//				contract, _ := txMap["to"].(string)
//				method, _ := txMap["method"].(string)
//				if method == "" {
//					method = "Unknown"
//				}
//				txDetails = append(txDetails, TxDetail{
//					TxHash:   txHash,
//					Contract: contract,
//					Method:   method,
//				})
//			}
//		}
//	}
//	blockInfo := map[string]interface{}{
//		"blockHash":      block.Hash,
//		"parentHash":     parentHash,
//		"blockHeight":    block.Number,
//		"merkleRoot":     block.StateRoot,
//		"nbTransactions": len(block.Transactions),
//		"transactions":   txDetails,
//	}
//	return *paradigm.NewSuccessResponse(blockInfo, int(req.QueryType))
//}
//
//// TxInfoHandler 处理交易信息查询
//type TxInfoHandler struct{}
//
//// 交易信息需要返回的参数：
//// TransactionDeial包含：Hash, contractAddress
//// receipt包含：blockNum
//// blockhash, 合约接口(?)
//
//func (h *TxInfoHandler) Handle(ctx context.Context, qw *QueryWorker, req paradigm.QueryRequest) paradigm.QueryResponse {
//	txHash, ok := req.Params["txHash"].(string)
//	if !ok || txHash == "" {
//		return *paradigm.NewErrorResponse(-1, "txHash parameter missing or invalid", int(req.QueryType))
//	}
//	hash := common.HexToHash(txHash)
//	tx, err := qw.client.GetTransactionByHash(ctx, hash, false)
//	if err != nil {
//		return *paradigm.NewErrorResponse(-1, fmt.Sprintf("Failed to get transaction: %v", err), int(req.QueryType))
//	}
//	// todo 这里应该和dev交互 所在区块Number能从receipt获得 调用GetBlockHashByNumber
//	// 这里暂时使用GetTransactionReceipt
//	receipt, err := qw.client.GetTransactionReceipt(ctx, hash, false)
//	blockNum := receipt.GetBlockNumber()
//	blockHash, err := qw.client.GetBlockHashByNumber(ctx, int64(blockNum))
//	// 直接将TransactionDetail结构体打包
//	txJSON, err := json.Marshal(tx)
//	if err != nil {
//		return *paradigm.NewErrorResponse(-1, fmt.Sprintf("Marshal failed: %v", err), int(req.QueryType))
//	}
//	TransactionInfo := map[string]interface{}{
//		"TransactionDetail": string(txJSON),
//		"hash":              tx.GetHash(),
//		"contractAddr":      tx.GetTo(),
//		"blockNum":          blockNum,
//		"blockHash":         blockHash,
//		"txType":            "Unknown",
//	}
//
//	return *paradigm.NewSuccessResponse(TransactionInfo, int(req.QueryType))
//}
