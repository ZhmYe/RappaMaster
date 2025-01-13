package Collector

import (
	"BHLayer2Node/Config"
	"BHLayer2Node/LogWriter"
	"BHLayer2Node/paradigm"
	"BHLayer2Node/pb/service"
	"fmt"
)

// Collector 根据前端的请求，开始收集以sign为标识的所有Slot_hash
// 因此这里需要能获取某个task的所有slot_hash
// todo @YZM dev的功能还需要扩大，这里暂时就写成dev会把一部分数据传递给collector保存
type Collector struct {
	taskSlots map[string][]paradigm.CollectSlotItem
	channel   *paradigm.RappaChannel
}

// processSlotUpdate 处理slot的更新，从channel中获取
// 这里后续要根据size来提取相应的数据块，通过贪心算法每次选择最大的，这里提前先对每个slot进行有序排列，每次使用二分，O(log_n)
func (c *Collector) processSlotUpdate(slot paradigm.CommitSlotItem) {
	//for slot := range c.channel.ToCollectorSlotChannel {
	collectSlotItem := paradigm.CollectSlotItem{
		Sign: slot.Sign,
		Hash: slot.SlotHash(),
		Size: slot.Process,
	}
	// 取出taskSlots list，二分查找首个size小于等于的
	taskSlotsList := c.taskSlots[slot.Sign]
	//taskSlotsList = append(taskSlotsList, collectSlotItem)
	// 找到index s.t. t[index].size >= slot.size >= t[index + 1].size
	binarySearch := func() int {
		// 找到插入的位置
		left, right := 0, len(taskSlotsList)-1
		index := -1
		for left <= right {
			mid := (left + right) / 2
			if taskSlotsList[mid].Size >= collectSlotItem.Size {
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
	currentLength := len(taskSlotsList)
	index := binarySearch()
	if index == -1 {
		// 说明全部都比它小
		taskSlotsList = append([]paradigm.CollectSlotItem{collectSlotItem}, taskSlotsList...)
	} else {
		taskSlotsList = append(taskSlotsList[:index+1], append([]paradigm.CollectSlotItem{collectSlotItem}, taskSlotsList[index+1:]...)...)
	}
	newLength := len(taskSlotsList)
	if newLength != currentLength+1 {
		panic("error in insert...")
	}

	c.taskSlots[slot.Sign] = taskSlotsList
	LogWriter.Log("COLLECT", fmt.Sprintf("Collector Update Slot %s to Task %s", slot.SlotHash(), slot.Sign))
	//}
}

func (c *Collector) processCollect(collectRequest paradigm.CollectRequest) {

	sign := collectRequest.Sign
	total := collectRequest.Size
	mission := collectRequest.Mission
	remain := collectRequest.Size
	// 取出有序的taskSlot
	taskSlot := c.taskSlots[sign]
	// 贪心
	var slotHashList []paradigm.SlotHash
	for _, slot := range taskSlot {
		if remain > 0 {
			remain -= slot.Size
			slotHashList = append(slotHashList, slot.Hash)
		} else {
			break
		}
	}
	LogWriter.Log("COLLECT", fmt.Sprintf("Start Collect Task %s, Size: %d, Slot to Collect: %v", sign, total, slotHashList))
	// 得到如果要收齐这个collect要求，可以对slotHashList里的slot进行collect
	// 这里为了不妨碍slot的更新，通过go func开始异步
	collectInstance := paradigm.CollectSlotInstance{
		Mission:         mission,
		SlotHashs:       slotHashList,
		Transfer:        collectRequest.TransferChannel,
		ResponseChannel: make(chan service.RecoverResponse, Config.DefaultBHLayer2NodeConfig.MaxCommitSlotItemPoolSize), // TODO
		Connection:      c.channel.SlotCollectChannel,
	}
	go collectInstance.Collect()
}

func (c *Collector) Start() {
	for {
		select {
		case request := <-c.channel.ToCollectorRequestChannel:
			c.processCollect(request)
		case slot := <-c.channel.ToCollectorSlotChannel:
			c.processSlotUpdate(slot)
		default:
			continue
		}
	}
}

func NewCollector(channel *paradigm.RappaChannel) *Collector {
	return &Collector{
		taskSlots: make(map[string][]paradigm.CollectSlotItem),
		channel:   channel,
	}
}
