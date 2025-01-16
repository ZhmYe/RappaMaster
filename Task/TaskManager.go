package Task

import (
	"BHLayer2Node/LogWriter"
	"BHLayer2Node/paradigm"
	pb "BHLayer2Node/pb/service"
	"fmt"
	"sync"
)

type EpochRecord = paradigm.EpochRecord
type Task = paradigm.Task

type TaskManager struct {
	tasks             map[string]*Task
	mu                sync.Mutex
	tracker           *Tracker
	channel           *paradigm.RappaChannel
	pendingCommitSlot map[paradigm.SlotHash]*paradigm.PendingCommitSlotTrack // 等待由Justified -> Finalized的slot
	//scheduledTasks    chan paradigm.TaskSchedule
	// 这里我们假定，正确节点需要在分发所有的数据块并抱枕所有数据块都落实后才发送commit
	// 但恶意节点可以故意发送commit消息，因此我们不能直接commit，需要等待一轮投票
	//commitSlot chan paradigm.CommitSlotItem // 这里是单纯的commit上来的justified或者finalize的
	//finalizeSlot        chan paradigm.CommitSlotItem // 这里是finalize的
	//unprocessedTasks    chan paradigm.UnprocessedTask
	//initTasks           chan paradigm.UnprocessedTask
	//pendingTransactions chan paradigm.Transaction
	//slotToVotes      chan paradigm.CommitSlotItem
	//epochChangeEvent chan bool // 外部触发的 epoch 更新信号
	currentEpoch int
	epochRecord  *EpochRecord
	//epochHeartbeat   chan *pb.HeartbeatRequest
}

func (t *TaskManager) Start() {
	processTasks := func() {
		for {
			select {
			case <-t.channel.EpochEvent:
				t.UpdateEpoch() // 如果epoch更新，那么先更新epoch，此时有新的任务也会进入下一个epoch
			case commitSlotItem := <-t.channel.CommitSlots: // 如果不需要更新epoch，那么优先commit或finalize
				// 判断类别,如果是新commit的则commit，如果已经通过投票，则finalize
				switch commitSlotItem.State() {
				case paradigm.UNDETERMINED:
					// 该slot刚被提交
					// 这里需要先确认一下这个slot是否是合法的, 如果这个slot已经是过时的了，没有必要进入投票
					//isValid := t.CheckSlotIsValid(commitSlotItem.Sign, commitSlotItem.Slot)
					//if isValid != paradigm.NONE {
					//	t.epochRecord.Abort(&commitSlotItem, isValid)
					//	continue
					//} // 不考虑过时
					// TODO @YZM 这里是没有判断重复的，是否可以这么考虑，如果节点给的数据多，那是好事，只要它两次提交的数据不是一样的，这部分判断在后面zkp相关里 先不管
					t.pendingCommitSlot[commitSlotItem.SlotHash()] = paradigm.NewPendingCommitSlotTrack(&commitSlotItem, t.CheckIsReliable(commitSlotItem.Sign)) // 等待verify
					// JUSITIFIED的slot必须在一段时间内完成存储和可信证明
					t.tracker.UpdateSlot(commitSlotItem.SlotHash())
					t.epochRecord.Commit(&commitSlotItem) // 无论如何都要放到commit里，用于投票
				case paradigm.JUSTIFIED:
					// 这里的JUSTIFIED只是说明通过投票了，在无需可信证明的情况下，可以上链
					// 这里直接commit，commit里不需要额外的check,随时可以上链

					// 接下来只需要将那个对应的pending给设置为win vote 剩下的由Tracker自己处理
					//t.pendingCommitSlot[commitSlotItem.SlotHash()].ReceiveProof()
					// TODO @YZM 这里slot的提交时间要理一下
					if _, exist := t.pendingCommitSlot[commitSlotItem.SlotHash()]; exist {
						t.pendingCommitSlot[commitSlotItem.SlotHash()].WonVote()
					}
					//err := t.Commit(commitSlotItem)
					//if err != nil {
					//	LogWriter.Log("ERROR", err.Error())
					//	continue
					//}
					//t.epochRecord.finalize(commitSlotItem)
					//t.pendingTransactions <- &paradigm.CommitSlotTransaction{
					//	CommitSlotItem: commitSlotItem,
					//	Epoch:          t.currentEpoch,
					//}
					//case paradigm.FINALIZE:
					// 这里的Finalize只是说明通过投票了，在无需可信证明的情况下，可以上链
					// 这里直接commit，commit里不需要额外的check,随时可以上链

					// 接下来只需要将那个对应的pending给设置为win vote 剩下的由Tracker自己处理
					//t.pendingCommitSlot[commitSlotItem.SlotHash()].ReceiveProof()
					//t.pendingCommitSlot[commitSlotItem.SlotHash()].WonVote()
				//err := t.Commit(commitSlotItem)
				//if err != nil {
				//	LogWriter.Log("ERROR", err.Error())
				//	continue
				//}
				//t.epochRecord.finalize(commitSlotItem)
				//t.pendingTransactions <- &paradigm.CommitSlotTransaction{
				//	CommitSlotItem: commitSlotItem,
				//	Epoch:          t.currentEpoch,
				//}
				case paradigm.INVALID:
					t.epochRecord.Abort(&commitSlotItem, commitSlotItem.InvalidType) // 如果在外面就判断出来不对，直接加入到invalid即可
				default:
					panic("An Unknown State CommitSlotItem should not be involved in commitSlot!!!")
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
			case initTask := <-t.channel.InitTasks:
				t.UpdateTask(initTask.Sign, initTask.Model, initTask.Size, initTask.Params)
			case schedule := <-t.channel.ScheduledTasks:
				_, err := t.UpdateTaskSchedule(schedule)
				// 不合法
				if err != nil {
					LogWriter.Log("ERROR", err.Error())
					continue
				}
				// 将任务添加到对应剩余时间的桶,这里只记录sign即可
				t.tracker.UpdateTask(schedule.Sign)
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
func (t *TaskManager) CheckSlotIsValid(sign string, slot int32) paradigm.InvalidCommitType {
	if _, exist := t.tasks[sign]; !exist {
		return paradigm.INVALID_SLOT
	} // todo 这里可以区分slot和sign
	task := t.tasks[sign]
	if task.Slot < slot {
		return paradigm.INVALID_SLOT
	}
	if task.Slot > slot {
		return paradigm.EXPIRE_SLOT
	}
	return paradigm.NONE
}
func (t *TaskManager) CheckIsReliable(sign string) bool {
	if _, exist := t.tasks[sign]; !exist {
		return false
	}
	task := t.tasks[sign]
	return task.IsReliable()
}
func (t *TaskManager) UpdateTask(sign string, model paradigm.SupportModelType, size int32, params map[string]interface{}) {
	if _, exist := t.tasks[sign]; !exist {
		task := paradigm.NewTask(sign, model, params, size)
		t.tasks[sign] = task
		nextSlot, _ := task.Next()
		go func(slot paradigm.UnprocessedTask) {
			t.channel.UnprocessedTasks <- slot
			// 上链新Task
			t.channel.PendingTransactions <- &paradigm.InitTaskTransaction{
				task,
			}
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
	if t.CheckSlotIsValid(sign, slot) != paradigm.NONE {
		return false, fmt.Errorf("invalid slot")
	}
	err := task.UpdateSchedule(schedule) // 更新slot
	if err != nil {
		return false, err
	}
	return true, nil
}
func (t *TaskManager) Commit(slot *paradigm.CommitSlotItem) error {
	task, exist := t.tasks[slot.Sign]
	if !exist {
		return fmt.Errorf("task %s does not exist", slot.Sign)
	}
	return task.Commit(slot)
}
func (t *TaskManager) UpdateEpoch() {
	t.currentEpoch++
	LogWriter.Log("TRACKER", fmt.Sprintf("Epoch update, current Epoch: %d", t.currentEpoch))
	outOfDateTasks, outOfDateCommitSlot := t.tracker.OutOfDate()
	//validTaskMap := make(map[string]int32)
	for _, h := range outOfDateCommitSlot {
		pendingSlot := t.pendingCommitSlot[h]
		if pendingSlot.Check() {
			// 这个commitSlot在指定时间内完成了存储任务(vote)和可信任务(zkp)
			commitSlotItem := pendingSlot.CommitSlotItem
			commitSlotItem.SetEpoch(int32(t.currentEpoch)) // 统一都设置这个epoch
			commitSlotItem.SetFinalize()
			err := t.Commit(commitSlotItem) // 正式更新任务
			if err != nil {
				LogWriter.Log("ERROR", err.Error())
				continue
			}
			t.epochRecord.Finalize(commitSlotItem)
			// 上链任务推进情况
			go func(transaction *paradigm.TaskProcessTransaction) {
				t.channel.PendingTransactions <- transaction
			}(&paradigm.TaskProcessTransaction{
				CommitSlotItem: commitSlotItem,
				Proof:          nil,
				Signatures:     nil,
			})
		} else {
			// 未在指定时间内完成，那么直接丢弃
			//slot.SetInvalid(paradigm.VERIFIED_FAILED)
			t.epochRecord.Abort(pendingSlot.CommitSlotItem, paradigm.VERIFIED_FAILED)
			// 这里会出现节点后面才额外提交zkp，但已经失效了，直接无视，也就是commitzkp(还没写)的时候发现没有这个任务，那么要么没有通过justified(这是commitSlot的前置，得到hash和seed)
			// 要么就是已经失效了，直接无视
		}
		delete(t.pendingCommitSlot, h) // 标记为已完成，不需要记录了
	}
	// 更新epoch的时候，构建心跳
	heartbeat := t.buildHeartbeat()
	t.channel.EpochHeartbeat <- heartbeat
	//fmt.Println(t.epochRecord)
	go func(epochRecord EpochRecord) {
		commits := make([]paradigm.SlotHash, 0)
		justified := make([]paradigm.SlotHash, 0)
		finalized := make([]paradigm.SlotHash, 0)
		//invalids := make(map[paradigm.SlotHash]paradigm.InvalidCommitType)
		for hash, _ := range epochRecord.Commits {
			commits = append(commits, hash)
		}
		for hash, _ := range epochRecord.Justifieds {
			justified = append(justified, hash)
		}
		for hash, _ := range epochRecord.Finalizes {
			finalized = append(finalized, hash)
		}
		// 上链epoch信息
		t.channel.PendingTransactions <- &paradigm.EpochRecordTransaction{
			EpochRecord:   &epochRecord,
			Id:            int32(t.currentEpoch),
			CommitsHash:   justified,
			JustifiedHash: finalized,
			Invalids:      epochRecord.Invalids,
		}
	}(*t.epochRecord)
	// 下面的内容和心跳无关
	t.epochRecord.Echo()
	for _, sign := range outOfDateTasks {
		task := t.tasks[sign]
		if t.CheckTaskIsFinish(sign) {
			LogWriter.Log("TRACKER", fmt.Sprintf("Task %s finished at slot %d, expected: %d, processed: %d", sign, task.Slot, task.Size, task.Process))
			continue
		}
		nextSlot, _ := task.Next()
		//validTaskMap[nextSlot.Sign] = int32(nextSlot.Slot)
		go func(slot paradigm.UnprocessedTask) {
			t.channel.UnprocessedTasks <- slot
		}(nextSlot)
	}
	t.epochRecord.Refresh()

}
func (t *TaskManager) buildHeartbeat() *pb.HeartbeatRequest {
	//fmt.Println(len(t.epochRecord.commits), len(t.epochRecord.finalizes), 111)
	return &pb.HeartbeatRequest{
		Commits:    t.epochRecord.Commits,
		Justifieds: t.epochRecord.Justifieds,
		Finalizes:  t.epochRecord.Finalizes,
		Invalids:   t.epochRecord.Invalids,
		//Tasks:     validTaskMap,
		Epoch: int32(t.currentEpoch),
	}
}
func NewTaskManager(channel *paradigm.RappaChannel) *TaskManager {
	config := channel.Config

	return &TaskManager{
		channel: channel,
		tasks:   make(map[string]*Task),
		mu:      sync.Mutex{},
		tracker: NewTracker(config),
		//scheduledTasks:      channel.ScheduledTasks,
		//commitSlot:          channel.CommitSlots,
		//unprocessedTasks:    channel.UnprocessedTasks,
		//epochChangeEvent:    channel.EpochEvent,
		//initTasks:           channel.InitTasks,
		//pendingTransactions: channel.PendingTransactions,
		//epochHeartbeat:      channel.EpochHeartbeat,
		//slotToVotes:      slotToVotes,
		epochRecord:       paradigm.NewEpochRecord(),
		pendingCommitSlot: make(map[paradigm.SlotHash]*paradigm.PendingCommitSlotTrack),
		currentEpoch:      -1,
	}
}
