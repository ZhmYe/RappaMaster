package Query

type HttpInitTaskRequest struct {
	Name  string // task name, defined by user
	Size  int64  // data size
	Model string // model
	//Params map[string]interface{} // model params
}

//
//type HttpOracleQueryRequest struct {
//	Query string                      // 查询类型
//	Data  map[interface{}]interface{} // 查询内容rawData
//}
//
//func (r *HttpOracleQueryRequest) BuildQueryFromGETRequest(c *gin.Context) (bool, paradigm.Query) {
//	query := c.DefaultQuery("query", "")
//	if query == "" {
//		//fmt.Println(222)
//		return false, nil
//	}
//	r.Query = query
//	switch query {
//	case "EvidencePreserveTaskTxQuery":
//		txHash := c.DefaultQuery("txHash", "")
//		if txHash == "" {
//			return false, nil
//		}
//		return true, NewEvidencePreserveTaskTxQuery(map[interface{}]interface{}{
//			"txHash": txHash,
//		})
//	case "EvidencePreserveTaskIDQuery":
//		taskID := c.DefaultQuery("taskID", "")
//		if taskID == "" {
//			return false, nil
//		}
//		//t, err := strconv.Atoi(taskID)
//		//if err != nil {
//		//	return false, nil
//		//}
//		return true, NewEvidencePreserveTaskIDQuery(map[interface{}]interface{}{
//			"taskID": taskID,
//		})
//	case "EvidencePreserveEpochTxQuery":
//		txHash := c.DefaultQuery("txHash", "")
//		if txHash == "" {
//			return false, nil
//		}
//		return true, NewEvidencePreserveEpochTxQuery(map[interface{}]interface{}{
//			"txHash": txHash,
//		})
//	case "EvidencePreserveEpochIDQuery":
//		epochID := c.DefaultQuery("epochID", "")
//		//fmt.Println(epochID)
//		if epochID == "" {
//			return false, nil
//		}
//		e, err := strconv.Atoi(epochID)
//		if err != nil {
//			return false, nil
//		}
//		return true, NewEvidencePreserveEpochIDQuery(map[interface{}]interface{}{
//			"epochID": e,
//		})
//	case "BlockchainLatestInfoQuery":
//		return true, NewBlockchainLatestInfoQuery()
//	case "BlockchainBlockHashQuery":
//		blockHash := c.DefaultQuery("blockHash", "")
//		if blockHash == "" {
//			return false, nil
//		}
//		return true, NewBlockchainBlockHashQuery(map[interface{}]interface{}{
//			"blockHash": blockHash,
//		})
//	case "BlockchainBlockNumberQuery":
//		blockNumber := c.DefaultQuery("blockNumber", "")
//		if blockNumber == "" {
//			return false, nil
//		}
//		b, err := strconv.Atoi(blockNumber)
//		if err != nil {
//			return false, nil
//		}
//		return true, NewBlockchainBlockNumberQuery(map[interface{}]interface{}{
//			"blockNumber": b,
//		})
//	case "BlockchainTransactionQuery":
//		txHash := c.DefaultQuery("txHash", "")
//		if txHash == "" {
//			return false, nil
//		}
//		return true, NewBlockchainTransactionQuery(map[interface{}]interface{}{
//			"txHash": txHash,
//		})
//	case "NodesStatusQuery":
//		return true, NewDataSynthMonitorQuery()
//	case "DateSynthDataQuery":
//		return true, NewDateSynthDataQuery()
//	case "DateTransactionQuery":
//		return true, NewDateTransactionQuery()
//	case "SynthTaskQuery":
//		return true, NewSynthTaskQuery()
//	case "CollectTaskQuery":
//		taskID := c.DefaultQuery("taskID", "")
//		if taskID == "" {
//			return false, nil
//		}
//		size := c.DefaultQuery("size", "")
//		if size == "" {
//			return false, nil
//		}
//		s, err := strconv.Atoi(size)
//		if err != nil {
//			return false, nil
//		}
//		return true, NewCollectTaskQuery(map[interface{}]interface{}{
//			"taskID": taskID,
//			"size":   s,
//		})
//	default:
//		return false, nil
//	}
//}
