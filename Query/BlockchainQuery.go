package Query

import (
	"BHLayer2Node/paradigm"
	"time"
)

/***
	区块链信息界面
***/

// BlockchainQuery 需要和链交互，因此有一个给client传递消息的channel
type BlockchainQuery struct {
	paradigm.BasicChannelQuery
	sdkChannel chan interface{}
}

func (q *BlockchainQuery) SendBlockchainInfo(info interface{}) {
	q.sdkChannel <- info
	close(q.sdkChannel)
}
func (q *BlockchainQuery) ReceiveBlockchainInfo() interface{} {
	return <-q.sdkChannel
}
func NewBlockchainQuery() BlockchainQuery {
	return BlockchainQuery{
		BasicChannelQuery: paradigm.NewBasicChannelQuery(),
		sdkChannel:        make(chan interface{}),
	}
}

// BlockchainLatestInfoQuery 获取最新的区块链信息
// 1. 左上角的信息
// 2. 最新纪元
// 3. 最新交易
type BlockchainLatestInfoQuery struct {
	paradigm.BasicChannelQuery // 这个是从orac
	// 这里没有别的参数
}

func (q *BlockchainLatestInfoQuery) GenerateResponse(data interface{}) paradigm.Response {
	info := data.(paradigm.LatestBlockchainInfo) // 传入内容很多
	response := make(map[string]interface{})
	response["NbBlock"] = info.NbBlock             // 区块数量
	response["NbTransaction"] = info.NbTransaction // 交易数量
	response["NbEpoch"] = info.NbEpoch             // 纪元数量
	response["SynthData"] = info.SynthData         // 历史合成数量
	response["NbFinalized"] = info.NbFinalized     // 完成提交数
	le := make([]map[string]interface{}, 0)
	for _, epoch := range info.LatestEpoch {
		le = append(le, map[string]interface{}{
			"EpochID":     epoch.EpochID,
			"NbCommit":    len(epoch.Commits),
			"NbJustified": len(epoch.Justifieds),
			"NbFinalized": len(epoch.Finalizes),
			"TxHash":      epoch.TxReceipt.TransactionHash,
		})
	}
	response["LatestEpoch"] = le
	lt := make([]map[string]interface{}, 0)
	for _, tx := range info.LatestTxs {
		txType := "InitTask"
		switch tx.Tx.(type) {
		case *paradigm.InitTaskTransaction:
			txType = "InitTask"
		case *paradigm.TaskProcessTransaction:
			txType = "TaskProcess"
		case *paradigm.EpochRecordTransaction:
			txType = "EpochRecord"
		default:
			continue
		}
		lt = append(lt, map[string]interface{}{
			"txHash":      tx.Receipt.TransactionHash,
			"txType":      txType,
			"blockHash":   tx.BlockHash, // TODO @XQ 这里能否有区块哈希，如果没有，就改成blockHeight
			"contract":    tx.Receipt.To,
			"upchainTime": time.Now(), // TODO
		})
	}
	response["LatestTx"] = lt
	return paradigm.NewSuccessResponse(response)
}
func (q *BlockchainLatestInfoQuery) ParseRawDataFromHttpEngine(rawData map[interface{}]interface{}) bool {
	return true
}
func (q *BlockchainLatestInfoQuery) ToHttpJson() map[string]interface{} {
	return map[string]interface{}{"query": "BlockchainLatestInfoQuery"}
}

// BlockchainBlockInfoQuery 查询某个区块，允许块高或区块哈希
type BlockchainBlockInfoQuery struct {
	BlockchainQuery
}

func (q *BlockchainBlockInfoQuery) GenerateResponse(data interface{}) paradigm.Response {
	block := data.(paradigm.BlockInfo) // 区块
	response := make(map[string]interface{})
	response["blockHash"] = block.BlockHash
	response["parentHash"] = block.ParentHash
	response["blockHeight"] = block.BlockHeight
	response["nbTransaction"] = len(block.Txs)
	response["txRoot"] = block.TransactionRoot // 交易的merkle root
	response["txs"] = block.Txs                // TODO @XQ 这里是否可以转换，我看它是interface{}
	// TODO 另外Block结构体我看已经有json标签了，按道理是不是可以直接转成json
	//jsonData, err := json.Marshal(block)
	//if err != nil {
	//	log.Fatal(err)
	//}
	return paradigm.NewSuccessResponse(response)
}

// BlockchainBlockNumberQuery 根据区块高度查询区块
type BlockchainBlockNumberQuery struct {
	BlockNumber int32 // 区块高度
	BlockchainBlockInfoQuery
}

func (q *BlockchainBlockNumberQuery) ParseRawDataFromHttpEngine(rawData map[interface{}]interface{}) bool {
	if blockNumber, ok := rawData["blockNumber"]; ok {
		q.BlockNumber = int32(blockNumber.(int))
		return true
	}
	return false
}
func (q *BlockchainBlockNumberQuery) ToHttpJson() map[string]interface{} {
	return map[string]interface{}{"query": "BlockchainBlockNumberQuery", "blockNumber": q.BlockNumber}
}

// BlockchainBlockHashQuery 根据区块哈希查询区块
type BlockchainBlockHashQuery struct {
	BlockHash string // 区块高度
	BlockchainBlockInfoQuery
}

func (q *BlockchainBlockHashQuery) ParseRawDataFromHttpEngine(rawData map[interface{}]interface{}) bool {
	if blockHash, ok := rawData["blockHash"]; ok {
		q.BlockHash = blockHash.(string)
		return true
	}
	return false
}
func (q *BlockchainBlockHashQuery) ToHttpJson() map[string]interface{} {
	return map[string]interface{}{"query": "BlockchainBlockHashQuery", "blockHash": q.BlockHash}
}

// BlockchainTransactionQuery 查询交易，只能查询交易Hash
type BlockchainTransactionQuery struct {
	TxHash string
	BlockchainQuery
	// paradigm.BasicChannelQuery
}

// func (q *BlockchainTransactionQuery) GenerateResponse(data interface{}) paradigm.Response {
// 	ref := data.(paradigm.DevReference) // 交易reference TODO @XQ 我看到你写的TransactionDetails里没有区块信息部分，要从oracle交互的话我就直接用这个了
// 	// TODO 但是要确认一点: 就是是否所有区块中的交易都会被记录在oracle里，我这边反正就如果发现没有ref，那么说不存在于oracle了
// 	response := make(map[string]interface{})
// 	response["txHash"] = ref.TxReceipt.TransactionHash // TODO 这个hash和details的Hash是一样的吗
// 	response["blockNumber"] = ref.TxReceipt.BlockNumber
// 	response["contract"] = ref.TxReceipt.To
// 	response["txBlockHash"] = ref.TxBlockHash
// 	// TODO 区块哈希，考虑要不要加上，这个好像和另外某个地方的todo是一样的，最终会加在ref里
// 	// 如果不好加就不要了
// 	return paradigm.NewSuccessResponse(response)

// }

func (q *BlockchainTransactionQuery) GenerateResponse(data interface{}) paradigm.Response {
	tx := data.(paradigm.TransactionInfo)
	response := make(map[string]interface{})
	// todo 查看对应参数名称
	response["txHash"] = tx.TxHash
	response["contract"] = tx.Contract
	response["abi"] = tx.Abi
	response["blockHash"] = tx.BlockHash
	return paradigm.NewSuccessResponse(response)

}

func (q *BlockchainTransactionQuery) ParseRawDataFromHttpEngine(rawData map[interface{}]interface{}) bool {
	if txHash, ok := rawData["txHash"]; ok {
		q.TxHash = txHash.(string)
		return true
	}
	return false
}
func (q *BlockchainTransactionQuery) ToHttpJson() map[string]interface{} {
	return map[string]interface{}{"query": "BlockchainTransactionQuery", "txHash": q.TxHash}
}

func NewBlockchainLatestInfoQuery() *BlockchainLatestInfoQuery {
	query := new(BlockchainLatestInfoQuery)
	//query.ParseRawDataFromHttpEngine(rawData)
	//query.responseChannel = responseChannel
	query.BasicChannelQuery = paradigm.NewBasicChannelQuery()
	return query
}

func NewBlockchainBlockHashQuery(rawData map[interface{}]interface{}) *BlockchainBlockHashQuery {
	query := new(BlockchainBlockHashQuery)
	query.ParseRawDataFromHttpEngine(rawData)
	//query.responseChannel = responseChannel
	query.BlockchainQuery = NewBlockchainQuery()
	return query
}
func NewBlockchainBlockNumberQuery(rawData map[interface{}]interface{}) *BlockchainBlockNumberQuery {
	query := new(BlockchainBlockNumberQuery)
	query.ParseRawDataFromHttpEngine(rawData)
	//query.responseChannel = responseChannel
	query.BlockchainQuery = NewBlockchainQuery()
	return query
}
func NewBlockchainTransactionQuery(rawData map[interface{}]interface{}) *BlockchainTransactionQuery {
	query := new(BlockchainTransactionQuery)
	query.ParseRawDataFromHttpEngine(rawData)
	//query.responseChannel = responseChannel
	//query.BlockchainQuery = NewBlockchainQuery()
	query.BasicChannelQuery = paradigm.NewBasicChannelQuery()
	return query
}
