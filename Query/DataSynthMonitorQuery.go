package Query

import (
	"BHLayer2Node/Date"
	"BHLayer2Node/paradigm"
)

// NodesStatusQuery 数据合成页面关于节点的展示
type NodesStatusQuery struct {
	paradigm.DoubleChannelQuery
}

func (q *NodesStatusQuery) GenerateResponse(data interface{}) paradigm.Response {
	info := data.([]*paradigm.NodeStatus)
	response := make(map[string]interface{})
	nodes := make([]map[string]interface{}, 0) // 节点信息
	totalStorage, usedStorage := int32(0), int32(0)
	nbError := 0
	for _, node := range info {
		nodeInfo := make(map[string]interface{})
		nodeInfo["NodeID"] = node.NodeID
		if node.IsError() {
			nodeInfo["Status"] = "Abnormal"
			nbError++
		} else {
			nodeInfo["Status"] = "Normal"
		}
		nodeInfo["NbFinishedTasks"] = len(node.FinishedSlots)

		// 写入分任务的合成数据
		dataGroup := make(map[string]int32)
		dataGroup["ABM"] = 0
		dataGroup["BAED"] = 0
		dataGroup["FINKAN"] = 0
		for key, value := range node.SynthData {
			dataGroup[paradigm.ModelTypeToString(key)] += value
		}
		nodeInfo["SynthData"] = dataGroup

		nodeInfo["NbPendingTasks"] = len(node.PendingSlots) // 进度根据这个算
		nodeInfo["storage"] = node.DiskStorage
		nodeInfo["cpu"] = node.AverageCPUUsage
		nodeInfo["disk"] = node.DiskUsage
		nodeInfo["errorMessage"] = node.ErrorMessage()
		//合成详情就给出这个节点的合成总量，和所有完成的任务 todo 按时间有个图？
		// 节点状态，就上面的状态的信息，和pending
		nodes = append(nodes, nodeInfo)
		totalStorage += node.DiskStorage
		usedStorage += node.DiskUsage

		//这里简单判断下
		taskRate := float32(0)
		if len(node.PendingSlots)+len(node.FinishedSlots) == 0 {
			taskRate = float32(len(node.PendingSlots)) / float32(len(node.FinishedSlots)+len(node.PendingSlots))
		}
		if taskRate > 0.5 {
			nodeInfo["Workload"] = "忙碌"
		} else {
			nodeInfo["Workload"] = "空闲"
		}

	}
	response["nodes"] = nodes
	// todo
	response["statusDistribution"] = map[string]interface{}{
		"normal": len(info) - nbError,
		"down":   nbError,
		"close":  0,
	}
	response["storageDistribution"] = map[string]interface{}{
		"used":     usedStorage,
		"not used": totalStorage - usedStorage,
	}
	return paradigm.NewSuccessResponse(response)
}
func (q *NodesStatusQuery) ParseRawDataFromHttpEngine(rawData map[interface{}]interface{}) bool {
	return true
}
func (q *NodesStatusQuery) ToHttpJson() map[string]interface{} {
	return map[string]interface{}{"query": "NodesStatusQuery"}
}

func NewDataSynthMonitorQuery() *NodesStatusQuery {
	query := new(NodesStatusQuery)
	//query.ParseRawDataFromHttpEngine(rawData)
	//query.responseChannel = responseChannel
	query.DoubleChannelQuery = paradigm.NewDoubleChannelQuery()
	return query
}

// DateSynthDataQuery 数据合成页面关于日合成数据的展示
type DateSynthDataQuery struct {
	paradigm.BasicChannelQuery
}

func (q *DateSynthDataQuery) GenerateResponse(data interface{}) paradigm.Response {
	// 传入的数据是dateRecords
	records := data.([]*Date.DateRecord)
	response := make(map[string]interface{})
	dates := make([]string, 0)               // 按序存储时间，便于前端排序,go的map无序
	synthData := make([]map[string]int32, 0) // 合成数据
	initTasks := make([]int32, 0)            // 新建任务
	finishTasks := make([]int32, 0)          // 完成任务
	totalTasks := int32(0)
	totalFinish := int32(0)
	datasetDistribution := make(map[string]int32)
	for _, record := range records {
		dates = append(dates, paradigm.DateFormat(record.Date()))
		dataGroup := make(map[string]int32)
		dataGroup["ABM"] = 0
		dataGroup["BAED"] = 0
		dataGroup["FINKAN"] = 0
		for key, value := range record.SynthData {
			dataGroup[paradigm.ModelTypeToString(key)] += value
		}
		synthData = append(synthData, dataGroup)
		initTasks = append(initTasks, record.NbInitTasks)
		finishTasks = append(finishTasks, record.NbFinishTasks)
		totalTasks += record.NbInitTasks
		totalFinish += record.NbFinishTasks
		for dataset, n := range record.DatasetDistribution {
			if _, exist := datasetDistribution[dataset]; !exist {
				datasetDistribution[dataset] = 0
			}
			datasetDistribution[dataset] += n
		}
	}
	response["date"] = dates
	response["init"] = initTasks
	response["finish"] = finishTasks
	response["synthData"] = synthData
	response["taskDistribution"] = map[string]interface{}{
		"processing": totalTasks - totalFinish,
		"finish":     totalFinish,
	}
	response["datasetDistribution"] = datasetDistribution
	return paradigm.NewSuccessResponse(response)

}

func (q *DateSynthDataQuery) ParseRawDataFromHttpEngine(rawData map[interface{}]interface{}) bool {
	return true
}
func (q *DateSynthDataQuery) ToHttpJson() map[string]interface{} {
	return map[string]interface{}{"query": "DateSynthDataQuery"}
}
func NewDateSynthDataQuery() *DateSynthDataQuery {
	query := new(DateSynthDataQuery)
	//query.ParseRawDataFromHttpEngine(rawData)
	//query.responseChannel = responseChannel
	query.BasicChannelQuery = paradigm.NewBasicChannelQuery()
	return query
}
