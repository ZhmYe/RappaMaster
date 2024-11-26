package Task

import (
	"BHCoordinator/paradigm"
	"fmt"
)

// Task 描述一个合成任务
type Task struct {
	Sign    string
	Slot    int
	Model   string
	Params  map[string]interface{}
	size    int // 总的数据量
	process int // 已经完成的数据量

	records []paradigm.SlotRecord // 记录每个slot的调度和完成情况
}

// UpdateSchedule 更新调度情况
func (t *Task) UpdateSchedule(schedule paradigm.TaskSchedule) error {
	slot := schedule.Slot
	if len(t.records) != slot {
		// 说明之前已经更新过slot的record或者还没到slot
		return fmt.Errorf("invalid schedule Slot")
	}
	record := paradigm.NewSlotRecord(slot)
	t.records = append(t.records, record) // 更新记录
	return nil
}
func (t *Task) Commit(slot paradigm.CommitSlotItem) error {
	if slot.Slot != len(t.records) {
		// 说明之前已经更新过slot的record或者还没到slot
		return fmt.Errorf("invalid schedule Slot")
	}
	slotRecord := t.records[slot.Slot]
	slotRecord.Process = append(slotRecord.Process, slot.Record())
	t.records[slot.Slot] = slotRecord
	return nil
}
func (t *Task) Next() (paradigm.UnprocessedTask, error) {
	if t.IsFinish() {
		return paradigm.UnprocessedTask{}, fmt.Errorf("task %s has been finished", t.Sign)
	}
	t.Slot++
	slot := paradigm.UnprocessedTask{
		Sign:   t.Sign,
		Slot:   t.Slot,
		Size:   t.Remain(),
		Model:  t.Model,
		Params: t.Params,
	}
	return slot, nil
}

func (t *Task) Remain() int {
	if t.IsFinish() {
		return 0
	}
	return t.size - t.process
}
func (t *Task) IsFinish() bool {
	return t.process >= t.size
}

func NewTask(sign string, model string, params map[string]interface{}, total int) *Task {
	return &Task{
		Sign:    sign,
		Slot:    0,
		Model:   model,
		Params:  params,
		size:    total,
		process: 0,
		records: make([]paradigm.SlotRecord, 0),
	}
}
