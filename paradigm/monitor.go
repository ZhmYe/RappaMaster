package paradigm

import "BHLayer2Node/Config"

type NodeStatus struct {
	NodeID          int32
	Address         Config.BHNodeAddress
	AverageCPUUsage int               // 平均cpu使用率,%
	DiskUsage       int32             // 存储状况，B为单位 todo 这里要考虑是否会溢出，以及是否用小数比较好
	FinishedSlots   []SlotHash        // 已经完成的任务， todo 这里其实要考虑节点虚报，对于真正的完整的monitor来说应该是需要根据oracle来统计的
	PendingSlots    map[SlotHash]bool // 处理中的任务 todo 同上
	SynthData       int32             // 合成数据总量 todo 同上
	DiskStorage     int32             // 磁盘总量，B为单位
	Rate            float64           // 评分，用于Monitor Advice，暂时先不写
}

func (s *NodeStatus) UpdateUsage(cpu int, disk int32, total int32) {
	if cpu >= 0 {
		s.AverageCPUUsage = cpu
	}
	if disk >= 0 {
		s.DiskUsage = disk
	}
	if total >= 0 {
		s.DiskStorage = total
	}
}
func (s *NodeStatus) UpdatePendingSlot(slotHash string) {
	s.PendingSlots[slotHash] = true
}
func (s *NodeStatus) UpdateFinishSlot(slotHash string, process int32) {
	delete(s.PendingSlots, slotHash)
	s.FinishedSlots = append(s.FinishedSlots, slotHash)
	s.SynthData += process
}
func NewNodeStatus(nodeID int32, address Config.BHNodeAddress) *NodeStatus {
	return &NodeStatus{
		NodeID:          nodeID,
		Address:         address,
		AverageCPUUsage: 0,
		DiskUsage:       0,
		FinishedSlots:   make([]SlotHash, 0),
		PendingSlots:    make(map[SlotHash]bool),
		SynthData:       0,
		DiskStorage:     0,
		Rate:            0,
	}
}

// AdviceRequest Scheduler向Monitor请求调度方式
// TODO 这里细节有待思考
type AdviceRequest struct {
	Size     int32 // 全量的大小
	response chan AdviceResponse
}

func (r *AdviceRequest) SendResponse(resp AdviceResponse) {
	r.response <- resp
}
func (r *AdviceRequest) ReceiveResponse() AdviceResponse {
	return <-r.response
}
func NewAdviceRequest(size int32) *AdviceRequest {
	return &AdviceRequest{
		Size:     size,
		response: make(chan AdviceResponse, 1),
	}
}

// AdviceResponse 调度结果，TODO
type AdviceResponse struct {
	NodeIDs      []int32
	ScheduleSize []int32
}

func NewAdviceResponse(nodeIDs []int32, size []int32) *AdviceResponse {
	return &AdviceResponse{
		NodeIDs:      nodeIDs,
		ScheduleSize: size,
	}
}
