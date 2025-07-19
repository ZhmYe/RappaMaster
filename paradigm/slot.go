package paradigm

type SlotStatus int

const (
	Finished SlotStatus = iota
	Processing
	Failed
)

// Slot 一个具体的节点合成任务实例
type Slot struct {
	SlotID       SlotHash        `gorm:"primaryKey;size:256;comment:Slot哈希，作为主键"`
	TaskID       TaskHash        `gorm:"not null;size:256;comment:任务哈希（TaskHash）"`
	ScheduleID   ScheduleHash    `gorm:"not null;comment:调度哈希（ScheduleHash）"`
	NodeID       int32           `gorm:"not null;comment:节点id"`
	ScheduleSize int32           `gorm:"not null;comment:调度的数量，以KB为单位"`
	Status       SlotStatus      `gorm:"type:tinyint;not null;comment:完成状态"`
	Err          string          `gorm:"type:text;comment:错误信息"`
	CommitSlot   *CommitSlotItem `gorm:"type:json;serializer:json;comment:提交的CommitSlot信息"`
	Epoch        int32           `gorm:"comment:Epoch ID"`
}

func (s *Slot) Json() map[string]interface{} {
	json := make(map[string]interface{})
	json["slotHash"] = s.SlotID
	json["scheduleID"] = s.ScheduleID
	json["scheduleSize"] = s.ScheduleSize
	json["status"] = s.Status
	json["Err"] = s.Err
	json["nodeID"] = s.NodeID
	if s.CommitSlot != nil {
		json["store"] = s.CommitSlot.Store
		json["commitment"] = s.CommitSlot.Commitment
		json["process"] = s.CommitSlot.Process
	}

	json["epoch"] = s.Epoch
	return json
}
func (s *Slot) SetError(errorMessage string) {
	s.Err = errorMessage
	s.Status = Failed
}
func (s *Slot) SetEpoch(epoch int32) {
	s.Epoch = epoch
}
func (s *Slot) ErrorMessage() string {
	return s.Err
}

// Commit 将节点提交的结果commit，这里不做完整性等校验，在外面校验完才放到这里
func (s *Slot) Commit(commitSlot *CommitSlotItem) {
	s.CommitSlot = commitSlot
	s.Status = Finished // 这里不区分是否全部做完，不允许多次提交 todo
	s.SetEpoch(commitSlot.Epoch)
}
func (s *Slot) UpdateSchedule(scheduleID ScheduleHash, taskID TaskHash, size int32) {
	//s.SlotID = slotID
	s.ScheduleID = scheduleID
	s.ScheduleSize = size
	s.TaskID = taskID
}
func NewSlotWithSlotID(slotID SlotHash) *Slot {
	return &Slot{
		SlotID:       slotID,
		ScheduleID:   -1,
		ScheduleSize: 0,
		Status:       Processing,
		Err:          "",
		CommitSlot:   nil,
		TaskID:       "",
	}
}
func NewSlot(slotID SlotHash, taskID TaskHash, scheduleID ScheduleHash, schedule int32) *Slot {
	return &Slot{
		TaskID:       taskID,
		SlotID:       slotID,
		ScheduleID:   scheduleID,
		ScheduleSize: schedule,
		Status:       Processing,
		Err:          "",
		CommitSlot:   nil,
		Epoch:        -1,
	}
}
