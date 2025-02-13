package Oracle

import (
	"BHLayer2Node/LogWriter"
	"BHLayer2Node/paradigm"
	"fmt"
)

// Oracle 提供前端页面的直接查询
/***
1. 存证溯源页面
	1.1. 合成任务溯源：从任务ID或交易哈希来追踪任务
		这里对所有任务存了map，通过TaskID可以获取任务，为了提供交易哈希因此需要有一个交易哈希到任务的映射
		此外，前端需要提供任务在Epoch的信息: 1.每个epoch的合成总量; 2. 每个epoch提交了哪些，异常了哪些，finalized了哪些
		因此需要在Task里维护一个epoch结构体
		此外前端需要展示schedule的完成情况：因此需要有一个总的统计：完成了多少，失败了多少，处理中多少
	1.2. Epoch溯源，从EpochID或交易哈希来追踪Epoch
		这里对所有epoch存了map，同上需要交易哈希到epoch的映射
		TODO
		(1) epoch里应该会展示invalid、finalized、commit的饼状图(len即可)
		(2) 不同任务的推进情况，这里需要一个图表
		(3) 不同Slot的信息，这个和task部分差不多，不过不是schedule，分三个数组展示
		(4) 左下角是Heartbeat


***/
type Oracle struct {
	// 这里记录所有的历史记录
	// 链上无需维护复杂的可追溯结构
	// 在这里设计一个可追溯结构，每个元素附带txHash即可
	channel  *paradigm.RappaChannel
	tasks    map[string]*paradigm.Task // 所有的任务, 以sign为map
	slotsMap map[paradigm.SlotHash]*paradigm.Slot
	//slotsMap sync.Map
	epochs       map[int32]*paradigm.DevEpoch // 所有的epoch,以epochID为ma
	tID          int
	txMap        map[string]paradigm.DevReference // 这里以txHash作为key,交易最终会指向一个task或者epoch
	latestTxs    []*paradigm.PackedTransaction    // 展示最新的20笔交易
	latestEpochs []*paradigm.DevEpoch             // 展示最新的20个epoch的信息：commit, justified, finalized,txHash, data
	synthData    int32                            // 合成总量
	nbFinalized  int32                            //提交总量，这里指Finalized
	//latestEpoch int32                            // 最新的epoch，这里要保证Epoch一定是连续的合法 TODO
	//slotsMaoMutex sync.RWMutex
}

func (d *Oracle) UpdateSlotFromSchedule(slot *paradigm.Slot) {
	if _, exist := d.slotsMap[slot.SlotID]; !exist {
		d.slotsMap[slot.SlotID] = slot
	} else {
		d.slotsMap[slot.SlotID].UpdateSchedule(slot.ScheduleID, slot.TaskID, slot.ScheduleSize)
	}
}
func (d *Oracle) SetSlotError(slotHash paradigm.SlotHash, e paradigm.InvalidCommitType, epoch int32) {
	slot := d.GetSlot(slotHash)
	//slot.CommitSlot.SetEpoch(epoch)
	slot.SetEpoch(epoch)
	slot.SetError(paradigm.InvalidCommitTypeToString(e))
}
func (d *Oracle) SetSlotFinish(slotHash paradigm.SlotHash, commitSlotItem *paradigm.CommitSlotItem) {
	slot := d.GetSlot(slotHash)
	// 更新slot状态，这里应该是指针
	slot.Commit(commitSlotItem) // 将slot提交

}
func (d *Oracle) GetSlot(slotHash paradigm.SlotHash) *paradigm.Slot {
	if _, exist := d.slotsMap[slotHash]; !exist {
		d.slotsMap[slotHash] = paradigm.NewSlotWithSlotID(slotHash)
	}
	slot := d.slotsMap[slotHash]
	return slot
}
func (d *Oracle) Start() {
	updateOracle := func() {
		for {
			select {
			case schedule := <-d.channel.OracleSchedules:
				// 这些是完成的调度
				if _, exist := d.tasks[schedule.TaskID]; !exist {
					panic("Unknown Task!!!")
				}
				task := d.tasks[schedule.TaskID]
				task.UpdateSchedule(schedule) // 这里的schedule已经包含了grpc的状态
				d.tasks[schedule.TaskID] = task
				for _, slot := range schedule.Slots {
					d.UpdateSlotFromSchedule(slot)
					//d.slotsMap[slot.SlotID] = slot // 这里记录所有的slot，todo 其实只要记录processing的
				}

			case ptxs := <-d.channel.DevTransactionChannel:
				for _, ptx := range ptxs {
					ptx.SetID(d.tID)      // 在这里统一设置交易id
					transaction := ptx.Tx // 一笔交易，根据交易类型判断更新什么
					switch transaction.(type) {
					case *paradigm.EpochRecordTransaction:
						// 上链了一个epoch历史记录
						// 那么需要新建一个epoch
						epochRecord := ptx.Tx.Blob().(*paradigm.EpochRecord)
						//fmt.Println(222, epochRecord)
						commits, justifieds, finalizeds := make([]*paradigm.Slot, 0), make([]*paradigm.Slot, 0), make([]*paradigm.Slot, 0)
						invalids := make([]*paradigm.Slot, 0)
						initTasks := make([]*paradigm.Task, 0)
						for slotHash, _ := range epochRecord.Commits {
							//slot := d.GetSlot(slotHash)
							commits = append(commits, d.GetSlot(slotHash))
						}
						for slotHash, _ := range epochRecord.Justifieds {
							justifieds = append(justifieds, d.GetSlot(slotHash))
						}
						for slotHash, _ := range epochRecord.Finalizes {
							finalizeds = append(finalizeds, d.GetSlot(slotHash))
						}
						for slotHash, e := range epochRecord.Invalids {
							d.SetSlotError(slotHash, e, int32(epochRecord.Id))
							invalids = append(invalids, d.GetSlot(slotHash))
						}
						for taskID, _ := range epochRecord.Tasks {
							if task, exist := d.tasks[taskID]; !exist {
								paradigm.RaiseError(paradigm.RuntimeError, "Task has not been init", false)
							} else {
								initTasks = append(initTasks, task)
								ref := d.txMap[task.TxReceipt.TransactionHash]
								ref.EpochID = int32(epochRecord.Id)
							}
						}
						epoch := &paradigm.DevEpoch{
							EpochID:    int32(epochRecord.Id),
							Process:    epochRecord.Process,
							Commits:    commits,
							Justifieds: justifieds,
							Finalizes:  finalizeds,
							Invalids:   invalids,
							InitTasks:  initTasks,
							TxReceipt:  ptx.Receipt,
							TxID:       ptx.Id,
							// TxBlock:    ptx.Block,
							TxBlockHash: ptx.BlockHash,
						}
						//epoch := paradigm.NewDevEpoch(ptx)
						if _, exist := d.epochs[epoch.EpochID]; exist {
							paradigm.RaiseError(paradigm.RuntimeError, "Error in EpochUpdate", false)
						}
						d.epochs[epoch.EpochID] = epoch // 记录epoch
						// 更新latest
						d.latestEpochs = append(d.latestEpochs, epoch)
						// TODO @SD 这个20可以考虑设置成参数，也可以不设置
						if len(d.latestEpochs) > 20 {
							d.latestEpochs = d.latestEpochs[len(d.latestEpochs)-20:]
						}

						//遍历epoch中的invalid，用于更新状态
						//for slotHash, e := range epoch.Invalids {
						//	d.SetSlotError(slotHash, e, int32(epoch.EpochRecord.Id))
						//slot, _ := d.slotsMap[slotHash]
						//slot.CommitSlot.SetEpoch(int32(epoch.EpochRecord.Id))
						//slot.SetError(paradigm.InvalidCommitTypeToString(e))
						//}
						// 更新txMap,对应的rf是epochTx
						reference := paradigm.DevReference{
							TxHash:      ptx.Receipt.TransactionHash,
							TxReceipt:   *ptx.Receipt,
							TxBlockHash: ptx.BlockHash,
							Rf:          paradigm.EpochTx,
							TaskID:      "",
							EpochID:     int32(epochRecord.Id),
							//ScheduleID: -1,
						}
						d.txMap[ptx.Receipt.TransactionHash] = reference
						//d.transactions[]
					case *paradigm.InitTaskTransaction:
						// 上链了一笔初始化任务的交易
						// 那么需要在tasks更新一个任务
						task := transaction.(*paradigm.InitTaskTransaction).Task
						task.UpdateTxInfo(ptx)
						if _, exist := d.tasks[task.Sign]; exist {
							panic("Error in init Task!!!")
						}
						d.tasks[task.Sign] = task
						d.channel.InitTasks <- task.InitTrack() // 上链后，发起新的任务，这样scheduler能接受到
						// 更新txMap，对应的rf是InitTaskTx
						reference := paradigm.DevReference{
							TxHash:      ptx.Receipt.TransactionHash,
							TxReceipt:   *ptx.Receipt,
							TxBlockHash: ptx.BlockHash,
							Rf:          paradigm.InitTaskTx,
							TaskID:      task.Sign,
							EpochID:     -1, // 这里的epochID需要等epoch更新才能拿到
							//ScheduleID: -1,
						}
						d.txMap[ptx.Receipt.TransactionHash] = reference

					case *paradigm.TaskProcessTransaction:
						// 上链了一笔任务推进交易
						commitRecord := paradigm.NewCommitRecord(ptx)
						// 这里就是要更新某个task
						taskSign := commitRecord.Sign
						if task, exist := d.tasks[taskSign]; exist {
							d.SetSlotFinish(commitRecord.SlotHash(), commitRecord.CommitSlotItem)
							err := task.Commit(commitRecord) // 将commitSlot添加到task的对应slotRecord中
							if err != nil {
								panic(err)
							}
							if task.IsFinish() && !task.HasbeenCollect {
								d.channel.FakeCollectSignChannel <- [2]interface{}{task.Sign, task.Process}
								task.SetCollected()
								LogWriter.Log("DEBUG", fmt.Sprintf("In Oracle, Task %s finished, expected: %d, processed: %d", task.Sign, task.Size, task.Process))
								task.Print()
								//LogWriter.Log("DEBUG", "Test Query Generation...")
								//query := NewEvidencePreserveTaskIDQuery(map[interface{}]interface{}{"taskID": task.Sign})
								//d.channel.QueryChannel <- query
								//go func() {
								//	response := query.ReceiveResponse()
								//	fmt.Println(response.ToHttpJson(), response.Error())
								//}()

								//query := Query.NewEvidencePreserveEpochIDQuery(map[interface{}]interface{}{"epochID": 8})
								//d.channel.QueryChannel <- query
								//go func() {
								//	response := query.ReceiveResponse()
								//	fmt.Println(response.ToHttpJson(), response.Error())
								//}()
								//query := new(EvidencePreserveTaskTxQuery)
								//query.ParseRawDataFromHttpEngine(map[interface{}]interface{}{"txHash": task.TxReceipt.TransactionHash})

								//response := query.GenerateResponse(task)
								//fmt.Println(response)
							}
							// 这里更新oracle的全局信息
							d.nbFinalized++                     // 又完成了一个finalized
							d.synthData += commitRecord.Process // 合成数据
							// 这里更新了task的slot，那么可以将这里的Slot传递给collector
							commitSlotItem := transaction.(*paradigm.TaskProcessTransaction).CommitSlotItem
							collectSlotItem := paradigm.CollectSlotItem{
								Sign:        commitSlotItem.Sign,
								Hash:        commitSlotItem.SlotHash(),
								Size:        commitSlotItem.Process,
								OutputType:  task.OutputType,
								PaddingSize: commitSlotItem.Padding,
								StoreMethod: commitSlotItem.Store,
							}
							d.channel.ToCollectorSlotChannel <- collectSlotItem
							// 更新reference
							reference := paradigm.DevReference{
								TxHash:      ptx.Receipt.TransactionHash,
								TxReceipt:   *ptx.Receipt,
								TxBlockHash: ptx.BlockHash,
								Rf:          paradigm.SlotTX,
								TaskID:      commitRecord.Sign,
								EpochID:     commitRecord.Epoch,
								//ScheduleID: slot.ScheduleID,
							}
							d.txMap[ptx.Receipt.TransactionHash] = reference
							//task.Print()
						} else {
							panic("Error in Process Epoch!!!")
						}

					default:
						panic("Unknown Transaction!!!")
					}
					d.tID++

				}
				// 更新latest
				d.latestTxs = append(d.latestTxs, ptxs...)
				// TODO @SD 这个20可以考虑设置成参数，也可以不设置
				if len(d.latestTxs) > 20 {
					d.latestTxs = d.latestTxs[len(d.latestTxs)-20:]
				}
			}

		}
	}
	// 处理Query

	go d.processQuery()
	updateOracle()
}

func NewOracle(channel *paradigm.RappaChannel) *Oracle {
	return &Oracle{
		channel:      channel,
		tasks:        make(map[string]*paradigm.Task),
		slotsMap:     make(map[paradigm.SlotHash]*paradigm.Slot),
		epochs:       make(map[int32]*paradigm.DevEpoch),
		txMap:        map[string]paradigm.DevReference{},
		latestEpochs: make([]*paradigm.DevEpoch, 0),
		latestTxs:    make([]*paradigm.PackedTransaction, 0),
		synthData:    0,
		nbFinalized:  0,

		//tx:                     channel.OracleTransactionChannel,
		//toCollectorSlotChannel: channel.ToCollectorSlotChannel,
		//taskFinishSignChannel:  channel.FakeCollectSignChannel,
		tID: 0,
	}
}
