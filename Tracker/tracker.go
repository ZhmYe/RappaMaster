package Tracker

import (
	"BHLayer2Node/LogWriter"
	"BHLayer2Node/paradigm"
	"fmt"
	"sync"
)

// Tracker 用于监视任务进度，包括commit和expire
// 用于监视Slot进度,包括零知识证明proof的提交
type Tracker struct {
	//config            *Config.BHLayer2NodeConfig
	taskTracks        map[paradigm.TaskHash]*paradigm.SynthTaskTrackItem     // 用于维护每个任务的进度(剩余多少size)
	pendingCommitSlot map[paradigm.SlotHash]*paradigm.PendingCommitSlotTrack // 等待由Justified -> Finalized的slot
	channel           *paradigm.RappaChannel
	taskBuckets       [][]string            // 二维切片，每个索引对应剩余时间，这里是task slot的过期时间，也就是过期要开启新的slot
	slotBuckets       [][]paradigm.SlotHash // 二维切片，每个索引对应剩余时间，这里是一个commit slot的过期时间，要求一个commit slot必须携带zkp
	//maxEpochDelay     int
	mu sync.Mutex
}

func (t *Tracker) InitTask(initTask *paradigm.SynthTaskTrackItem) {
	t.taskTracks[initTask.TaskID] = initTask // todo 这里还没有判断重复
	// 新建完任务后，就可以生成一个unprocessTask用于调度
	LogWriter.Log("TRACK", fmt.Sprintf("Init New Task %s", initTask.TaskID))
	t.channel.UnprocessedTasks <- initTask.Next()
}

// UpdateTask 更新一个任务，已有的任务会根据OutOfDate()不断更新
func (t *Tracker) UpdateTask(sign string) {
	remainingEpoch := t.channel.Config.MaxEpochDelay
	//if remainingEpoch >= len(t.taskBuckets) {
	t.expandBuckets(remainingEpoch)
	//}
	t.taskBuckets[remainingEpoch] = append(t.taskBuckets[remainingEpoch], sign)
}
func (t *Tracker) Commit(slot *paradigm.CommitSlotItem) error {
	track, exist := t.taskTracks[slot.Sign]
	if !exist {
		return fmt.Errorf("task %s does not exist", slot.Sign)
	}
	return track.Commit(*slot)
}

// UpdateSlot 更新一个slot,等待其提交
func (t *Tracker) UpdateSlot(commitSlotItem paradigm.CommitSlotItem) {
	t.pendingCommitSlot[commitSlotItem.SlotHash()] = paradigm.NewPendingCommitSlotTrack(&commitSlotItem, t.checkIsReliable(commitSlotItem.Sign)) // 等待verify

	remainingEpoch := t.channel.Config.MaxEpochDelay
	//if remainingEpoch >= len(t.slotBuckets) {
	t.expandBuckets(remainingEpoch)
	//}
	t.slotBuckets[remainingEpoch] = append(t.slotBuckets[remainingEpoch], commitSlotItem.SlotHash())
}
func (t *Tracker) WonVote(commitSlotItem paradigm.CommitSlotItem) {
	if _, exist := t.pendingCommitSlot[commitSlotItem.SlotHash()]; exist {
		t.pendingCommitSlot[commitSlotItem.SlotHash()].WonVote()
	}
}

// OutOfDate 返回epoch中提交的slot和abort的slot
// todo 这里有延时 @YZM 可以多用一点空间换
func (t *Tracker) OutOfDate() ([]*paradigm.CommitSlotItem, []*paradigm.CommitSlotItem) {
	// 处理任务，这里无需返回
	processExpireTask := func() {
		if len(t.taskBuckets) == 0 {
			return
		}
		outOfDateTaskSlot := t.taskBuckets[0]
		t.taskBuckets = t.taskBuckets[1:]
		// 得到过期的task，需要重新调度
		for _, taskID := range outOfDateTaskSlot {
			task := t.taskTracks[taskID]
			if task.IsFinish() {
				LogWriter.Log("TRACKER", fmt.Sprintf("Task %s finished, expected: %d, processed: %d", taskID, task.Size, task.History))
				continue
			} else {
				fmt.Println(task.IsFinish(), task.Size)
			}
			//t.channel.UnprocessedTasks <- task.Next()
			//validTaskMap[nextSlot.Sign] = int32(nextSlot.Slot)
			//go func(task paradigm.SynthTaskTrackItem) {
			LogWriter.Log("TRACKER", fmt.Sprintf("Task %s Expire, unprocessed: %d, pass to schedule", task.TaskID, task.Size))
			t.channel.UnprocessedTasks <- task.Next()
			//}(*task)
		}
		//return outOfDateTaskSlot
	}
	processExpireSlot := func() ([]*paradigm.CommitSlotItem, []*paradigm.CommitSlotItem) {
		if len(t.slotBuckets) == 0 {
			return []*paradigm.CommitSlotItem{}, []*paradigm.CommitSlotItem{}
		}
		outOfDateCommitSlot := t.slotBuckets[0]
		t.slotBuckets = t.slotBuckets[1:]
		finalized := make([]*paradigm.CommitSlotItem, 0)
		abort := make([]*paradigm.CommitSlotItem, 0)
		for _, h := range outOfDateCommitSlot {
			pendingSlot := t.pendingCommitSlot[h]
			commitSlotItem := pendingSlot.CommitSlotItem
			if pendingSlot.Check() {
				// 这个commitSlot在指定时间内完成了存储任务(vote)和可信任务(zkp)
				//commitSlotItem.SetEpoch(int32(t.currentEpoch)) // 统一都设置这个epoch
				//commitSlotItem.SetFinalize()
				//err := t.Commit(commitSlotItem) // 正式更新任务
				//if err != nil {
				//	LogWriter.Log("ERROR", err.Error())
				//	continue
				//}
				//t.epochRecord.Finalize(commitSlotItem)
				finalized = append(finalized, commitSlotItem)
				// 上链任务推进情况
				//go func(transaction *paradigm.TaskProcessTransaction) {
				//	t.channel.PendingTransactions <- transaction
				//}(&paradigm.TaskProcessTransaction{
				//	CommitSlotItem: commitSlotItem,
				//	Proof:          nil,
				//	Signatures:     nil,
				//})
			} else {
				// 未在指定时间内完成，那么直接丢弃
				commitSlotItem.SetInvalid(paradigm.VERIFIED_FAILED)
				abort = append(abort, commitSlotItem)
				//t.epochRecord.Abort(pendingSlot.CommitSlotItem, paradigm.VERIFIED_FAILED)
				// 这里会出现节点后面才额外提交zkp，但已经失效了，直接无视，也就是commitzkp(还没写)的时候发现没有这个任务，那么要么没有通过justified(这是commitSlot的前置，得到hash和seed)
				// 要么就是已经失效了，直接无视
			}
			delete(t.pendingCommitSlot, h) // 标记为已完成，不需要记录了
		}
		return finalized, abort
	}
	processExpireTask()
	return processExpireSlot()

}

// expandBuckets 扩展桶的数量
func (t *Tracker) expandBuckets(required int) {
	for len(t.taskBuckets) <= required {
		t.taskBuckets = append(t.taskBuckets, []string{})
	}
	for len(t.slotBuckets) <= required {
		t.slotBuckets = append(t.slotBuckets, []paradigm.SlotHash{})
	}
}
func (t *Tracker) checkIsReliable(sign string) bool {
	if _, exist := t.taskTracks[sign]; !exist {
		return false
	}
	track := t.taskTracks[sign]
	return track.IsReliable
}
func (t *Tracker) checkTaskIsFinish(sign string) bool {
	if task, exist := t.taskTracks[sign]; !exist {
		return false
	} else {
		return task.IsFinish()
	}
}
func (t *Tracker) Start() {
	for {
		select {
		case initTask := <-t.channel.InitTasks:
			t.InitTask(initTask)
		case scheduledTask := <-t.channel.ScheduledTasks:
			t.UpdateTask(scheduledTask.TaskID) // 开始计时

		}
	}
	//for scheduledTask := range t.channel.ScheduledTasks {
	//	//select {
	//	//case initTask := <-t.channel.InitTasks:
	//	//	t.taskTracks[initTask.TaskID] = initTask // todo 这里还没有判断重复
	//	//	// 新建完任务后，就可以生成一个unprocessTask用于调度
	//	//	t.channel.UnprocessedTasks <- initTask.Next()
	//	//case scheduledTask := <-t.channel.ScheduledTasks:
	//	// 调度完的任务
	//	t.UpdateTask(scheduledTask.TaskID) // 开始计时
	//	//}
	//}

}

//// Setup 初始化 Tracker
//func (t *Tracker) Setup(config *Config.BHLayer2NodeConfig) {
//	t.config = config
//	t.taskBuckets = make([][]string, 0)
//	t.maxEpochDelay = config.MaxEpochDelay
//}

// NewTracker 创建新的 Tracker
func NewTracker(channel *paradigm.RappaChannel) *Tracker {
	return &Tracker{
		taskTracks:        make(map[paradigm.TaskHash]*paradigm.SynthTaskTrackItem),
		pendingCommitSlot: make(map[paradigm.SlotHash]*paradigm.PendingCommitSlotTrack),
		channel:           channel,
		taskBuckets:       make([][]string, 0),
		slotBuckets:       make([][]paradigm.SlotHash, 0),
		mu:                sync.Mutex{},
	}
}
