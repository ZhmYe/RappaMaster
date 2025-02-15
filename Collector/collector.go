package Collector

import (
	"BHLayer2Node/paradigm"
	"BHLayer2Node/pb/service"
	"fmt"
)

// Collector 作为Task的一部分，收集所有的当前任务的提交的slot
// 以process的大小进行排列，用于使用贪心爬山算法来给出最小数量的文件符合要求
type Collector struct {
	taskID paradigm.TaskHash
	items  []paradigm.CollectSlotItem
	//taskSlots map[string][]paradigm.CollectSlotItem
	channel    *paradigm.RappaChannel
	outputType paradigm.ModelOutputType
}

// ProcessSlotUpdate 处理slot的更新，从channel中获取
// 这里后续要根据size来提取相应的数据块，通过贪心算法每次选择最大的，这里提前先对每个slot进行有序排列，每次使用二分，O(log_n)
func (c *Collector) ProcessSlotUpdate(slot paradigm.CollectSlotItem) {
	if slot.Sign != c.taskID {
		paradigm.Error(paradigm.RuntimeError, fmt.Sprintf("Slot's Sign does not match TaskID in collector, %s != %s", slot.Sign, c.taskID))
	}
	//for slot := range c.channel.ToCollectorSlotChannel {
	//collectSlotItem := paradigm.CollectSlotItem{
	//	Sign: slot.Sign,
	//	Hash: slot.SlotHash(),
	//	Size: slot.Process,
	//}
	// 取出taskSlots list，二分查找首个size小于等于的
	//taskSlotsList := c.taskSlots[slot.Sign]
	//taskSlotsList = append(taskSlotsList, collectSlotItem)
	// 找到index s.t. t[index].size >= slot.size >= t[index + 1].size
	binarySearch := func() int {
		// 找到插入的位置
		left, right := 0, len(c.items)-1
		index := -1
		for left <= right {
			mid := (left + right) / 2
			if c.items[mid].Size >= slot.Size {
				// 找到了一个符合条件的，要找到的是最小的，往右边找
				index = mid
				left = mid + 1
			} else {
				// 往左边走
				right = mid - 1
			}
		}
		return index
	}
	currentLength := len(c.items)
	index := binarySearch()
	if index == -1 {
		// 说明全部都比它小
		c.items = append([]paradigm.CollectSlotItem{slot}, c.items...)
	} else {
		c.items = append(c.items[:index+1], append([]paradigm.CollectSlotItem{slot}, c.items[index+1:]...)...)
	}
	if len(c.items) != currentLength+1 {
		panic("error in insert...")
	}

	paradigm.Log("COLLECT", fmt.Sprintf("Collector Update Slot %s to Epoch %s", slot.Hash, slot.Sign))
	//}
}

func (c *Collector) ProcessCollect(collectRequest paradigm.HttpCollectRequest) (interface{}, error) {

	sign := collectRequest.Sign
	if sign != c.taskID {
		e := paradigm.Error(paradigm.RuntimeError, fmt.Sprintf("Collect Request's Sign does not match TaskID, %s != %s", sign, c.taskID))
		return nil, fmt.Errorf(e.Error())
	}
	total := collectRequest.Size
	//mission := collectRequest.Mission
	remain := collectRequest.Size
	// 取出有序的taskSlot
	// 贪心
	var slotList []paradigm.CollectSlotItem
	for _, slot := range c.items {
		if remain > 0 {
			remain -= slot.Size
			slotList = append(slotList, slot)
		} else {
			break
		}
	}
	paradigm.Print("COLLECT", fmt.Sprintf("Start Collect Task %s, Size: %d, Slot to Collect: %d", sign, total, len(slotList)))
	// 得到如果要收齐这个collect要求，可以对slotHashList里的slot进行collect
	// 这里为了不妨碍slot的更新，通过go func开始异步
	collectInstance := CollectSlotInstance{
		//Mission:         mission,
		OutputType: c.outputType,
		Slots:      slotList,
		//Transfer:        collectRequest.TransferChannel,
		ResponseChannel: make(chan service.RecoverResponse, paradigm.DefaultBHLayer2NodeConfig.MaxCommitSlotItemPoolSize), // TODO
		//Connection:      c.channel.SlotCollectChannel,
		Channel: c.channel,
	}
	return collectInstance.Collect(), nil
}

//func (c *Collector) Start() {
//	for {
//		select {
//		case request := <-c.channel.ToCollectorRequestChannel:
//			c.ProcessCollect(request)
//		case slot := <-c.channel.ToCollectorSlotChannel:
//			c.ProcessSlotUpdate(slot)
//		default:
//			continue
//		}
//	}
//}

func NewCollector(taskID paradigm.TaskHash, channel *paradigm.RappaChannel) *Collector {
	return &Collector{
		taskID:  taskID,
		items:   make([]paradigm.CollectSlotItem, 0),
		channel: channel,
	}
}
