package Oracle

import (
	"BHLayer2Node/Database"
	"BHLayer2Node/paradigm"
	"fmt"
	"time"
)

type PersistedOracle struct {
	channel *paradigm.RappaChannel
	// mySQLConfig paradigm.DBConnection              //mysql连接配置
	dbService  *Database.DatabaseService
	collectors map[string]paradigm.RappaCollector //定义任务收集器
}

func (o *PersistedOracle) Start() {
	updateOracle := func() {
		for {
			select {
			case schedule := <-o.channel.OracleSchedules:
				// 这些是完成的调度
				o.dbService.UpdateScheduleInTask(schedule)
				for _, slot := range schedule.Slots {
					o.dbService.UpdateSlotFromSchedule(slot)
					//d.slotsMap[slot.SlotID] = slot // 这里记录所有的slot，todo 其实只要记录processing的
				}
				o.channel.MonitorOracleChannel <- schedule // 传递给monitor，更新节点未完成的任务

			case ptxs := <-o.channel.DevTransactionChannel:
				for _, ptx := range ptxs {
					ptx.SetUpchainTime(time.Now()) // 在这里统一设置上链时间
					dateRecord := o.dbService.GetDateRecord(ptx.UpchainTime)
					transaction := ptx.Tx // 一笔交易，根据交易类型判断更新什么
					switch transaction.(type) {
					case *paradigm.EpochRecordTransaction:
						// 上链了一个epoch历史记录
						// 那么需要新建一个epoch
						epochRecord := ptx.Tx.Blob().(*paradigm.EpochRecord)
						//fmt.Println(222, epochRecord)
						commits, justifieds, finalizeds := make(map[paradigm.SupportModelType][]paradigm.SlotHash, 0), make(map[paradigm.SupportModelType][]paradigm.SlotHash, 0), make(map[paradigm.SupportModelType][]paradigm.SlotHash, 0)
						invalids := make([]*paradigm.Slot, 0)
						initTasks := make([]*paradigm.Task, 0)

						for slotHash, _ := range epochRecord.Commits {
							//slot := d.getSlot(slotHash)
							slot := o.dbService.GetSlot(slotHash)
							tempTask, err := o.dbService.GetTask(slot.TaskID)
							if err != nil {
								paradigm.Error(paradigm.RuntimeError, fmt.Sprintf("task not found of %s", slot.TaskID))
							}
							slotType := tempTask.Model
							if value, ok := commits[slotType]; ok {
								commits[slotType] = append(value, slotHash)
							} else {
								commitOfType := make([]paradigm.SlotHash, 0)
								commitOfType = append(commitOfType, slotHash)
								commits[slotType] = commitOfType
							}
						}

						for slotHash, _ := range epochRecord.Justifieds {
							slot := o.dbService.GetSlot(slotHash)
							tempTask, err := o.dbService.GetTask(slot.TaskID)
							if err != nil {
								paradigm.Error(paradigm.RuntimeError, fmt.Sprintf("task not found of %s", slot.TaskID))
							}
							slotType := tempTask.Model
							if value, ok := justifieds[slotType]; ok {
								justifieds[slotType] = append(value, slotHash)
							} else {
								justifiedOfType := make([]paradigm.SlotHash, 0)
								justifiedOfType = append(justifiedOfType, slotHash)
								justifieds[slotType] = justifiedOfType
							}
						}

						epochProcess := make(map[paradigm.SupportModelType]int32)

						for slotHash, _ := range epochRecord.Finalizes {
							slot := o.dbService.GetSlot(slotHash)
							tempTask, err := o.dbService.GetTask(slot.TaskID)
							if err != nil {
								paradigm.Error(paradigm.RuntimeError, fmt.Sprintf("task not found of %s", slot.TaskID))
							}
							slotType := tempTask.Model
							if value, ok := finalizeds[slotType]; ok {
								finalizeds[slotType] = append(value, slotHash)
							} else {
								finalizedOfType := make([]paradigm.SlotHash, 0)
								finalizedOfType = append(finalizedOfType, slotHash)
								finalizeds[slotType] = finalizedOfType
							}
							if value, ok := epochProcess[slotType]; ok {
								epochProcess[slotType] = value + slot.ScheduleSize
							} else {
								epochProcess[slotType] = slot.ScheduleSize
							}

						}

						for slotHash, e := range epochRecord.Invalids {
							o.dbService.SetSlotError(slotHash, e, int32(epochRecord.Id))
							invalids = append(invalids, o.dbService.GetSlot(slotHash))
						}
						for taskID, _ := range epochRecord.Tasks {
							if task, err := o.dbService.GetTask(taskID); err != nil {
								paradigm.Error(paradigm.RuntimeError, "Task has not been init")
							} else {
								initTasks = append(initTasks, task)
								ref, err := o.dbService.GetTransaction(task.TxReceipt.TransactionHash)
								if err != nil {
									paradigm.Error(paradigm.RuntimeError, "TX has not been init")
								}
								ref.EpochID = int32(epochRecord.Id)
							}
						}

						// 更新txMap,对应的rf是epochTx
						reference := &paradigm.DevReference{
							TxHash:      ptx.Receipt.TransactionHash,
							TxReceipt:   *ptx.Receipt,
							TxBlockHash: ptx.BlockHash,
							Rf:          paradigm.EpochTx,
							TaskID:      "",
							EpochID:     int32(epochRecord.Id),
							UpchainTime: ptx.UpchainTime,
							//ScheduleID: -1,
						}
						o.dbService.SetTransaction(reference)

						epoch := &paradigm.DevEpoch{
							EpochID:    int32(epochRecord.Id),
							Process:    epochProcess,
							Commits:    commits,
							Justifieds: justifieds,
							Finalizes:  finalizeds,
							Invalids:   invalids,
							InitTasks:  initTasks,
							TID:        reference.TID,
							// TxReceipt:  ptx.Receipt,
							// // TxBlock:    ptx.Block,
							// TxBlockHash: ptx.BlockHash,
						}
						// 记录epoch
						err := o.dbService.SetEpoch(epoch)
						if err != nil {
							paradigm.Error(paradigm.RuntimeError, "Failed to set epoch")
						}
						dateRecord.UpdateTransactions(1)
						o.dbService.UpdateDateRecord(dateRecord)
					case *paradigm.InitTaskTransaction:
						// 上链了一笔初始化任务的交易
						// 那么需要在tasks更新一个任务
						task := transaction.(*paradigm.InitTaskTransaction).Task
						//task.UpdateTxInfo(ptx)
						o.channel.InitTasks <- task.InitTrack() // 上链后，发起新的任务，这样scheduler能接受到
						// 更新txMap，对应的rf是InitTaskTx
						reference := &paradigm.DevReference{
							TxHash:      ptx.Receipt.TransactionHash,
							TxReceipt:   *ptx.Receipt,
							TxBlockHash: ptx.BlockHash,
							Rf:          paradigm.InitTaskTx,
							TaskID:      task.Sign,
							EpochID:     -1, // 这里的epochID需要等epoch更新才能拿到
							UpchainTime: ptx.UpchainTime,
							//ScheduleID: -1,
						}
						o.dbService.SetTransaction(reference)
						task.TID = reference.TID
						o.dbService.SetTask(task)
						o.collectors[task.Sign] = task.Collector

						// 更新date
						dateRecord.UpdateInitTasks(1)
						dateRecord.UpdateDateset(task.GetDataset())
						dateRecord.UpdateTransactions(1)
						o.dbService.UpdateDateRecord(dateRecord)
					case *paradigm.TaskProcessTransaction:
						// 上链了一笔任务推进交易
						commitRecord := paradigm.NewCommitRecord(ptx)
						// 这里就是要更新某个task
						taskSign := commitRecord.Sign
						if task, err := o.dbService.GetTask(taskSign); err == nil {
							// 传递给monitor更新完成的任务
							// TODO 这里暂时将任务的类型写入
							task.Collector = o.collectors[task.Sign]
							transaction.(*paradigm.TaskProcessTransaction).Model = task.Model
							o.channel.MonitorOracleChannel <- transaction
							reference := &paradigm.DevReference{
								TxHash:      ptx.Receipt.TransactionHash,
								TxReceipt:   *ptx.Receipt,
								TxBlockHash: ptx.BlockHash,
								Rf:          paradigm.SlotTX,
								TaskID:      commitRecord.Sign,
								EpochID:     commitRecord.Epoch,
								UpchainTime: ptx.UpchainTime,
								//ScheduleID: slot.ScheduleID,
							}
							o.dbService.SetTransaction(reference)
							o.dbService.SetSlotFinish(commitRecord.SlotHash(), commitRecord.CommitSlotItem)
							commitRecord.TxID = reference.TID
							err := task.Commit(commitRecord) // 将commitSlot添加到task的对应slotRecord中
							if err != nil {
								e := paradigm.Error(paradigm.RuntimeError, err.Error())
								panic(e.Error())
							}

							if task.IsFinish() && !task.HasbeenCollect {
								// 更新date
								dateRecord.UpdateFinishTasks(1)
								task.SetEndTime()
								//d.channel.FakeCollectSignChannel <- [2]interface{}{task.Sign, task.Process}
								task.SetCollected()
								paradigm.Print("INFO", fmt.Sprintf("Task %s finished, expected: %d, processed: %d", task.Sign, task.Size, task.Process))
							}
							// 这里更新oracle的全局信息
							//d.nbFinalized++ // 又完成了一个finalized
							//if value, ok := d.synthData[task.Model]; ok {
							//	d.synthData[task.Model] = value + commitRecord.Process
							//} else {
							//	d.synthData[task.Model] = commitRecord.Process
							//}
							//更新task
							o.dbService.UpdateTask(task)
							dateRecord.UpdateFinalized(1)
							dateRecord.UpdateProcess(commitRecord.Process, task.Model)
							dateRecord.UpdateTransactions(1)
							o.dbService.UpdateDateRecord(dateRecord)
							//task.Print()
						} else {
							e := paradigm.Error(paradigm.RuntimeError, "Task not in Oracle")
							panic(e.Error())
						}

					default:
						e := paradigm.Error(paradigm.RuntimeError, "Unknown Transaction!!!")
						panic(e.Error())
					}

				}
				//// 更新latest
				//d.latestTxs = append(d.latestTxs, ptxs...)
				//// TODO @SD 这个20可以考虑设置成参数，也可以不设置
				//if len(d.latestTxs) > 20 {
				//	d.latestTxs = d.latestTxs[len(d.latestTxs)-20:]
				//}
			}

		}
	}
	// 处理Query

	go o.processDBQuery()
	updateOracle()
}

func NewPersistedOracle(channel *paradigm.RappaChannel, service *Database.DatabaseService) *PersistedOracle {
	return &PersistedOracle{
		channel:    channel,
		dbService:  service,
		collectors: make(map[string]paradigm.RappaCollector),
	}
}
