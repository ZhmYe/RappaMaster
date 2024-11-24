package paradigm

// BHLayer2NodeSchedule 描述任务分配到的节点
type BHLayer2NodeSchedule struct {
	//Sign   string
	//Slot   int
	Size int
	NID  int
	//Model  string
	//Params map[string]interface{}
}

// HttpTaskRequest 格式化前端发来的请求
type HttpTaskRequest struct {
	Sign string // task sign
	//Slot   int                    // slot index
	Size   int                    // data size
	Model  string                 // 模型名称
	Params map[string]interface{} // 不确定的模型参数
}
type PendingSlotItem struct {
	Sign     string                 // task sign
	Slot     int                    // slot index
	Size     int                    // data size
	Model    string                 // 模型名称
	Params   map[string]interface{} // 不确定的模型参数
	Schedule []BHLayer2NodeSchedule // 调度
}

func (s *PendingSlotItem) UpdateSchedule(schedule []BHLayer2NodeSchedule) {
	s.Schedule = schedule
}

type SlotRecord struct {
	Size   int // 处理了多少数据
	Sign   string
	Slot   int
	Active []int // 活跃的节点数
	Miss   []int // 未完成的节点数 todo 这个东西很难定义，要和heartbeat一起考虑
}

func NewSlotRecord() SlotRecord {
	return SlotRecord{}
}
