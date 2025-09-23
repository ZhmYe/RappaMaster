package Epoch

import (
	"RappaMaster/config"
	"RappaMaster/helper"
	"context"
	"time"
)

// EpochManager advances new epochs, when a timeout is exceeded
// It create a new epoch into db, and get the information of the last epoch to generate a heartbeat
type EpochManager struct {
	ticker *time.Ticker
}

func (em *EpochManager) advanceNextEpoch(ctx context.Context) {
	for {
		select {
		case <-em.ticker.C:
			err := helper.GlobalServiceHelper.DB.AdvanceEpoch()
			if err != nil {
				helper.GlobalServiceHelper.ReportError(err) // todo panic?
			}
		case <-ctx.Done():
			return

		}
	}
}

// todo 这里要在更新next epoch的同时，维护出上一个epoch的一些关键信息（比如以每个node为group，得到每个node在这个epoch里commit的slot，然后计算merkle（是否可以在代码中维护）
//func (em *EpochManager) processLastEpoch(ctx context.Context)

func (em *EpochManager) Start(ctx context.Context) {
	defer em.ticker.Stop()
	go em.advanceNextEpoch(ctx)

}

func NewEpochManager(config config.ComponentConfig) *EpochManager {
	return &EpochManager{
		ticker: time.NewTicker(config.EpochTimeDuration),
	}
}

//
//func (t *EpochManager) Start() {
//	go t.tracker.Start()
//	processTasks := func() {
//		for {
//			select {
//			case <-t.channel.EpochEvent:
//
//			case commitSlotItem := <-t.channel.CommitSlots:
//				// 判断类别,如果是新commit的则commit，如果已经通过投票，则finalize
//				switch commitSlotItem.State() {
//				case paradigm.UNDETERMINED:
//					//t.pendingCommitSlot[commitSlotItem.SlotHash()] = paradigm.NewPendingCommitSlotTrack(&commitSlotItem, t.CheckIsReliable(commitSlotItem.Sign)) // 等待verify
//					// JUSITIFIED的slot必须在一段时间内完成可信证明
//					t.tracker.UpdateSlot(commitSlotItem)
//					t.epochRecord.Commit(&commitSlotItem) // 无论如何都要放到commit里，用于投票
//				case paradigm.JUSTIFIED:
//					// 这里的JUSTIFIED只是说明通过投票了，在无需可信证明的情况下，可以上链
//					// 这里直接commit，commit里不需要额外的check,随时可以上链
//
//					// 接下来只需要将那个对应的pending给设置为win vote 剩下的由Tracker自己处理
//					//t.pendingCommitSlot[commitSlotItem.SlotHash()].ReceiveProof()
//					// TODO @YZM 这里slot的提交时间要理一下
//					//if _, exist := t.pendingCommitSlot[commitSlotItem.SlotHash()]; exist {
//					//	t.pendingCommitSlot[commitSlotItem.SlotHash()].WonVote()
//					//}
//					//t.epochRecord.Justified(commitSlotItem)
//					t.tracker.WonVote(commitSlotItem.SlotHash())
//				case paradigm.FINALIZE:
//					commitSlotItem.SetEpoch(int32(t.currentEpoch)) // 统一都设置这个epoch
//					commitSlotItem.SetFinalize()
//					err := t.tracker.Commit(&commitSlotItem) // 正式更新任务
//					if err != nil {
//						//paradigm.Error(Runt, err.Error())
//						continue
//					}
//					t.epochRecord.Finalize(&commitSlotItem)
//					// 上链任务推进情况
//					go func(transaction *transaction.TaskProcessTransaction) {
//						t.channel.PendingTransactions <- transaction
//					}(&transaction.TaskProcessTransaction{
//						CommitSlotItem: &commitSlotItem,
//						Proof:          nil,
//						Signatures:     nil,
//					})
//				case paradigm.INVALID:
//					t.epochRecord.Abort(&commitSlotItem, commitSlotItem.InvalidType) // 如果在外面就判断出来不对，直接加入到invalid即可
//				default:
//					panic("An Unknown State CommitSlotItem should not be involved in commitSlot!!!")
//				}
//
//				//t.UpdateTask(initTask.Sign, initTask.Model, initTask.Size, initTask.Params)
//			//case schedule := <-t.channel.ScheduledTasks:
//			//	_, err := t.UpdateTaskSchedule(schedule)
//			//	// 不合法
//			//	if err != nil {
//			//		LogWriter.Log("ERROR", err.Error())
//			//		continue
//			//	}
//			//	// 将任务添加到对应剩余时间的桶,这里只记录sign即可
//			//	t.tracker.UpdateTask(schedule.Sign)
//			default:
//				continue
//			}
//		}
//	}
//	go processTasks()
//}
//
//func (t *EpochManager) UpdateEpoch() {
//	t.mu.Lock()
//	t.currentEpoch++
//	currentEpoch := t.currentEpoch
//	t.mu.Unlock()
//
//	// 更新epoch的时候，构建心跳
//	heartbeat := t.buildHeartbeat()
//	t.channel.EpochHeartbeat <- heartbeat
//	tmp := *t.epochRecord
//	go func(epochRecord EpochRecord) {
//		commits := make([]paradigm.SlotHash, 0)
//		justified := make([]paradigm.SlotHash, 0)
//		finalized := make([]paradigm.SlotHash, 0)
//		//invalids := make(map[paradigm.SlotHash]paradigm.InvalidCommitType)
//		for hash, _ := range epochRecord.Commits {
//			commits = append(commits, hash)
//		}
//		for hash, _ := range epochRecord.Justifieds {
//			justified = append(justified, hash)
//		}
//		for hash, _ := range epochRecord.Finalizes {
//			finalized = append(finalized, hash)
//		}
//		// 上链epoch信息
//		// TODO @YZM
//		t.channel.PendingTransactions <- &transaction.EpochRecordTransaction{
//			EpochRecord:   &epochRecord,
//			Id:            int32(currentEpoch),
//			CommitsHash:   commits,
//			JustifiedHash: finalized,
//			Invalids:      epochRecord.Invalids,
//		}
//	}(tmp)
//	// 下面的内容和心跳无关
//	t.epochRecord.Echo()
//	//for _, sign := range outOfDateTasks {
//	//	task := t.tasks[sign]
//	//	if t.CheckTaskIsFinish(sign) {
//	//		LogWriter.Log("TRACKER", fmt.Sprintf("Epoch %s finished at slot %d, expected: %d, processed: %d", sign, task.Slot, task.Size, task.Process))
//	//		continue
//	//	}
//	//	nextSlot, _ := task.Next()
//	//	//validTaskMap[nextSlot.Sign] = int32(nextSlot.Slot)
//	//	go func(slot paradigm.UnprocessedTask) {
//	//		t.channel.UnprocessedTasks <- slot
//	//	}(nextSlot)
//	//}
//	t.epochRecord.Refresh()
//
//}
//func (t *EpochManager) buildHeartbeat() *pb.HeartbeatRequest {
//	//fmt.Println(len(t.epochRecord.commits), len(t.epochRecord.finalizes), 111)
//	return &pb.HeartbeatRequest{
//		Commits:    t.epochRecord.Commits,
//		Justifieds: t.epochRecord.Justifieds,
//		Finalizes:  t.epochRecord.Finalizes,
//		Invalids:   t.epochRecord.Invalids,
//		//Tasks:     validTaskMap,
//		Epoch: int32(t.currentEpoch),
//	}
//}

//func NewEpochManager(channel *helper.RappaChannel, recovery *Recovery.RappaRecovery) *EpochManager {
//	return &EpochManager{
//		channel: channel,
//		//tasks:             make(map[string]*Task),
//		mu:          sync.Mutex{},
//		tracker:     Tracker.NewTracker(channel),
//		epochRecord: paradigm.NewEpochRecord(int(recovery.EpochID) + 1),
//		//pendingCommitSlot: make(map[paradigm.SlotHash]*paradigm.PendingCommitSlotTrack),
//		// currentEpoch: -1,
//		currentEpoch: int(recovery.EpochID),
//	}
//}
