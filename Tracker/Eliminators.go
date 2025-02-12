package Tracker

import (
	"BHLayer2Node/paradigm"
	"container/heap"
)

// Eliminator 负责定期检查过期任务
type Eliminator struct {
	expireHeap          paradigm.ExpireHeap      // 任务小顶堆
	expireOutputChannel chan paradigm.ExpireItem // 过期任务的输出通道
	expireInputChannel  chan paradigm.ExpireItem // 任务输入通道
}

// NewEliminators 创建 Eliminators
func NewEliminator(input chan paradigm.ExpireItem, output chan paradigm.ExpireItem) *Eliminator {
	e := &Eliminator{
		expireHeap:          make(paradigm.ExpireHeap, 0),
		expireOutputChannel: output,
		expireInputChannel:  input,
	}
	heap.Init(&e.expireHeap)
	return e
}

// Start 任务调度，每次进一出一，避免卡顿
func (e *Eliminator) Start() {
	for {
		// 1. 取出堆顶元素（如果有）
		if len(e.expireHeap) > 0 {
			item := heap.Pop(&e.expireHeap).(paradigm.ExpireItem)
			if item.IsExpire() {
				e.expireOutputChannel <- item // 过期任务输出
			} else {
				heap.Push(&e.expireHeap, item) // 未过期重新入队
			}
		}

		// 2. 尝试从 `expireInputChannel` 取出新任务（非阻塞）
		select {
		case newItem := <-e.expireInputChannel:
			heap.Push(&e.expireHeap, newItem) // 添加新任务
		default:
			// 没有新任务，不做任何操作，避免阻塞
		}
	}
}
