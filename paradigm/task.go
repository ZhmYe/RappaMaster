package paradigm

import (
	"BHLayer2Node/LogWriter"
	"fmt"
)

// Task 描述一个合成任务
type Task struct {
	Sign       string
	Slot       int32
	Model      SupportModelType
	Params     map[string]interface{}
	Size       int32           // 总的数据量
	Process    int32           // 已经完成的数据量
	isReliable bool            // 是否可信 TODO: @YZM 这里需要加入任务是否可信的部分，这个需要在http前端得到
	OutputType ModelOutputType // 模型类型，这里假定一个任务一个模型
	//records    []paradigm.SlotRecord // 记录每个slot的调度和完成情况
}

// UpdateSchedule 更新调度情况
func (t *Task) UpdateSchedule(schedule TaskSchedule) error {
	//slot := schedule.Slot
	//if t.Slot != slot {
	// 说明之前已经更新过slot的record或者还没到slot

	//fmt.Println(len(t.records), slot)
	//return fmt.Errorf(fmt.Sprintf("invalid schedule Slot, expected: %d, given: %d", len(t.records), slot))
	//}
	//for len(t.records) <= slot {
	//	t.records = append(t.records, paradigm.NewSlotRecord(len(t.records)))
	//}
	//record := t.records[slot]
	//record.Schedule = schedule
	//t.records[slot] = record
	return nil
}

func (t *Task) Commit(slot *CommitSlotItem) error {
	if slot.State() != FINALIZE {
		return fmt.Errorf("the commit Slot is not finalized") // 只能提交finalized的，因为已经通过投票了所以不需要check
	}
	//for len(t.records) <= int(slot.Slot) {
	//	t.records = append(t.records, paradigm.NewSlotRecord(len(t.records)))
	//}
	//slotRecord := t.records[slot.Slot]
	//slotRecord.Process = append(slotRecord.Process, slot.Record())
	t.Process += slot.Process
	LogWriter.Log("DEBUG", fmt.Sprintf("Task %s process %d by node %d", slot.Sign, slot.Process, slot.Nid))
	//t.records[slot.Slot] = slotRecord
	return nil
}
func (t *Task) IsReliable() bool {
	return t.isReliable
}
func (t *Task) Next() (UnprocessedTask, error) {
	if t.IsFinish() {
		return UnprocessedTask{}, fmt.Errorf("task %s has been finished", t.Sign)
	}
	t.Slot++
	slot := UnprocessedTask{
		Sign:   t.Sign,
		Slot:   t.Slot,
		Size:   t.Remain(),
		Model:  t.Model,
		Params: t.Params,
	}
	return slot, nil
}

func (t *Task) Remain() int32 {
	if t.IsFinish() {
		return 0
	}
	return t.Size - t.Process
}
func (t *Task) IsFinish() bool {
	return t.Process >= t.Size
}

func NewTask(sign string, model SupportModelType, params map[string]interface{}, total int32) *Task {
	outputType := DATAFRAME
	switch model {
	case CTGAN:
		outputType = DATAFRAME
	case AGSS:
		outputType = NETWORK
	default:
		panic("Unsupported Model Type!!!")
	}
	return &Task{
		Sign:       sign,
		Slot:       -1,
		Model:      model,
		OutputType: outputType,
		Params:     params,
		Size:       total,
		Process:    0,
		//records:    make([]paradigm.SlotRecord, 0),
		isReliable: true, // todo 这里先统一写成true
	}
}
