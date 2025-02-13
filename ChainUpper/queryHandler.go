package ChainUpper

import (
	"BHLayer2Node/Query"
	"BHLayer2Node/paradigm"
	"context"
	"fmt"
	"time"

	"github.com/FISCO-BCOS/go-sdk/v3/types"
	"github.com/ethereum/go-ethereum/common"
)

func (c *ChainUpper) handleQuery() {
	for query := range c.channel.BlockchainQueryChannel {
		c.handle(query)
	}
}

func (c *ChainUpper) handle(query paradigm.Query) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	switch query.(type) {
	// 交易和最新信息的query都在oracle处理了,TODO @XQ 判断一下在oracle里的信息是否完整/正确
	case *Query.BlockchainBlockHashQuery:
		// 通过client获取到block
		item := query.(*Query.BlockchainBlockHashQuery)
		blockHash := item.BlockHash
		if blockHash == "" {
			item.SendInfo(paradigm.NewInvalidBlockInfo("blockHash parameter missing or invalid"))
			return
			//return *paradigm.NewErrorResponse(-1, "blockHash parameter missing or invalid", int(req.QueryType))
		}
		hash := common.HexToHash(blockHash)
		block, err := c.client.GetBlockByHash(ctx, hash, false, false)
		if err != nil {
			item.SendInfo(paradigm.NewInvalidBlockInfo(fmt.Sprintf("Failed to get block: %v", err)))
			//return *paradigm.NewErrorResponse(-1, fmt.Sprintf("Failed to get block: %v", err), int(req.QueryType))
			return
		}
		blockInfo := c.getBlockInfo(*block)
		item.SendInfo(blockInfo)

	case *Query.BlockchainBlockNumberQuery:
		// 通过client获取到block
		item := query.(*Query.BlockchainBlockNumberQuery)
		blockNumber := item.BlockNumber
		block, err := c.client.GetBlockByNumber(ctx, int64(blockNumber), false, false)
		if err != nil {
			item.SendInfo(paradigm.NewInvalidBlockInfo(fmt.Sprintf("Failed to get block: %v", err)))
			return
		}
		blockInfo := c.getBlockInfo(*block)
		item.SendInfo(blockInfo)

		// case *Query.BlockchainTransactionQuery:
		// 	item := query.(*Query.BlockchainTransactionQuery)
		// 	txHash := common.HexToHash(item.TxHash)
		// 	tx, err := c.client.GetTransactionByHash(ctx, txHash, false)
		// 	if err != nil {
		// 		item.SendBlockchainInfo(paradigm.NewInvalidTransactionInfo(fmt.Sprintf("Failed to get transaction: %v", err)))
		// 		return
		// 	}
		// 	abi := tx.Abi
		// 	// // 通过tx查询获得receipt -> receipt.blockNumber -> blockhash
		// 	// receipt, err := c.client.GetTransactionReceipt(ctx, txHash, false)
		// 	// blockHash, err := c.client.GetBlockHashByNumber(ctx, int64(receipt.BlockNumber))
		// 	// txInfo := c.getTransactionInfo(*tx, blockHash.Hex())
		// 	receipt, err := c.client.GetTransactionReceipt(ctx, txHash, false)
		// 	if err != nil {
		// 		item.SendBlockchainInfo(paradigm.NewInvalidTransactionInfo(fmt.Sprintf("Failed to get transaction: %v", err)))
		// 		return
		// 	}
		// 	blockHash, err := c.client.GetBlockHashByNumber(ctx, int64(receipt.BlockNumber))
		// 	if err != nil {
		// 		item.SendBlockchainInfo(paradigm.NewInvalidTransactionInfo(fmt.Sprintf("Failed to get transaction blockHash: %v", err)))
		// 		return
		// 	}
		// 	txInfo := c.getTransactionInfo(receipt, blockHash.Hex(), abi)
		// 	item.SendBlockchainInfo(txInfo)

		item.SendInfo(blockInfo)
	default:
		paradigm.Error(paradigm.RuntimeError, "Unsupported Query Type In ChainUpper")
	}
}

func (c *ChainUpper) getBlockInfo(block types.Block) paradigm.BlockInfo {
	var parentHash string
	if len(block.ParentInfo) > 0 {
		// 取第一个父区块的哈希（通常只有一个父区块）
		parentHash = block.ParentInfo[0].GetBlockHash()
	} else {
		parentHash = "genesis" // 创世区块没有父哈希
	}

	//var txDetails []TxDetail
	txs := make([]paradigm.TransactionInfo, 0) // 只要txHash，剩余在Oracle里获取
	// LogWriter.Log("DEBUG", fmt.Sprintf("Block.GetTransactions() return %s", block.GetTransactions()...))
	for _, tx := range block.Transactions {
		// 首先尝试将 tx 转换为 *types.TransactionDetail
		//var txHash string
		if txObj, ok := tx.(*types.TransactionDetail); ok {
			txs = append(txs, paradigm.TransactionInfo{
				TxHash:       txObj.GetHash(),
				Contract:     txObj.GetTo(),
				Abi:          txObj.GetAbi(),
				BlockHash:    block.Hash,
				Invalid:      true,
				ErrorMessage: "",
			})
			//txs = append(txs, txObj.Hash().String())
			//c.client.GetTransactionReceipt()
			//		txDetails = append(txDetails, TxDetail{
			//			TxHash:   txObj.DataHash.String(),
			//			Contract: txObj.To().String(),
			//			Method:   "Unknown", // todo: 无法直接获取调用接口，此处用占位符
			//		})
		} else {
			//		// 如果转换失败，尝试转换为 map[string]interface{}
			if txMap, ok := tx.(map[string]interface{}); ok {
				// TODO @XQ 这里修正一下，包括exist的判断
				txs = append(txs, paradigm.TransactionInfo{
					TxHash:       txMap["hash"].(string),
					Contract:     txMap["to"].(string),
					Abi:          txMap["abi"].(string),
					BlockHash:    block.Hash,
					Invalid:      true,
					ErrorMessage: "",
				})

				//txHash, _ := txMap["DataHash"].(string)
				//txs = append(txs, txHash)
				//			contract, _ := txMap["to"].(string)
				//			method, _ := txMap["method"].(string)
				//			if method == "" {
				//				method = "Unknown"
				//			}
				//			txDetails = append(txDetails, TxDetail{
				//				TxHash:   txHash,
				//				Contract: contract,
				//				Method:   method,
				//			})
				//		}
			}
		}
	}
	return paradigm.BlockInfo{
		BlockHash:       block.Hash,
		ParentHash:      parentHash,
		BlockHeight:     int32(block.Number),
		TransactionRoot: block.TxsRoot,
		Txs:             txs,
		Invalid:         true,
		ErrorMessage:    "",
	}
}

func (c *ChainUpper) getTransactionInfo(receipt *types.Receipt, blockHash string, abi string) paradigm.TransactionInfo {
	return paradigm.TransactionInfo{
		TxHash:   receipt.TransactionHash,
		Contract: receipt.To,
		// Abi:       "processTask", // todo
		Abi:       abi, // 检查是否有值
		BlockHash: blockHash,
	}
}
