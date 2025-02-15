package Tracker

import (
	"BHLayer2Node/paradigm"
	"fmt"
	"sync/atomic"
	"time"
)

// Tracker 用于监视任务进度，包括commit和expire
// 用于监视Slot进度,包括零知识证明proof的提交
type Tracker struct {
	//config            *Config.BHLayer2NodeConfig
	taskTracks        map[paradigm.TaskHash]*paradigm.SynthTaskTrackItem     // 用于维护每个任务的进度(剩余多少size)
	pendingCommitSlot map[paradigm.SlotHash]*paradigm.PendingCommitSlotTrack // 等待由Justified -> Finalized的slot
	channel           *paradigm.RappaChannel
	//taskBuckets         [][]string // 二维切片，每个索引对应剩余时间，这里是task slot的过期时间，也就是过期要开启新的slot
	expireInputChannel  chan paradigm.ExpireItem
	expireOutputChannel chan paradigm.ExpireItem
	eliminators         []*Eliminator
	//
	//slotBuckets [][]paradigm.SlotHash // 二维切片，每个索引对应剩余时间，这里是一个commit slot的过期时间，要求一个commit slot必须携带zkp,过期时间为了防止作恶占存储空间 TODO 因此executor需要根据invalid来删除
	//maxEpochDelay     int
}

func (t *Tracker) InitTask(initTask *paradigm.SynthTaskTrackItem) {
	t.taskTracks[initTask.TaskID] = initTask // todo 这里还没有判断重复
	// 新建完任务后，就可以生成一个unprocessTask用于调度
	paradigm.Print("INFO", fmt.Sprintf("Init New Task, TaskID: %s", initTask.TaskID))
	t.channel.UnprocessedTasks <- initTask.Next()
}

// UpdateTask 更新一个任务，已有的任务会根据OutOfDate()不断更新
func (t *Tracker) UpdateTask(sign string) {
	// 每个任务设置10s的时间
	expireTime := time.Now().Add(10 * time.Second) // todo @SD 这里的时间写成config
	expireTask := &paradigm.ExpireTask{
		BasicTimeExpire: paradigm.NewBasicTimeExpire(expireTime),
		TaskID:          sign,
	}
	t.expireInputChannel <- expireTask

}
func (t *Tracker) Commit(slot *paradigm.CommitSlotItem) error {
	track, exist := t.taskTracks[slot.Sign]
	if !exist {
		e := paradigm.Error(paradigm.RuntimeError, fmt.Sprintf("Task %s does not exist in Tracker", slot.Sign))
		return fmt.Errorf(e.Error())
	}
	return track.Commit(*slot)
}

// UpdateSlot 更新一个slot,等待其提交
func (t *Tracker) UpdateSlot(commitSlotItem paradigm.CommitSlotItem) {
	slot := paradigm.NewPendingCommitSlotTrack(&commitSlotItem, t.checkIsReliable(commitSlotItem.Sign)) // 等待verify
	t.pendingCommitSlot[commitSlotItem.SlotHash()] = slot
	// 每个slot设置比较长的时间，因为是zkp，设置1分钟吧先 todo @SD 这里的时间写成config，按秒
	expireTime := time.Now().Add(1 * time.Minute)
	expireSlot := &paradigm.ExpireSlot{
		BasicTimeExpire: paradigm.NewBasicTimeExpire(expireTime),
		SlotHash:        commitSlotItem.SlotHash(),
		PendingSlot:     slot,
	}
	t.expireInputChannel <- expireSlot
}
func (t *Tracker) WonVote(slotHash string) {
	if slot, exist := t.pendingCommitSlot[slotHash]; exist {
		if atomic.LoadInt32(&slot.IsFinalized) == 1 {
			return // 如果已经提交过，直接跳过
		}
		slot.WonVote()
		t.CheckFinalized(*slot)
	}

}
func (t *Tracker) CheckFinalized(slot paradigm.PendingCommitSlotTrack) {
	if t.pendingCommitSlot[slot.SlotHash()].Check() {
		commitSlot := slot.CommitSlotItem
		commitSlot.SetFinalize()
		atomic.StoreInt32(&slot.IsFinalized, 1)
		t.channel.CommitSlots <- *commitSlot
	}
}
func (t *Tracker) ReceiveProof(slotHash string) {
	if slot, exist := t.pendingCommitSlot[slotHash]; exist {
		if atomic.LoadInt32(&slot.IsFinalized) == 1 {
			return // 如果已经提交过，直接跳过
		}
		slot.ReceiveProof()
		t.CheckFinalized(*slot)
	}

}
func (t *Tracker) CollectExpire() {
	for expireItem := range t.expireOutputChannel {
		switch expireItem.(type) {
		case *paradigm.ExpireTask:
			expireTask := expireItem.(*paradigm.ExpireTask)
			task := t.taskTracks[expireTask.TaskID]
			if task.IsFinish() {
				paradigm.Print("TRACKER", fmt.Sprintf("Task %s finished in Tracker, Not need schedule, expected: %d, processed: %d", expireTask.TaskID, task.Size, task.History))
				continue
			}
			paradigm.Print("TRACKER", fmt.Sprintf("Task %s Expire, unprocessed: %d, pass to schedule", task.TaskID, task.Size))
			t.channel.UnprocessedTasks <- task.Next()
		case *paradigm.ExpireSlot:

			expireSlot := expireItem.(*paradigm.ExpireSlot)
			// TODO @YZM 这里按理会有并发错误，目前不太可能出现
			if !expireSlot.PendingSlot.Check() {
				// finalized 按道理这里不可能
				//commitSlot := slot.CommitSlotItem
				//commitSlot.SetFinalize()
				//t.channel.CommitSlots <- *commitSlot
				//} else {
				// abort
				commitSlot := expireSlot.PendingSlot.CommitSlotItem
				commitSlot.SetInvalid(paradigm.VERIFIED_FAILED)
				t.channel.CommitSlots <- *commitSlot
			}
			delete(t.pendingCommitSlot, expireSlot.SlotHash) // 标记为已完成，不需要记录了

		default:
			paradigm.Error(paradigm.RuntimeError, "Unknown ExpireItem Type")

		}
	}
}

// OutOfDate 返回epoch中提交的slot和abort的slot
// todo 这里有延时 @YZM 可以多用一点空间换
//func (t *Tracker) OutOfDate() ([]*paradigm.CommitSlotItem, []*paradigm.CommitSlotItem) {

// 处理任务，这里无需返回
//processExpireTask := func() {
//	if len(t.taskBuckets) == 0 {
//		return
//	}
//	outOfDateTaskSlot := t.taskBuckets[0]
//	t.taskBuckets = t.taskBuckets[1:]
//	// 得到过期的task，需要重新调度
//	for _, taskID := range outOfDateTaskSlot {
//		task := t.taskTracks[taskID]
//		if task.IsFinish() {
//			LogWriter.Log("TRACKER", fmt.Sprintf("Task %s finished, expected: %d, processed: %d", taskID, task.Size, task.History))
//			continue
//		}
//		//fmt.Println(task.IsFinish(), task.Size)
//
//		//t.channel.UnprocessedTasks <- task.Next()
//		//validTaskMap[nextSlot.Sign] = int32(nextSlot.Slot)
//		//go func(task paradigm.SynthTaskTrackItem) {
//		LogWriter.Log("TRACKER", fmt.Sprintf("Task %s Expire, unprocessed: %d, pass to schedule", task.TaskID, task.Size))
//		t.channel.UnprocessedTasks <- task.Next()
//		//}(*task)
//	}
//	//return outOfDateTaskSlot
//}
//processExpireSlot := func() ([]*paradigm.CommitSlotItem, []*paradigm.CommitSlotItem) {
//	if len(t.slotBuckets) == 0 {
//		return []*paradigm.CommitSlotItem{}, []*paradigm.CommitSlotItem{}
//	}
//	outOfDateCommitSlot := t.slotBuckets[0]
//	t.slotBuckets = t.slotBuckets[1:]
//	finalized := make([]*paradigm.CommitSlotItem, 0)
//	abort := make([]*paradigm.CommitSlotItem, 0)
//	for _, h := range outOfDateCommitSlot {
//		pendingSlot := t.pendingCommitSlot[h]
//		commitSlotItem := pendingSlot.CommitSlotItem
//		if pendingSlot.Check() {
//			finalized = append(finalized, commitSlotItem)
//		} else {
//			// 未在指定时间内完成，那么直接丢弃
//			commitSlotItem.SetInvalid(paradigm.VERIFIED_FAILED)
//			abort = append(abort, commitSlotItem)
//		}
//		delete(t.pendingCommitSlot, h) // 标记为已完成，不需要记录了
//	}
//	return finalized, abort
//}
//processExpireTask()
//return processExpireSlot()

//}

// expandBuckets 扩展桶的数量
//
//	func (t *Tracker) expandBuckets(required int) {
//		for len(t.taskBuckets) <= required {
//			t.taskBuckets = append(t.taskBuckets, []string{})
//		}
//		for len(t.slotBuckets) <= required {
//			t.slotBuckets = append(t.slotBuckets, []paradigm.SlotHash{})
//		}
//	}
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
	for _, e := range t.eliminators {
		go e.Start()
	}
	go t.CollectExpire()
	for {
		select {
		case initTask := <-t.channel.InitTasks:
			t.InitTask(initTask)
		case scheduledTask := <-t.channel.ScheduledTasks:
			t.UpdateTask(scheduledTask.TaskID) // 开始计时

		}
	}
}

// NewTracker 创建新的 Tracker
func NewTracker(channel *paradigm.RappaChannel) *Tracker {

	t := &Tracker{
		taskTracks:          make(map[paradigm.TaskHash]*paradigm.SynthTaskTrackItem),
		pendingCommitSlot:   make(map[paradigm.SlotHash]*paradigm.PendingCommitSlotTrack),
		channel:             channel,
		expireOutputChannel: make(chan paradigm.ExpireItem, 100), // todo @SD 配置
		expireInputChannel:  make(chan paradigm.ExpireItem, 100), // todo @SD 配置
		eliminators:         make([]*Eliminator, 0),
		//taskBuckets:       make([][]string, 0),
		//slotBuckets:       make([][]paradigm.SlotHash, 0),
	}
	e := make([]*Eliminator, 0)
	// todo @SD配置 数量
	for i := 0; i < 2; i++ {
		e = append(e, NewEliminator(t.expireInputChannel, t.expireOutputChannel))
	}
	t.eliminators = e
	return t
}
