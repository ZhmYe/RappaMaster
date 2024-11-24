package Task

import (
	"BHLayer2node/paradigm"
	"fmt"
)

// Task 描述一个未完成的任务
type Task struct {
	Sign   string
	Model  string
	Params map[string]interface{}
	Slot   int // 下一个要分配的 Slot 编号

	total  int  // 所有的数据量
	size   int  // 剩余未处理的数据量
	finish bool // 是否已经完成
	// todo 这里是否还需要记录每个slot的运行情况
	record []paradigm.SlotRecord
}

// Update grpc会在发送一个task的slot后开始计时，在一段timeout（动态调整？）或收齐数据后，反馈上一个slot已经处理了多少数据了
// 然后由scheduler更新这一task的数据，并确认该task是否已经完成，如果已经完成，那么就需要上链最终信息(todo)
// 如果还没有完成，则调用task.remain()获取剩余数据并再次创建slot
func (t *Task) Update(r paradigm.SlotRecord) {
	t.record = append(t.record, r) // 更新记录
	t.size -= r.Size               // 减去这一slot处理的数据
	// todo暂时不考虑是否会出现负数的情况，按道理不会
	if t.size <= 0 {
		t.finish = true
	}
}
func (t *Task) Next(alloc int) (paradigm.PendingSlotItem, error) {
	if t.IsFinish() {
		return paradigm.PendingSlotItem{}, fmt.Errorf("task %s has been finished", t.Sign)
	}
	t.Slot++
	slot := paradigm.PendingSlotItem{
		Sign:     t.Sign,
		Slot:     t.Slot,
		Size:     alloc,
		Model:    t.Model,
		Params:   t.Params,
		Schedule: make([]paradigm.BHLayer2NodeSchedule, 0),
	}
	return slot, nil
}

func (t *Task) Remain() int {
	return t.size
}
func (t *Task) IsFinish() bool {
	return t.finish
}

func NewTask(sign string, model string, params map[string]interface{}, total int) *Task {
	return &Task{
		Sign:   sign,
		Model:  model,
		Params: params,
		//Slot:   slot,
		total:  total,
		size:   total,
		finish: false,
		record: make([]paradigm.SlotRecord, 0),
	}
}
