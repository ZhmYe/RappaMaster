package Oracle

import (
	"BHLayer2Node/LogWriter"
	"BHLayer2Node/paradigm"
	"fmt"
)

type Oracle struct {
	// 这里记录所有的历史记录
	// 链上无需维护复杂的可追溯结构
	// 在这里设计一个可追溯结构，每个元素附带txHash即可
	channel  *paradigm.RappaChannel
	tasks    map[string]*paradigm.Task // 所有的任务, 以sign为map
	slotsMap map[paradigm.SlotHash]*paradigm.Slot
	epochs   map[int]*paradigm.DevEpoch // 所有的epoch,以epochID为ma
	tID      int
	// todo @YZM 这里加上transaction receipt，便于查看
}

func (d *Oracle) Start() {
	processTransactions := func() {
		for ptxs := range d.channel.DevTransactionChannel {
			for _, ptx := range ptxs {
				ptx.SetID(d.tID)      // 在这里统一设置交易id
				transaction := ptx.Tx // 一笔交易，根据交易类型判断更新什么
				switch transaction.(type) {
				case *paradigm.EpochRecordTransaction:
					// 上链了一个epoch历史记录
					// 那么需要新建一个epoch
					epoch := paradigm.NewDevEpoch(ptx)
					if _, exist := d.epochs[epoch.EpochRecord.Id]; exist {
						panic("Error in epoch count!!!")
					}
					d.epochs[epoch.EpochRecord.Id] = epoch // 记录epoch
					// 遍历epoch中的invalid，用于更新状态
					for slotHash, e := range epoch.Invalids {
						slot := d.slotsMap[slotHash]
						slot.SetError(paradigm.InvalidCommitTypeToString(e))
					}
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
					//d.tasks = append(d.tasks, task) // 这里更新任务列表
					//d.transactions[ptx.Id] = task   // 这里直接映射到相应的任务
				case *paradigm.TaskProcessTransaction:
					// 上链了一笔任务推进交易
					commitRecord := paradigm.NewCommitRecord(ptx)
					// 这里就是要更新某个task
					taskSign := commitRecord.Sign
					if task, exist := d.tasks[taskSign]; exist {
						err := task.Commit(commitRecord) // 将commitSlot添加到task的对应slotRecord中
						if err != nil {
							panic(err)
						}
						if _, exist := d.slotsMap[commitRecord.SlotHash()]; !exist {
							// 如果没有这个slot，那么说明可能是schedule运行较慢，这里采用更多轮几次
							d.channel.DevTransactionChannel <- []*paradigm.PackedTransaction{ptx}
							continue
						}
						slot := d.slotsMap[commitRecord.SlotHash()]
						// 更新slot状态，这里应该是指针
						slot.Commit(commitRecord.CommitSlotItem) // 将slot提交
						if task.IsFinish() && !task.HasbeenCollect {
							d.channel.FakeCollectSignChannel <- [2]interface{}{task.Sign, task.Process}
							task.SetCollected()
							LogWriter.Log("DEBUG", fmt.Sprintf("In Oracle, Task %s finished, expected: %d, processed: %s", task.Sign, task.Size, task.Process))
							task.Print()
						}
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
						//task.Print()
					} else {
						panic("Error in Process Epoch!!!")
					}

				default:
					panic("Unknown Transaction!!!")
				}
				d.tID++
			}
		}
	}
	processSchedule := func() {
		for schedule := range d.channel.OracleSchedules {
			// 这些是完成的调度
			if _, exist := d.tasks[schedule.TaskID]; !exist {
				panic("Unknown Task!!!")
			}
			task := d.tasks[schedule.TaskID]
			task.UpdateSchedule(schedule) // 这里的schedule已经包含了grpc的状态
			d.tasks[schedule.TaskID] = task
			for _, slot := range schedule.Slots {
				d.slotsMap[slot.SlotID] = slot // 这里记录所有的slot，todo 其实只要记录processing的
			}
		}
	}
	//processSlot := func() {
	//	// 更新slot的状态
	//}
	go processTransactions()
	//go processSlot()
	processSchedule()
}

func NewOracle(channel *paradigm.RappaChannel) *Oracle {
	return &Oracle{
		channel:  channel,
		tasks:    make(map[string]*paradigm.Task),
		slotsMap: make(map[paradigm.SlotHash]*paradigm.Slot),
		epochs:   make(map[int]*paradigm.DevEpoch),
		//tx:                     channel.OracleTransactionChannel,
		//toCollectorSlotChannel: channel.ToCollectorSlotChannel,
		//taskFinishSignChannel:  channel.FakeCollectSignChannel,
		tID: 0,
	}
}
