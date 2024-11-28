package paradigm

// TaskSchedule 描述任务分配到的节点
type TaskSchedule struct {
	Sign      string // 任务标识
	Slot      int    // 第几次被调用，这里会出现一次调度不被接受，因此需要多次调度的，称为slot
	Size      int    // 数据总量
	Model     string // 模型名称
	Params    map[string]interface{}
	Schedules []ScheduleItem
}
type ScheduleItem struct {
	//Sign   string
	//Slot   int
	Size int
	NID  int
	//Model  string
	//Params map[string]interface{}
}

// UnprocessedTask 格式化前端发来的请求
type UnprocessedTask struct {
	Sign   string                 // task sign
	Slot   int                    // slot index
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
	Schedule []TaskSchedule         // 调度
}

func (s *PendingSlotItem) UpdateSchedule(schedule []TaskSchedule) {
	s.Schedule = schedule
}

type SlotRecord struct {
	Slot     int            // id
	Schedule TaskSchedule   // 调度
	Process  []ScheduleItem // 完成情况
}

func NewSlotRecord(slot int) SlotRecord {
	return SlotRecord{
		Slot:     slot,
		Schedule: TaskSchedule{},
		Process:  make([]ScheduleItem, 0),
	}
}
