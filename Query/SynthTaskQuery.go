package Query

import (
	"BHLayer2Node/paradigm"
	"BHLayer2Node/utils"
	"fmt"
	"strconv"
	"time"
)

// CollectTaskQuery 合成任务界面下载数据
type CollectTaskQuery struct {
	request paradigm.HttpCollectRequest
	paradigm.BasicChannelQuery
}

type UploadTaskQuery struct {
	request HttpUploadTaskRequest
	paradigm.BasicChannelQuery
}

func (q *UploadTaskQuery) TaskID() paradigm.TaskHash {
	return q.request.TaskID
}

func (q *CollectTaskQuery) TaskID() paradigm.TaskHash {
	return q.request.Sign
}

func (q *UploadTaskQuery) GenerateResponse(data interface{}) paradigm.Response {
	task := data.(paradigm.Task)
	collector := task.GetCollector()
	request := paradigm.HttpCollectRequest{
		Sign: task.Sign,
		Size: task.Size,
	}
	// 先创建收集任务
	output, err := collector.ProcessCollect(request)
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
		return fmt.Sprintf("%s_%d_%s.%s", request.Sign, request.Size, time.Now().Format("2006-01-02_15-04-05"), fileType)
	}
	result["fileupload"] = true

	datasetType := ""

	switch task.Model {
	case paradigm.FINKAN:
		datasetType = "表格数据"
	case paradigm.ABM:
		datasetType = "时序数据"
	case paradigm.BAED:
		datasetType = "图数据"
	default:
		break
	}
	//发送上传请求
	err = utils.UploadFile("http://oneplatbank.i2soft.cn:19195/jeecaboot/datap/uploadFile", map[string]string{
		"dataType":    "用户画像",
		"datasetCode": fmt.Sprintf("%d", task.Model),
		"datasetName": task.GetDataset(),
		"datasetType": datasetType,
		"dataCnt":     strconv.Itoa(int(task.Size)),
		"purpose":     task.Name,
		"description": "说明:测试数据",
		"createBy":    "User02",
	}, fileByte, generateFileName(), "file")
	if err != nil {
		return paradigm.NewErrorResponse(paradigm.NewRappaError(paradigm.NetworkError, err.Error()))
	}
	return paradigm.NewSuccessResponse(result)
}

func (q *UploadTaskQuery) ToHttpJson() map[string]interface{} {
	return map[string]interface{}{
		"query": "UploadTaskQuery", // 建议用方法获取类型名，比如 "UploadTaskQuery"
		"request": HttpUploadTaskRequest{
			TaskID:      q.request.TaskID, // 假设 UploadTaskQuery 中有 taskID 字段
			Purpose:     q.request.Purpose,
			Description: q.request.Description,
			CreateBy:    q.request.CreateBy,
		},
	}
}

func (q *UploadTaskQuery) ParseRawDataFromHttpEngine(rawData map[interface{}]interface{}) bool {
	r := HttpUploadTaskRequest{
		TaskID:      "",
		Purpose:     "",
		Description: "",
		CreateBy:    "",
	}

	if taskID, ok := rawData["taskID"].(string); ok {
		r.TaskID = taskID
	} else {
		return false
	}

	if purpose, ok := rawData["purpose"].(string); ok {
		r.Purpose = purpose
	} else {
		return false
	}

	if description, ok := rawData["description"].(string); ok {
		r.Description = description
	} else {
		return false
	}

	if createBy, ok := rawData["createBy"].(string); ok {
		r.CreateBy = createBy
	} else {
		return false
	}

	// 如果都成功，则赋值给 q.request
	q.request = r
	return true
}

// 上传任务
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

func (q *TaskOnNodesQuery) ParseRawDataFromHttpEngine(rawData map[interface{}]interface{}) bool {
	if s, ok := rawData["taskID"].(string); ok {
		q.Sign = s
	} else {
		return false
	}
	return true
}

func (q *TaskOnNodesQuery) ToHttpJson() map[string]interface{} {
	return map[string]interface{}{"query": "TaskOnNodesQuery", "taskId": q.Sign}
}

// TaskOnNodesQuery 查询task在不同节点上的并行合成数
type TaskOnNodesQuery struct {
	Sign string
	paradigm.BasicChannelQuery
}

func (q *TaskOnNodesQuery) GenerateResponse(data interface{}) paradigm.Response {
	slots := data.([]*paradigm.Slot)
	nodeInfo := make(map[int32]int32)
	response := make(map[string]interface{})
	for _, slot := range slots {
		if data, exist := nodeInfo[slot.NodeID]; exist {
			nodeInfo[slot.NodeID] = data + slot.ScheduleSize
		} else {
			nodeInfo[slot.NodeID] = slot.ScheduleSize
		}
	}
	response["nodeInfo"] = nodeInfo
	return paradigm.NewSuccessResponse(response)
}
func (q *SynthTaskQuery) ParseRawDataFromHttpEngine(rawData map[interface{}]interface{}) bool {
	return true
}
func (q *SynthTaskQuery) ToHttpJson() map[string]interface{} {
	return map[string]interface{}{"query": "SynthTaskQuery"}
}

// SlotIntegrityVerification 对slot进行完整性验证, 目前使用Merkle Tree结构
type SlotIntegrityVerification struct {
	SlotHash string
	paradigm.BasicChannelQuery
}

func (q *SlotIntegrityVerification) GenerateResponse(data interface{}) paradigm.Response {
	slots := data.([]*paradigm.Slot)
	var leaves [][]byte
	indexOfTarget := -1

	for _, slot := range slots {
		if slot.CommitSlot == nil || len(slot.CommitSlot.Commitment) == 0 {
			continue
		}
		leaves = append(leaves, slot.CommitSlot.Commitment)
		if string(slot.SlotID) == q.SlotHash {
			indexOfTarget = len(leaves) - 1
		}
	}

	if indexOfTarget == -1 {
		return paradigm.NewErrorResponse(
			paradigm.NewRappaError(paradigm.SlotLifeError, "target slot not found or no valid commitments"))
	}

	tree, root := utils.BuildMerkleTree(leaves)
	// 提取每层用于构建验证的两个节点（当前节点 + 兄弟节点）
	proofPairs := []map[string]interface{}{}
	index := indexOfTarget
	for level := 0; level < len(tree)-1; level++ {
		currentLevel := tree[level]
		siblingIndex := index ^ 1
		if siblingIndex >= len(currentLevel) {
			// 没有兄弟节点，不构建该层 proofPair
			index /= 2
			continue
		}

		currentNode := currentLevel[index]
		siblingNode := currentLevel[siblingIndex]

		proofPairs = append(proofPairs, map[string]interface{}{
			"level": level,
			"current": map[string]interface{}{
				"index": index,
				"hash":  fmt.Sprintf("0x%x", currentNode),
			},
			"sibling": map[string]interface{}{
				"index": siblingIndex,
				"hash":  fmt.Sprintf("0x%x", siblingNode),
			},
		})

		index /= 2
	}
	// 反转 proofPairs，使其自顶向下排序
	for i, j := 0, len(proofPairs)-1; i < j; i, j = i+1, j-1 {
		proofPairs[i], proofPairs[j] = proofPairs[j], proofPairs[i]
	}

	leafHex := fmt.Sprintf("0x%x", leaves[indexOfTarget])
	rootHex := fmt.Sprintf("0x%x", root)

	response := map[string]interface{}{
		"slotHash":    q.SlotHash,
		"leaf":        leafHex,
		"merkleRoot":  rootHex,
		"proof":       proofPairs,
		"verified":    true,
		"leavesCount": len(leaves),
		"targetIndex": indexOfTarget,
	}

	return paradigm.NewSuccessResponse(response)
}
func (q *SlotIntegrityVerification) ParseRawDataFromHttpEngine(rawData map[interface{}]interface{}) bool {
	if s, ok := rawData["slotHash"].(string); ok {
		q.SlotHash = s
	} else {
		return false
	}
	return true
}
func (q *SlotIntegrityVerification) ToHttpJson() map[string]interface{} {
	return map[string]interface{}{"query": "SlotIntegrityVerification", "slotHash": q.SlotHash}
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
func NewTaskOnNodesQuery(rawData map[interface{}]interface{}) *TaskOnNodesQuery {
	query := new(TaskOnNodesQuery)
	query.ParseRawDataFromHttpEngine(rawData)
	query.BasicChannelQuery = paradigm.NewBasicChannelQuery()
	return query
}
func NewSlotIntegrityVerification(rawData map[interface{}]interface{}) *SlotIntegrityVerification {
	query := new(SlotIntegrityVerification)
	query.ParseRawDataFromHttpEngine(rawData)
	query.BasicChannelQuery = paradigm.NewBasicChannelQuery()
	return query
}

func NewUploadTaskQuery(rawData map[interface{}]interface{}) *UploadTaskQuery {
	query := new(UploadTaskQuery)
	query.ParseRawDataFromHttpEngine(rawData)
	query.BasicChannelQuery = paradigm.NewBasicChannelQuery()
	return query
}
