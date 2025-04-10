package Query

import (
	"BHLayer2Node/paradigm"
	"fmt"
	"time"
)

// CollectTaskQuery 合成任务界面下载数据
type CollectTaskQuery struct {
	request paradigm.HttpCollectRequest
	paradigm.BasicChannelQuery
}

func (q *CollectTaskQuery) TaskID() paradigm.TaskHash {
	return q.request.Sign
}

func (q *CollectTaskQuery) GenerateResponse(data interface{}) paradigm.Response {
	collector := data.(paradigm.RappaCollector)
	output, err := collector.ProcessCollect(q.request)
	if err != nil {
		return paradigm.NewErrorResponse(paradigm.NewRappaError(paradigm.ChunkRecoverError, err.Error()))
	}
	if output == nil {
		return paradigm.NewErrorResponse(paradigm.NewRappaError(paradigm.ChunkRecoverError, "Recover Output is nil"))
	}
	fileByte, fileType, err := paradigm.DataToFile(output)
	if err != nil {
		return paradigm.NewErrorResponse(paradigm.NewRappaError(paradigm.ChunkRecoverError, err.Error()))
	}
	//fmt.Println(fileByte)
	result := make(map[string]interface{})
	generateFileName := func() string {
		return fmt.Sprintf("%s_%d_%s.%s", q.request.Sign, q.request.Size, time.Now().Format("2006-01-02_15-04-05"), fileType)
	}
	result["filename"] = generateFileName()
	result["file"] = fileByte
	return paradigm.NewSuccessResponse(result)

}
func (q *CollectTaskQuery) ParseRawDataFromHttpEngine(rawData map[interface{}]interface{}) bool {
	r := paradigm.HttpCollectRequest{
		Sign: "",
		Size: 0,
		//TransferChannel: nil,
	}
	if s, ok := rawData["taskID"].(string); ok {
		r.Sign = s
	} else {
		return false
	}
	if size, ok := rawData["size"].(int); ok {
		r.Size = int32(size)
	} else {
		return false
	}
	q.request = r
	return true
}
func (q *CollectTaskQuery) ToHttpJson() map[string]interface{} {
	return map[string]interface{}{"query": "CollectTaskQuery", "taskID": q.request.Sign, "size": q.request.Size}
}

// SynthTaskQuery 合成任务界面关于所有task的查询
type SynthTaskQuery struct {
	paradigm.BasicChannelQuery
}

func (q *SynthTaskQuery) GenerateResponse(data interface{}) paradigm.Response {
	info := data.([]*paradigm.Task)
	response := make(map[string]interface{})
	tasks := make([]map[string]interface{}, 0, len(info))
	for _, task := range info {
		taskInfo := make(map[string]interface{})
		taskInfo["taskID"] = task.Sign
		taskInfo["taskName"] = task.Name
		taskInfo["txHash"] = task.TxReceipt.TransactionHash
		taskInfo["total"] = task.Size // 数据总量
		//taskInfo["process"] = min(task.Process, task.Size) // 已合成
		taskInfo["process"] = task.Process
		taskInfo["status"] = task.Status
		taskInfo["model"] = paradigm.ModelTypeToString(task.Model)
		taskInfo["startTime"] = paradigm.TimeFormat(task.StartTime)
		if task.Status == paradigm.Finished {
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

func NewCollectTaskQuery(rawData map[interface{}]interface{}) *CollectTaskQuery {
	query := new(CollectTaskQuery)
	query.ParseRawDataFromHttpEngine(rawData)
	query.BasicChannelQuery = paradigm.NewBasicChannelQuery()
	return query
}
func NewSynthTaskQuery() *SynthTaskQuery {
	query := new(SynthTaskQuery)
	//query.ParseRawDataFromHttpEngine(rawData)
	//query.responseChannel = responseChannel
	query.BasicChannelQuery = paradigm.NewBasicChannelQuery()
	return query
}
