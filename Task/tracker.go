package Task

import (
	"BHLayer2Node/Config"
	"sync"
)

// Tracker 追踪任务状态
type Tracker struct {
	config        *Config.BHLayer2NodeConfig
	taskBuckets   [][]string // 二维切片，每个索引对应剩余时间
	maxEpochDelay int
	mu            sync.Mutex
}

func (t *Tracker) Update(sign string) {
	remainingEpoch := t.maxEpochDelay
	if remainingEpoch >= len(t.taskBuckets) {
		t.expandBuckets(remainingEpoch)
	}
	t.taskBuckets[remainingEpoch] = append(t.taskBuckets[remainingEpoch], sign)

}

//// Start 开始监听任务调度
//func (t *Tracker) Start() {
//	// 处理已经调度的任务，追踪其任务状态，过了maxEpochDelay后将其作为过期任务，生成新的slot
//	processScheduledTasks := func() {
//		for task := range t.scheduledTasks {
//			select {
//			case <-t.epochChangeEvent:
//				t.UpdateEpoch() // 如果epoch更新，那么先更新epoch，此时有新的任务也会进入下一个epoch
//			default:
//				t.mu.Lock()
//				remainingEpoch := t.maxEpochDelay
//				if remainingEpoch >= len(t.taskBuckets) {
//					t.expandBuckets(remainingEpoch)
//				}
//				if _, exist := t.slotTrack[task.Sign]; !exist {
//					// 说明是第一次收到这个任务，那么slot应该是0
//					t.slotTrack[task.Sign] = 0
//				}
//				// 不合法
//				//if !t.IsValidSlot(task.Sign, task.Slot) {
//				//	LogWriter.Log("ERROR", fmt.Sprintf("Invalid Task %s Slot, expect: %d, given: %d", task.Sign, t.slotTrack[task.Sign], task.Slot))
//				//}
//				_, err := t.manager.UpdateTaskSchedule(task)
//				if err != nil {
//					LogWriter.Log("ERROR", err.Error())
//					continue
//				}
//				// 将任务添加到对应剩余时间的桶,这里只记录sign即可
//				t.taskBuckets[remainingEpoch] = append(t.taskBuckets[remainingEpoch], task.Sign)
//				t.mu.Unlock()
//			}
//		}
//	}
//	go processScheduledTasks()
//}

//// IsValidSlot 确认是不是合法的slot
//func (t *Tracker) IsValidSlot(sign string, slot int) bool {
//	if _, exist := t.slotTrack[sign]; !exist {
//		t.slotTrack[sign] = 0
//	}
//	return slot == t.slotTrack[sign]
//}

func (t *Tracker) OutOfDate() []string {

	// 处理剩余时间为 0 的任务
	if len(t.taskBuckets) > 0 {
		//for _, task := range t.taskBuckets[0] {
		//	// 更新任务 Slot 并重新分配
		//	if task.Size == 0 {
		//		LogWriter.Log("DEBUG", fmt.Sprintf("find task %s finish, slot: %d", task.Sign, task.Slot))
		//		continue // 如果已经做完了就清除即可，这里还要考虑什么提交任务结束 todo
		//	}
		//	// 如果还没做完，但是已经到时间了，那么发起一个新的slot，再次调度
		//	task.Slot += 1
		//	t.slotTrack[task.Sign] = task.Slot // 更新合法的slot
		//	t.unprocessedTasks <- *task
		//	//t.AddTaskToBucket(t.maxEpochDelay, task)
		//}
		//// 移除过期的交易，从此更新所有合法的slot，遇到过期的slot直接忽略
		outOfDateTaskSlot := t.taskBuckets[0]
		t.taskBuckets = t.taskBuckets[1:]
		return outOfDateTaskSlot
	}
	return []string{}

}

// expandBuckets 扩展桶的数量
func (t *Tracker) expandBuckets(required int) {
	for len(t.taskBuckets) <= required {
		t.taskBuckets = append(t.taskBuckets, []string{})
	}
}

// Setup 初始化 Tracker
func (t *Tracker) Setup(config *Config.BHLayer2NodeConfig) {
	t.config = config
	t.taskBuckets = make([][]string, 0)
	t.maxEpochDelay = 1 // todo
}

// NewTracker 创建新的 Tracker
func NewTracker(config *Config.BHLayer2NodeConfig) *Tracker {
	tracker := &Tracker{}
	tracker.Setup(config)
	return tracker
}
