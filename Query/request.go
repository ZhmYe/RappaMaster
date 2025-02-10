package Query

import (
	"BHLayer2Node/paradigm"
	"github.com/gin-gonic/gin"
	"strconv"
)

type HttpInitTaskRequest struct {
	Sign string // task sign
	//Slot       int32                  // slot index
	Size       int32                  // data size
	Model      string                 // 模型名称
	Params     map[string]interface{} // 不确定的模型参数
	IsReliable bool                   // 是否需要可信证明
}

type HttpOracleQueryRequest struct {
	Query string                      // 查询类型
	Data  map[interface{}]interface{} // 查询内容rawData
}

func (r *HttpOracleQueryRequest) BuildQueryFromGETRequest(c *gin.Context) (bool, paradigm.Query) {
	query := c.DefaultQuery("query", "")
	if query == "" {
		//fmt.Println(222)
		return false, nil
	}
	//fmt.Println(111, query)
	r.Query = query
	switch query {
	case "EvidencePreserveTaskTxQuery":
		txHash := c.DefaultQuery("txHash", "")
		if txHash == "" {
			return false, nil
		}
		return true, NewEvidencePreserveTaskTxQuery(map[interface{}]interface{}{
			"txHash": txHash,
		})
	case "EvidencePreserveTaskIDQuery":
		taskID := c.DefaultQuery("taskID", "")
		if taskID == "" {
			return false, nil
		}
		//t, err := strconv.Atoi(taskID)
		//if err != nil {
		//	return false, nil
		//}
		return true, NewEvidencePreserveTaskIDQuery(map[interface{}]interface{}{
			"taskID": taskID,
		})
	case "EvidencePreserveEpochTxQuery":
		txHash := c.DefaultQuery("txHash", "")
		if txHash == "" {
			return false, nil
		}
		return true, NewEvidencePreserveEpochTxQuery(map[interface{}]interface{}{
			"txHash": txHash,
		})
	case "EvidencePreserveEpochIDQuery":
		epochID := c.DefaultQuery("epochID", "")
		//fmt.Println(epochID)
		if epochID == "" {
			return false, nil
		}
		e, err := strconv.Atoi(epochID)
		if err != nil {
			return false, nil
		}
		return true, NewEvidencePreserveEpochIDQuery(map[interface{}]interface{}{
			"epochID": e,
		})
	case "BlockchainLatestInfoQuery":
		return true, NewBlockchainLatestInfoQuery()
	case "BlockchainBlockHashQuery":
		blockHash := c.DefaultQuery("blockHash", "")
		if blockHash == "" {
			return false, nil
		}
		return true, NewBlockchainBlockHashQuery(map[interface{}]interface{}{
			"blockHash": blockHash,
		})
	case "BlockchainBlockNumberQuery":
		blockNumber := c.DefaultQuery("blockNumber", "")
		if blockNumber == "" {
			return false, nil
		}
		b, err := strconv.Atoi(blockNumber)
		if err != nil {
			return false, nil
		}
		return true, NewBlockchainBlockNumberQuery(map[interface{}]interface{}{
			"blockNumber": b,
		})
	case "BlockchainTransactionQuery":
		txHash := c.DefaultQuery("txHash", "")
		if txHash == "" {
			return false, nil
		}
		return true, NewBlockchainTransactionQuery(map[interface{}]interface{}{
			"txHash": txHash,
		})
	case "DataSynthMonitorQuery":
		return true, NewDataSynthMonitorQuery()

	default:
		return false, nil
	}
}
