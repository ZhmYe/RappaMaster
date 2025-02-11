package Monitor

import (
	"BHLayer2Node/LogWriter"
	"BHLayer2Node/Query"
	"BHLayer2Node/paradigm"
	"fmt"
	"strconv"
)

// Monitor 监视节点状态
// 存储节点的状态信息，并维护一些统计值
type Monitor struct {
	//config  Config.BHLayer2NodeConfig
	channel    *paradigm.RappaChannel
	nodeStatus []*paradigm.NodeStatus
}

// processHeartbeatResponse 处理节点的心跳回复，其中包含节点最新的磁盘、cpu等信息用于展示
func (m *Monitor) processHeartbeatResponse() {
	for heartbeat := range m.channel.MonitorHeartbeatChannel {
		// TODO 这里一旦发现heartbeat内容就continue，应该在coordinator里就检测
		// 检测内容：1. 存在Map的key; 2. total >= disk
		status := heartbeat.NodeStatus
		if _, exist := status["cpu"]; !exist {
			paradigm.RaiseError(paradigm.ValueError, "error status key", false)
			continue
		}
		if _, exist := status["disk"]; !exist {
			paradigm.RaiseError(paradigm.ValueError, "error status key", false)
			continue
		}
		if _, exist := status["total"]; !exist {
			paradigm.RaiseError(paradigm.ValueError, "error status key", false)
			continue
		}
		c, d, t := status["cpu"], status["disk"], status["total"]
		cpuUsage, ok := strconv.Atoi(c)
		if ok != nil {
			paradigm.RaiseError(paradigm.ValueError, "error status value", false)
			continue
		}
		diskUsage, ok := strconv.Atoi(d)
		if ok != nil {
			paradigm.RaiseError(paradigm.ValueError, "error status value", false)
			continue
		}
		diskStorage, ok := strconv.Atoi(t)
		if ok != nil {
			paradigm.RaiseError(paradigm.ValueError, "error status value", false)
			continue
		}
		LogWriter.Log("INFO", fmt.Sprintf("Monitor Update Node %d Status, CPU Usage: %d %%, Disk Usage: %d, Total Disk Space: %d", heartbeat.NodeId, cpuUsage, diskStorage, diskStorage))

		m.nodeStatus[int(heartbeat.NodeId)].UpdateUsage(cpuUsage, int32(diskUsage), int32(diskStorage))
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
				LogWriter.Log("INFO", fmt.Sprintf("Monitor Update Node %d Status, New Pending Slot: %s", nodeID, slot.SlotID))

			}
		case *paradigm.TaskProcessTransaction:
			tx := info.(*paradigm.TaskProcessTransaction)
			nodeID := tx.Nid
			m.nodeStatus[int(nodeID)].UpdateFinishSlot(tx.SlotHash(), tx.Process)
			LogWriter.Log("INFO", fmt.Sprintf("Monitor Update Node %d Status, New Finish Slot: %s, process: %d", nodeID, tx.SlotHash(), tx.Process))

		default:
			paradigm.RaiseError(paradigm.RuntimeError, "Error type in oracle channel", false)
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
			paradigm.RaiseError(paradigm.RuntimeError, "Unsupported Query Type In Monitor", false)

		}
	}
}

// advice
func (m *Monitor) advice(request *paradigm.AdviceRequest) {
	// TODO 调度方式，目前就写成所有节点，均分
	nodeIDs := make([]int32, len(m.nodeStatus))
	scheduleSize := make([]int32, len(m.nodeStatus))
	for i := 0; i < len(m.nodeStatus); i++ {
		nodeIDs[i] = int32(i)
		adviceSize := request.Size / int32(len(nodeIDs))
		// 慢点来 后续考虑维护一个全局的均值
		if adviceSize > 10 {
			adviceSize = 10
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
