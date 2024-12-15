package Task

import (
	"BHLayer2Node/Config"
	"BHLayer2Node/LogWriter"
	"BHLayer2Node/paradigm"
	pb "BHLayer2Node/pb/service"
	"fmt"
	"sync"
)

type TaskManager struct {
	tasks   map[string]*Task
	mu      sync.Mutex
	tracker *Tracker

	scheduledTasks chan paradigm.TaskSchedule
	// 这里我们假定，正确节点需要在分发所有的数据块并抱枕所有数据块都落实后才发送commit
	// 但恶意节点可以故意发送commit消息，因此我们不能直接commit，需要等待一轮投票
	commitSlot chan paradigm.CommitSlotItem // 这里是单纯的commit上来的justified或者finalize的
	//finalizeSlot        chan paradigm.CommitSlotItem // 这里是finalize的
	unprocessedTasks    chan paradigm.UnprocessedTask
	initTasks           chan paradigm.UnprocessedTask
	pendingTransactions chan paradigm.Transaction
	//slotToVotes      chan paradigm.CommitSlotItem
	epochChangeEvent chan bool // 外部触发的 epoch 更新信号
	currentEpoch     int
	epochRecord      *EpochRecord
	epochHeartbeat   chan *pb.HeartbeatRequest
}

func (t *TaskManager) Start() {
	processTasks := func() {
		for {
			select {
			case <-t.epochChangeEvent:
				t.UpdateEpoch() // 如果epoch更新，那么先更新epoch，此时有新的任务也会进入下一个epoch
			case commitSlotItem := <-t.commitSlot: // 如果不需要更新epoch，那么优先commit或finalize
				// 判断类别,如果是新commit的则commit，如果已经通过投票，则finalize
				switch commitSlotItem.State() {
				case paradigm.JUSTIFIED:
					// 这里需要先确认一下这个slot是否是合法的, 如果这个slot已经是过时的了，没有必要进入投票
					err := t.CheckSlotIsValid(commitSlotItem.Sign, int(commitSlotItem.Slot))
					if err != nil {
						LogWriter.Log("ERROR", err.Error())
						continue
					}
					commitSlotItem.SetEpoch(int32(t.currentEpoch))
					t.epochRecord.commit(commitSlotItem)
				case paradigm.FINALIZE:
					// 这里直接commit，commit里不需要额外的check,随时可以上链
					err := t.Commit(commitSlotItem)
					if err != nil {
						LogWriter.Log("ERROR", err.Error())
						continue
					}
					t.epochRecord.finalize(commitSlotItem)
					t.pendingTransactions <- &paradigm.CommitSlotTransaction{
						CommitSlotItem: commitSlotItem,
						Epoch:          t.currentEpoch,
					}
				default:
					panic("An Invalid or Abort CommitSlotItem should not be involved in commitSlot!!!")
				}
				//err := t.Commit(commitSlotItem)
				//if err != nil {
				//	LogWriter.Log("ERROR", err.Error())
				//	continue
				//}
				//t.pendingTransactions <- &paradigm.CommitSlotTransaction{
				//	CommitSlotItem: commitSlotItem,
				//	Epoch:          t.currentEpoch,
				//}
			case initTask := <-t.initTasks:
				t.UpdateTask(initTask.Sign, initTask.Model, initTask.Size, initTask.Params)
			case schedule := <-t.scheduledTasks:
				_, err := t.UpdateTaskSchedule(schedule)
				// 不合法
				if err != nil {
					LogWriter.Log("ERROR", err.Error())
					continue
				}
				// 将任务添加到对应剩余时间的桶,这里只记录sign即可
				t.tracker.Update(schedule.Sign)
			default:
				continue
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
func (t *TaskManager) CheckSlotIsValid(sign string, slot int) error {
	if _, exist := t.tasks[sign]; !exist {
		return fmt.Errorf("invalid Task")
	}
	task := t.tasks[sign]
	if task.Slot != slot {
		return fmt.Errorf("invalid Task Slot, expected: %d, given: %d", task.Slot, slot)
	}
	return nil
}
func (t *TaskManager) UpdateTask(sign string, model string, size int32, params map[string]interface{}) {
	if _, exist := t.tasks[sign]; !exist {
		task := NewTask(sign, model, params, size)
		t.tasks[sign] = task
		nextSlot, _ := task.Next()
		go func(slot paradigm.UnprocessedTask) {
			t.unprocessedTasks <- slot
		}(nextSlot)
		LogWriter.Log("TRACKER", fmt.Sprintf("Update New Task, sign: %s, slot: 0", sign))
	}
}

func (t *TaskManager) UpdateTaskSchedule(schedule paradigm.TaskSchedule) (bool, error) {
	sign, slot := schedule.Sign, schedule.Slot
	//if _, exist := t.tasks[sign]; !exist {
	//	//t.tasks[sign] = NewTask(sign, schedule.Model, schedule.Params, schedule.Size)
	//	LogWriter.Log("ERROR", fmt.Sprintf("Task %s does not exist", sign))
	//}
	t.UpdateTask(sign, schedule.Model, schedule.Size, schedule.Params)
	task := t.tasks[sign]
	err := t.CheckSlotIsValid(sign, slot)
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
	LogWriter.Log("TRACKER", fmt.Sprintf("Epoch update, current Epoch: %d", t.currentEpoch))
	outOfDateTasks := t.tracker.OutOfDate()
	validTaskMap := make(map[string]int32)
	for _, sign := range outOfDateTasks {
		task := t.tasks[sign]
		if t.CheckTaskIsFinish(sign) {
			LogWriter.Log("TRACKER", fmt.Sprintf("Task %s finished at slot %d, expected: %d, processed: %d", sign, task.Slot, task.size, task.process))
			continue
		}
		nextSlot, _ := task.Next()
		validTaskMap[nextSlot.Sign] = int32(nextSlot.Slot)
		go func(slot paradigm.UnprocessedTask) {
			t.unprocessedTasks <- slot
		}(nextSlot)
	}
	// 更新epoch的时候，构建心跳
	heartbeat := t.buildHeartbeat(validTaskMap)
	t.epochHeartbeat <- heartbeat
	t.epochRecord.Refresh()

}
func (t *TaskManager) buildHeartbeat(validTaskMap map[string]int32) *pb.HeartbeatRequest {
	//fmt.Println(len(t.epochRecord.commits), len(t.epochRecord.finalizes), 111)
	return &pb.HeartbeatRequest{
		Commits:   t.epochRecord.commits,
		Finalizes: t.epochRecord.finalizes,
		Tasks:     validTaskMap,
		Epoch:     int32(t.currentEpoch),
	}
}
func NewTaskManager(config Config.BHLayer2NodeConfig, scheduledTasks chan paradigm.TaskSchedule,
	commitSlot chan paradigm.CommitSlotItem, unprocessedTasks chan paradigm.UnprocessedTask, epochChangeEvent chan bool,
	initTasks chan paradigm.UnprocessedTask, pendingTransactions chan paradigm.Transaction, epochHeartbeat chan *pb.HeartbeatRequest) *TaskManager {
	return &TaskManager{
		tasks:               make(map[string]*Task),
		mu:                  sync.Mutex{},
		tracker:             NewTracker(config),
		scheduledTasks:      scheduledTasks,
		commitSlot:          commitSlot,
		unprocessedTasks:    unprocessedTasks,
		epochChangeEvent:    epochChangeEvent,
		initTasks:           initTasks,
		pendingTransactions: pendingTransactions,
		epochHeartbeat:      epochHeartbeat,
		//slotToVotes:      slotToVotes,
		epochRecord:  NewEpochRecord(),
		currentEpoch: -1,
	}
}
