package Dev

import (
	"BHLayer2Node/paradigm"
)

type Dev struct {
	// 这里记录所有的历史记录
	// 链上无需维护复杂的可追溯结构
	// 在这里设计一个可追溯结构，每个元素附带txHash即可
	tasks  map[string]*paradigm.DevTask // 所有的任务, 以sign为map
	epochs map[int]*paradigm.DevEpoch   // 所有的epoch,以epochID为map
	//updateEpoch  chan *paradigm.EpochRecord        // 更新epoch来传递内容，由chainupper给定txhash
	//transactions map[int]interface{}               // 这里一笔交易对应一个task或者一个slot或者一个epoch
	tx                     chan []*paradigm.PackedTransaction // 上链完成后的交易,在异步上链组件中批量给出
	toCollectorSlotChannel chan paradigm.CommitSlotItem       // 传递给collector
	tID                    int
	// todo @YZM 这里加上transaction receipt，便于查看
}

func (d *Dev) Start() {
	for ptxs := range d.tx {
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
				//d.transactions[]
			case *paradigm.InitTaskTransaction:
				// 上链了一笔初始化任务的交易
				// 那么需要在tasks更新一个任务
				task := paradigm.NewDevTask(ptx)
				if _, exist := d.tasks[task.Task.Sign]; exist {
					panic("Error in init Task!!!")
				}
				d.tasks[task.Task.Sign] = task
				//d.tasks = append(d.tasks, task) // 这里更新任务列表
				//d.transactions[ptx.Id] = task   // 这里直接映射到相应的任务
			case *paradigm.TaskProcessTransaction:
				// 上链了一笔任务推进交易
				slot := paradigm.NewCommitRecord(ptx)
				// 这里就是要更新某个task
				taskSign := slot.Sign
				if task, exist := d.tasks[taskSign]; exist {
					task.UpdateCommitSlot(slot) // 将commitSlot添加到task的对应slotRecord中
					//task.Print()
				} else {
					panic("Error in Process Task!!!")
				}
				// 这里更新了task的slot，那么可以将这里的Slot传递给collector
				commitSlotItem := transaction.(*paradigm.TaskProcessTransaction).CommitSlotItem
				d.toCollectorSlotChannel <- *commitSlotItem

			default:
				panic("Unknown Transaction!!!")
			}
			d.tID++
		}
	}
}

func NewDev(channel *paradigm.RappaChannel) *Dev {
	return &Dev{
		tasks:                  make(map[string]*paradigm.DevTask),
		epochs:                 make(map[int]*paradigm.DevEpoch),
		tx:                     channel.DevTransactionChannel,
		toCollectorSlotChannel: channel.ToCollectorSlotChannel,
		tID:                    0,
	}
}
