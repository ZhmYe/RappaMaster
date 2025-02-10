package Query

import (
	"BHLayer2Node/paradigm"
)

type DataSynthMonitorQuery struct {
	paradigm.DoubleChannelQuery
}

func (q *DataSynthMonitorQuery) GenerateResponse(data interface{}) paradigm.Response {
	info := data.([]*paradigm.NodeStatus)
	response := make(map[string]interface{})
	nodes := make([]map[string]interface{}, 0) // 节点信息
	totalStorage, usedStorage := int32(0), int32(0)
	for _, node := range info {
		nodeInfo := make(map[string]interface{})
		nodeInfo["NodeID"] = node.NodeID
		nodeInfo["Status"] = "Normal" // todo
		nodeInfo["Workload"] = "空闲"   // todo
		nodeInfo["NbFinishedTasks"] = len(node.FinishedSlots)
		nodeInfo["SynthData"] = node.SynthData
		nodeInfo["NbPendingTasks"] = len(node.PendingSlots) // 进度根据这个算
		nodeInfo["storage"] = node.DiskStorage
		nodeInfo["cpu"] = node.AverageCPUUsage
		nodeInfo["disk"] = node.DiskUsage
		//合成详情就给出这个节点的合成总量，和所有完成的任务 todo 按时间有个图？
		// 节点状态，就上面的状态的信息，和pending
		nodes = append(nodes, nodeInfo)
		totalStorage += node.DiskStorage
		usedStorage += node.DiskUsage
	}
	response["nodes"] = nodes
	// todo
	response["statusDistribution"] = map[string]interface{}{
		"normal": len(info),
		"down":   0,
		"close":  0,
	}
	response["storageDistribution"] = map[string]interface{}{
		"used":     usedStorage,
		"not used": totalStorage - usedStorage,
	}
	return paradigm.NewSuccessResponse(response)
}
func (q *DataSynthMonitorQuery) ParseRawDataFromHttpEngine(rawData map[interface{}]interface{}) bool {
	return true
}
func (q *DataSynthMonitorQuery) ToHttpJson() map[string]interface{} {
	return map[string]interface{}{"query": "DataSynthMonitorQuery"}
}

func NewDataSynthMonitorQuery() *DataSynthMonitorQuery {
	query := new(DataSynthMonitorQuery)
	//query.ParseRawDataFromHttpEngine(rawData)
	//query.responseChannel = responseChannel
	query.DoubleChannelQuery = paradigm.NewDoubleChannelQuery()
	return query
}
