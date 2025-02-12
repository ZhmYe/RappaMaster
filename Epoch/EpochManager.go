package Epoch

import (
	"BHLayer2Node/LogWriter"
	"BHLayer2Node/Tracker"
	"BHLayer2Node/paradigm"
	pb "BHLayer2Node/pb/service"
	"fmt"
	"sync"
)

type EpochRecord = paradigm.EpochRecord
type Task = paradigm.Task

type EpochManager struct {
	//tasks             map[string]*Task 不需要task，task放到Oracle里
	mu      sync.Mutex
	tracker *Tracker.Tracker // 监视任务和slot的expire
	channel *paradigm.RappaChannel
	//pendingCommitSlot map[paradigm.SlotHash]*paradigm.PendingCommitSlotTrack // 等待由Justified -> Finalized的slot
	currentEpoch int
	epochRecord  *EpochRecord
}

func (t *EpochManager) Start() {
	go t.tracker.Start()
	processTasks := func() {
		for {
			select {
			case <-t.channel.EpochEvent:
				t.UpdateEpoch() // 如果epoch更新，那么先更新epoch，此时有新的任务也会进入下一个epoch
			//case initTask := <-t.channel.InitTasks:
			//	t.tracker.InitTask(initTask)
			//	// TODO @YZM 应该先上链再更新
			//	go func(track paradigm.SynthTaskTrackItem) {
			//		//t.channel.UnprocessedTasks <- slot
			//		// 上链新Task
			//		task := paradigm.NewTask(track.TaskID, track.Model, track.Params, track.Size)
			//		t.channel.PendingTransactions <- &paradigm.InitTaskTransaction{
			//			task,
			//		}
			//	}(*initTask)
			case commitSlotItem := <-t.channel.CommitSlots:
				// 判断类别,如果是新commit的则commit，如果已经通过投票，则finalize
				switch commitSlotItem.State() {
				case paradigm.UNDETERMINED:
					//t.pendingCommitSlot[commitSlotItem.SlotHash()] = paradigm.NewPendingCommitSlotTrack(&commitSlotItem, t.CheckIsReliable(commitSlotItem.Sign)) // 等待verify
					// JUSITIFIED的slot必须在一段时间内完成可信证明
					t.tracker.UpdateSlot(commitSlotItem)
					t.epochRecord.Commit(&commitSlotItem) // 无论如何都要放到commit里，用于投票
				case paradigm.JUSTIFIED:
					// 这里的JUSTIFIED只是说明通过投票了，在无需可信证明的情况下，可以上链
					// 这里直接commit，commit里不需要额外的check,随时可以上链

					// 接下来只需要将那个对应的pending给设置为win vote 剩下的由Tracker自己处理
					//t.pendingCommitSlot[commitSlotItem.SlotHash()].ReceiveProof()
					// TODO @YZM 这里slot的提交时间要理一下
					//if _, exist := t.pendingCommitSlot[commitSlotItem.SlotHash()]; exist {
					//	t.pendingCommitSlot[commitSlotItem.SlotHash()].WonVote()
					//}
					//t.epochRecord.Justified(commitSlotItem)
					t.tracker.WonVote(commitSlotItem.SlotHash())
				case paradigm.FINALIZE:
					commitSlotItem.SetEpoch(int32(t.currentEpoch)) // 统一都设置这个epoch
					commitSlotItem.SetFinalize()
					err := t.tracker.Commit(&commitSlotItem) // 正式更新任务
					if err != nil {
						LogWriter.Log("ERROR", err.Error())
						continue
					}
					t.epochRecord.Finalize(&commitSlotItem)
					// 上链任务推进情况
					go func(transaction *paradigm.TaskProcessTransaction) {
						t.channel.PendingTransactions <- transaction
					}(&paradigm.TaskProcessTransaction{
						CommitSlotItem: &commitSlotItem,
						Proof:          nil,
						Signatures:     nil,
					})
				case paradigm.INVALID:
					t.epochRecord.Abort(&commitSlotItem, commitSlotItem.InvalidType) // 如果在外面就判断出来不对，直接加入到invalid即可
				default:
					panic("An Unknown State CommitSlotItem should not be involved in commitSlot!!!")
				}

				//t.UpdateTask(initTask.Sign, initTask.Model, initTask.Size, initTask.Params)
			//case schedule := <-t.channel.ScheduledTasks:
			//	_, err := t.UpdateTaskSchedule(schedule)
			//	// 不合法
			//	if err != nil {
			//		LogWriter.Log("ERROR", err.Error())
			//		continue
			//	}
			//	// 将任务添加到对应剩余时间的桶,这里只记录sign即可
			//	t.tracker.UpdateTask(schedule.Sign)
			default:
				continue
			}
		}
	}
	go processTasks()
}

//func (t *EpochManager) CheckTaskIsFinish(sign string) bool {
//	if task, exist := t.tasks[sign]; !exist {
//		return false
//	} else {
//		return task.IsFinish()
//	}
//}
//

//func (t *EpochManager) CheckSlotIsValid(sign string, slot int32) paradigm.InvalidCommitType {
//	if _, exist := t.tasks[sign]; !exist {
//		return paradigm.INVALID_SLOT
//	} // todo 这里可以区分slot和sign
//	task := t.tasks[sign]
//	if task.Slot < slot {
//		return paradigm.INVALID_SLOT
//	}
//	if task.Slot > slot {
//		return paradigm.EXPIRE_SLOT
//	}
//	return paradigm.NONE
//}

//func (t *EpochManager) UpdateTask(sign string, model paradigm.SupportModelType, size int32, params map[string]interface{}) {
//	if _, exist := t.tasks[sign]; !exist {
//		task := paradigm.NewTask(sign, model, params, size)
//		t.tasks[sign] = task
//		nextSlot, _ := task.Next()
//		go func(slot paradigm.UnprocessedTask) {
//			t.channel.UnprocessedTasks <- slot
//			// 上链新Task
//			t.channel.PendingTransactions <- &paradigm.InitTaskTransaction{
//				task,
//			}
//		}(nextSlot)
//		LogWriter.Log("TRACKER", fmt.Sprintf("Update New Epoch, sign: %s, slot: 0", sign))
//	}
//}
//
//func (t *EpochManager) UpdateTaskSchedule(schedule paradigm.TaskSchedule) (bool, error) {
//	sign, slot := schedule.Sign, schedule.Slot
//	//if _, exist := t.tasks[sign]; !exist {
//	//	//t.tasks[sign] = NewTask(sign, schedule.Model, schedule.Params, schedule.Size)
//	//	LogWriter.Log("ERROR", fmt.Sprintf("Epoch %s does not exist", sign))
//	//}
//	t.UpdateTask(sign, schedule.Model, schedule.Size, schedule.Params)
//	task := t.tasks[sign]
//	if t.CheckSlotIsValid(sign, slot) != paradigm.NONE {
//		return false, fmt.Errorf("invalid slot")
//	}
//	err := task.UpdateSchedule(schedule) // 更新slot
//	if err != nil {
//		return false, err
//	}
//	return true, nil
//}

//func (t *EpochManager) Commit(slot *paradigm.CommitSlotItem) error {
//	task, exist := t.tasks[slot.Sign]
//	if !exist {
//		return fmt.Errorf("task %s does not exist", slot.Sign)
//	}
//	return task.Commit(slot)
//}

func (t *EpochManager) UpdateEpoch() {
	t.currentEpoch++
	LogWriter.Log("TRACKER", fmt.Sprintf("Epoch update, current Epoch: %d", t.currentEpoch))
	//finalizedSlots, abortSlots := t.tracker.OutOfDate()
	//validTaskMap := make(map[string]int32)
	//for _, commitSlotItem := range finalizedSlots {
	//	commitSlotItem.SetEpoch(int32(t.currentEpoch)) // 统一都设置这个epoch
	//	commitSlotItem.SetFinalize()
	//	err := t.tracker.Commit(commitSlotItem) // 正式更新任务
	//	if err != nil {
	//		LogWriter.Log("ERROR", err.Error())
	//		continue
	//	}
	//	t.epochRecord.Finalize(commitSlotItem)
	//	// 上链任务推进情况
	//	go func(transaction *paradigm.TaskProcessTransaction) {
	//		t.channel.PendingTransactions <- transaction
	//	}(&paradigm.TaskProcessTransaction{
	//		CommitSlotItem: commitSlotItem,
	//		Proof:          nil,
	//		Signatures:     nil,
	//	})
	//}
	//for _, slot := range abortSlots {
	//	//slot.SetEpoch(t.currentEpoch)
	//	t.epochRecord.Abort(slot, paradigm.VERIFIED_FAILED)
	//}

	// 更新epoch的时候，构建心跳
	heartbeat := t.buildHeartbeat()
	t.channel.EpochHeartbeat <- heartbeat
	tmp := *t.epochRecord
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
		// TODO @YZM
		t.channel.PendingTransactions <- &paradigm.EpochRecordTransaction{
			EpochRecord:   &epochRecord,
			Id:            int32(t.currentEpoch),
			CommitsHash:   commits,
			JustifiedHash: finalized,
			Invalids:      epochRecord.Invalids,
		}
	}(tmp)
	// 下面的内容和心跳无关
	t.epochRecord.Echo()
	//for _, sign := range outOfDateTasks {
	//	task := t.tasks[sign]
	//	if t.CheckTaskIsFinish(sign) {
	//		LogWriter.Log("TRACKER", fmt.Sprintf("Epoch %s finished at slot %d, expected: %d, processed: %d", sign, task.Slot, task.Size, task.Process))
	//		continue
	//	}
	//	nextSlot, _ := task.Next()
	//	//validTaskMap[nextSlot.Sign] = int32(nextSlot.Slot)
	//	go func(slot paradigm.UnprocessedTask) {
	//		t.channel.UnprocessedTasks <- slot
	//	}(nextSlot)
	//}
	t.epochRecord.Refresh()

}
func (t *EpochManager) buildHeartbeat() *pb.HeartbeatRequest {
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
func NewEpochManager(channel *paradigm.RappaChannel) *EpochManager {
	//config := channel.Config
	return &EpochManager{
		channel: channel,
		//tasks:             make(map[string]*Task),
		mu:          sync.Mutex{},
		tracker:     Tracker.NewTracker(channel),
		epochRecord: paradigm.NewEpochRecord(),
		//pendingCommitSlot: make(map[paradigm.SlotHash]*paradigm.PendingCommitSlotTrack),
		currentEpoch: -1,
	}
}
