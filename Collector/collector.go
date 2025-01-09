package Collector

import (
	"BHLayer2Node/LogWriter"
	"BHLayer2Node/paradigm"
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
func (c *Collector) processSlotUpdate() {
	for slot := range c.channel.ToCollectorSlotChannel {
		collectSlotItem := paradigm.CollectSlotItem{
			Sign: slot.Sign,
			Hash: slot.SlotHash(),
			Size: slot.Process,
		}
		taskSlotsList := c.taskSlots[slot.Sign]
		taskSlotsList = append(taskSlotsList, collectSlotItem)
		c.taskSlots[slot.Sign] = taskSlotsList
		LogWriter.Log("COLLECT", fmt.Sprintf("Collector Update Slot %s to Task %s", slot.SlotHash(), slot.Sign))
	}
}

func (c *Collector) Start() {
	go c.processSlotUpdate()
}

func NewCollector(channel *paradigm.RappaChannel) *Collector {
	return &Collector{
		taskSlots: make(map[string][]paradigm.CollectSlotItem),
		channel:   channel,
	}
}
