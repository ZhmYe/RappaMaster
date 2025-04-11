package Monitor

import (
	"BHLayer2Node/Query"
	"BHLayer2Node/paradigm"
)

// Monitor 监视节点状态
// 存储节点的状态信息，并维护一些统计值
type Monitor struct {
	channel    *paradigm.RappaChannel
	nodeStatus []*paradigm.NodeStatus
}

// processHeartbeatResponse 处理节点的心跳回复，其中包含节点最新的磁盘、cpu等信息用于展示
func (m *Monitor) processHeartbeatResponse() {
	for report := range m.channel.MonitorHeartbeatChannel {
		if report.IsError {
			// 发现是错误的，那么对NodeStatus更新错误
			m.nodeStatus[int(report.NodeID)].SetError(report.ErrMessage)
		} else {
			m.nodeStatus[int(report.NodeID)].UpdateUsage(report.CPUUsage, report.DiskUsage, report.DiskStorage)
		}
	}
}

// processOracleInfo 处理来自Oracle的更新信息，用于更新nodeStatus中的任务和调度相关的内容
func (m *Monitor) processOracleInfo() {
	for info := range m.channel.MonitorOracleChannel {
		switch info.(type) {
		// 传过来的内容要么是schedule，要么是taskprocessTransaction
		case *paradigm.SynthTaskSchedule:
			schedule := info.(*paradigm.SynthTaskSchedule)
			for nodeID, index := range schedule.NodeIDMap {
				slot := schedule.Slots[index]
				// 调度时就失败的就不用更新了
				if slot.Status == paradigm.Failed {
					continue
				}
				m.nodeStatus[nodeID].UpdatePendingSlot(slot.SlotID)
				//paradigm.Log("INFO", fmt.Sprintf("Monitor Update Node %d Status, New Pending Slot: %s", nodeID, slot.SlotID))

			}
		case *paradigm.TaskProcessTransaction:
			tx := info.(*paradigm.TaskProcessTransaction)
			nodeID := tx.Nid
			m.nodeStatus[int(nodeID)].UpdateFinishSlot(tx.SlotHash(), tx.Process, tx.Model)
			//paradigm.Log("INFO", fmt.Sprintf("Monitor Update Node %d Status, New Finish Slot: %s, process: %d", nodeID, tx.SlotHash(), tx.Process))

		default:
			paradigm.Error(paradigm.RuntimeError, "Error type in oracle channel")
		}
	}
}

// processAdviceRequest 处理来自Scheduler的请求
func (m *Monitor) processAdviceRequest() {
	for request := range m.channel.MonitorAdviceChannel {
		m.advice(request)
	}

}

// processQuery 处理查询，这里针对monitor的查询就是所有的节点
func (m *Monitor) processQuery() {
	for query := range m.channel.MonitorQueryChannel {
		// 就是全部返回
		switch query.(type) {
		case *Query.NodesStatusQuery:
			item := query.(*Query.NodesStatusQuery)
			item.SendInfo(m.nodeStatus)
		default:
			paradigm.Error(paradigm.RuntimeError, "Unsupported Query Type In Monitor")

		}
	}
}

// advice
func (m *Monitor) advice(request *paradigm.AdviceRequest) {
	nodeIDs := make([]int32, len(m.nodeStatus))
	scheduleSize := make([]int32, len(m.nodeStatus))
	for i := 0; i < len(m.nodeStatus); i++ {
		nodeIDs[i] = int32(i)
		adviceSize := request.Size / int32(len(nodeIDs))
		// 慢点来 后续考虑维护一个全局的均值
		if adviceSize > 3000 {
			adviceSize = 3000
		}
		if adviceSize == 0 {
			adviceSize = 1
		}
		scheduleSize[i] = adviceSize
	}

	scheduleSize[0] += request.Size % int32(len(nodeIDs))
	response := paradigm.NewAdviceResponse(nodeIDs, scheduleSize)
	request.SendResponse(*response)
}
func (m *Monitor) Start() {
	go m.processHeartbeatResponse()
	go m.processOracleInfo()
	go m.processAdviceRequest()
	go m.processQuery()
}

func NewMonitor(channel *paradigm.RappaChannel) *Monitor {
	nodeStatus := make([]*paradigm.NodeStatus, len(channel.Config.BHNodeAddressMap)) // 这里假设key是对应的[n]
	for nodeID, address := range channel.Config.BHNodeAddressMap {
		nodeStatus[nodeID] = paradigm.NewNodeStatus(int32(nodeID), *address)
	}
	return &Monitor{
		channel:    channel,
		nodeStatus: nodeStatus,
	}

}
