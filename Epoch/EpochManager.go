package Epoch

import (
	"BHLayer2Node/PKI"
	"BHLayer2Node/Recovery"
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
	pkiManager   *PKI.PKIManager
}

func (t *EpochManager) Start() {
	go t.tracker.Start()
	processTasks := func() {
		for {
			select {
			case <-t.channel.EpochEvent:
				t.UpdateEpoch() // 如果epoch更新，那么先更新epoch，此时有新的任务也会进入下一个epoch
			case commitSlotItem := <-t.channel.CommitSlots:
				// 判断类别,如果是新commit的则commit，如果已经通过投票，则finalize
				switch commitSlotItem.State() {
				case paradigm.UNDETERMINED:
					//t.pendingCommitSlot[commitSlotItem.SlotHash()] = paradigm.NewPendingCommitSlotTrack(&commitSlotItem, t.CheckIsReliable(commitSlotItem.Sign)) // 等待verify
					// JUSITIFIED的slot必须在一段时间内完成可信证明
					//验证节点ca是否有效
					if t.pkiManager.VerifyNodeCA(commitSlotItem.Nid, int32(t.currentEpoch), commitSlotItem.GetNodeCA()) {
						t.tracker.UpdateSlot(commitSlotItem)
						t.epochRecord.Commit(&commitSlotItem) // 无论如何都要放到commit里，用于投票
					} else {
						// 无效直接抛弃,按理说要在链上验证签名，这里简单本地验证
						// TODO 这里可能说明这个节点ca过期了，需要重新注册，目前没相关逻辑
						commitSlotItem.SetInvalid(paradigm.VERIFIED_FAILED)
						t.epochRecord.Invalids[commitSlotItem.SlotHash()] =
							paradigm.NewInvalidCommitError(paradigm.VERIFIED_FAILED, fmt.Sprintf("the slot from %d has invalid ca", commitSlotItem.Nid))
					}
				case paradigm.JUSTIFIED:
					// 这里的JUSTIFIED只是说明通过投票了，在无需可信证明的情况下，可以上链
					// 这里直接commit，commit里不需要额外的check,随时可以上链
					// 接下来只需要将那个对应的pending给设置为win vote 剩下的由Tracker自己处理
					t.tracker.WonVote(commitSlotItem.SlotHash())
				case paradigm.FINALIZE:
					commitSlotItem.SetEpoch(int32(t.currentEpoch)) // 统一都设置这个epoch
					commitSlotItem.SetFinalize()
					err := t.tracker.Commit(&commitSlotItem) // 正式更新任务
					if err != nil {
						//paradigm.Error(Runt, err.Error())
						continue
					}
					t.epochRecord.Finalize(&commitSlotItem)
					// 上链任务推进情况
					//TODO 这里简单上链签名和ca证书
					go func(transaction *paradigm.TaskProcessTransaction) {
						t.channel.PendingTransactions <- transaction
					}(&paradigm.TaskProcessTransaction{
						CommitSlotItem: commitSlotItem.CommitSlotItem,
						Proof:          nil,
						Signatures:     [][]byte{[]byte(commitSlotItem.GetSlotSignature()), []byte(commitSlotItem.GetNodeCA())},
					})
				case paradigm.INVALID:
					t.epochRecord.Abort(&commitSlotItem, commitSlotItem.InvalidType) // 如果在外面就判断出来不对，直接加入到invalid即可
				default:
					panic("An Unknown State CommitSlotItem should not be involved in commitSlot!!!")
				}
			case task := <-t.channel.EpochInitTaskChannel:
				// 记录epoch定义的task
				t.epochRecord.UpdateTask(task)
			case slot := <-t.channel.UnScheduledSlotChannel:
				// 这里要记录没法调度的slot
				t.epochRecord.Invalids[slot.SlotID] = paradigm.NewInvalidCommitError(paradigm.INVALID_SLOT, slot.ErrorMessage())
			default:
				continue
			}
		}
	}
	go processTasks()
}

func (t *EpochManager) UpdateEpoch() {
	t.mu.Lock()
	t.currentEpoch++
	currentEpoch := t.currentEpoch
	t.mu.Unlock()

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
			Id:            int32(currentEpoch),
			CommitsHash:   commits,
			JustifiedHash: finalized,
			Invalids:      epochRecord.SampleInvalids(),
		}
	}(tmp)
	// 下面的内容和心跳无关
	t.epochRecord.Echo()
	t.epochRecord.Refresh()

}
func (t *EpochManager) buildHeartbeat() *pb.HeartbeatRequest {
	//fmt.Println(len(t.epochRecord.commits), len(t.epochRecord.finalizes), 111)
	return &pb.HeartbeatRequest{
		Commits:    t.epochRecord.Commits,
		Justifieds: t.epochRecord.Justifieds,
		Finalizes:  t.epochRecord.Finalizes,
		Invalids:   t.epochRecord.SampleInvalids(),
		//Tasks:     validTaskMap,
		Epoch: int32(t.currentEpoch),
	}
}
func NewEpochManager(channel *paradigm.RappaChannel, recovery *Recovery.RappaRecovery, manager *PKI.PKIManager) *EpochManager {
	return &EpochManager{
		channel: channel,
		//tasks:             make(map[string]*Task),
		mu:          sync.Mutex{},
		tracker:     Tracker.NewTracker(channel),
		epochRecord: paradigm.NewEpochRecord(int(recovery.EpochID) + 1),
		pkiManager:  manager,
		//pendingCommitSlot: make(map[paradigm.SlotHash]*paradigm.PendingCommitSlotTrack),
		// currentEpoch: -1,
		currentEpoch: int(recovery.EpochID),
	}
}
