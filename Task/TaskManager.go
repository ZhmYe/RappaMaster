package Task

import (
	"BHLayer2Node/Config"
	"BHLayer2Node/LogWriter"
	"BHLayer2Node/paradigm"
	"fmt"
	"sync"
)

type TaskManager struct {
	tasks            map[string]*Task
	mu               sync.Mutex
	tracker          *Tracker
	scheduledTasks   chan paradigm.TaskSchedule
	commitSlot       chan paradigm.CommitSlotItem
	unprocessedTasks chan paradigm.UnprocessedTask

	epochChangeEvent chan bool // 外部触发的 epoch 更新信号
	currentEpoch     int
}

func (t *TaskManager) Start() {
	processTasks := func() {
		for {
			select {
			case <-t.epochChangeEvent:
				t.UpdateEpoch() // 如果epoch更新，那么先更新epoch，此时有新的任务也会进入下一个epoch
			case commitSlotItem := <-t.commitSlot: // 如果不需要更新epoch，那么优先提交
				err := t.Commit(commitSlotItem)
				if err != nil {
					LogWriter.Log("ERROR", err.Error())
					continue
				}
			case schedule := <-t.scheduledTasks:
				_, err := t.UpdateTaskSchedule(schedule)
				// 不合法
				if err != nil {
					LogWriter.Log("ERROR", err.Error())
					continue
				}
				// 将任务添加到对应剩余时间的桶,这里只记录sign即可
				t.tracker.Update(schedule.Sign)
			}
		}
	}
	go processTasks()
}

func (t *TaskManager) CheckTaskIsFinish(sign string) bool {
	if task, exist := t.tasks[sign]; !exist {
		return false
	} else {
		return task.IsFinish()
	}
}
func (t *TaskManager) CheckSlotIsValid(sign string, slot int) (bool, error) {
	if _, exist := t.tasks[sign]; !exist {
		return false, fmt.Errorf("invalid Task")
	}
	task := t.tasks[sign]
	if task.Slot != slot {
		return false, fmt.Errorf("invalid Task Slot, expected: %d, given: %d", task.Slot, slot)
	}
	return true, nil
}
func (t *TaskManager) UpdateTaskSchedule(schedule paradigm.TaskSchedule) (bool, error) {
	sign, slot := schedule.Sign, schedule.Slot
	if _, exist := t.tasks[sign]; !exist {
		t.tasks[sign] = NewTask(sign, schedule.Model, schedule.Params, schedule.Size)
		LogWriter.Log("TRACKER", fmt.Sprintf("Update New Task, sign: %s, slot: 0", sign))
	}
	task := t.tasks[sign]
	_, err := t.CheckSlotIsValid(sign, slot)
	if err != nil {
		return false, err
	}
	err = task.UpdateSchedule(schedule) // 更新slot
	if err != nil {
		return false, err
	}
	return true, nil
}
func (t *TaskManager) Commit(slot paradigm.CommitSlotItem) error {
	task, exist := t.tasks[slot.Sign]
	if !exist {
		return fmt.Errorf("task %s does not exist", slot.Sign)
	}
	return task.Commit(slot)
}
func (t *TaskManager) UpdateEpoch() {
	t.currentEpoch++
	outOfDateTasks := t.tracker.OutOfDate()
	for _, sign := range outOfDateTasks {
		if t.CheckTaskIsFinish(sign) {
			continue
		}
		task := t.tasks[sign]
		nextSlot, _ := task.Next()
		go func(slot paradigm.UnprocessedTask) {
			t.unprocessedTasks <- slot
		}(nextSlot)
	}
}

func NewTaskManager(config Config.BHLayer2NodeConfig, scheduledTasks chan paradigm.TaskSchedule,
	commitSlot chan paradigm.CommitSlotItem, unprocessedTasks chan paradigm.UnprocessedTask, epochChangeEvent chan bool) *TaskManager {
	return &TaskManager{
		tasks:            make(map[string]*Task),
		mu:               sync.Mutex{},
		tracker:          NewTracker(config),
		scheduledTasks:   scheduledTasks,
		commitSlot:       commitSlot,
		unprocessedTasks: unprocessedTasks,
		epochChangeEvent: epochChangeEvent,
		currentEpoch:     -1,
	}
}
