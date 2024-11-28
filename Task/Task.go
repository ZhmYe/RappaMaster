package Task

import (
	"BHLayer2Node/LogWriter"
	"BHLayer2Node/paradigm"
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
	if t.Slot != slot {
		// 说明之前已经更新过slot的record或者还没到slot

		//fmt.Println(len(t.records), slot)
		return fmt.Errorf(fmt.Sprintf("invalid schedule Slot, expected: %d, given: %d", len(t.records), slot))
	}
	for len(t.records) <= slot {
		t.records = append(t.records, paradigm.NewSlotRecord(len(t.records)))
	}
	record := t.records[slot]
	record.Schedule = schedule
	t.records[slot] = record
	return nil
}
func (t *Task) Commit(slot paradigm.CommitSlotItem) error {
	if slot.Slot != t.Slot {
		return fmt.Errorf(fmt.Sprintf("invalid commit Slot, expected: %d, given: %d", t.Slot, slot.Slot))
	}
	for len(t.records) <= slot.Slot {
		t.records = append(t.records, paradigm.NewSlotRecord(len(t.records)))
	}
	slotRecord := t.records[slot.Slot]
	slotRecord.Process = append(slotRecord.Process, slot.Record())
	t.process += slot.Process
	LogWriter.Log("DEBUG", fmt.Sprintf("Task %s process %d by node %d", slot.Sign, slot.Process, slot.Nid))
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
