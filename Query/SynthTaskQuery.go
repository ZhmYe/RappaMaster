package Query

import "BHLayer2Node/paradigm"

// SynthTaskQuery 合成任务界面关于所有task的查询
type SynthTaskQuery struct {
	paradigm.BasicChannelQuery
}

func (q *SynthTaskQuery) GenerateResponse(data interface{}) paradigm.Response {
	info := data.(map[string]*paradigm.Task)
	response := make(map[string]interface{})
	tasks := make([]map[string]interface{}, 0)
	for _, task := range info {
		taskInfo := make(map[string]interface{})
		taskInfo["taskID"] = task.Sign
		taskInfo["txHash"] = task.TxReceipt.TransactionHash
		taskInfo["total"] = task.Size        // 数据总量
		taskInfo["process"] = task.Process   // 已合成
		taskInfo["status"] = task.IsFinish() // 是否完成
		taskInfo["startTime"] = paradigm.TimeFormat(task.StartTime)
		if task.IsFinish() {
			taskInfo["endTime"] = paradigm.TimeFormat(task.EndTime)
		} else {
			taskInfo["endTime"] = ""
		}
		tasks = append(tasks, taskInfo)
	}
	response["tasks"] = tasks

	return paradigm.NewSuccessResponse(response)
}
func (q *SynthTaskQuery) ParseRawDataFromHttpEngine(rawData map[interface{}]interface{}) bool {
	return true
}
func (q *SynthTaskQuery) ToHttpJson() map[string]interface{} {
	return map[string]interface{}{"query": "SynthTaskQuery"}
}

func NewSynthTaskQuery() *SynthTaskQuery {
	query := new(SynthTaskQuery)
	//query.ParseRawDataFromHttpEngine(rawData)
	//query.responseChannel = responseChannel
	query.BasicChannelQuery = paradigm.NewBasicChannelQuery()
	return query
}
