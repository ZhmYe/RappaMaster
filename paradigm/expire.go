package paradigm

import (
	"time"
)

// ExpireItem 任务/Slot 需要实现的接口
type ExpireItem interface {
	IsExpire() bool        // 判断是否过期
	ExpireTime() time.Time // 获取过期时间
}
type BasicTimeExpire struct {
	expireTime time.Time
}

func (e *BasicTimeExpire) IsExpire() bool {
	return !time.Now().Before(e.expireTime)
}
func (e *BasicTimeExpire) ExpireTime() time.Time {
	return e.expireTime
}
func NewBasicTimeExpire(time time.Time) BasicTimeExpire {
	return BasicTimeExpire{expireTime: time}
}

// ExpireHeap 小顶堆实现
type ExpireHeap []ExpireItem

func (h ExpireHeap) Len() int           { return len(h) }
func (h ExpireHeap) Less(i, j int) bool { return h[i].ExpireTime().Before(h[j].ExpireTime()) }
func (h ExpireHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *ExpireHeap) Push(x interface{}) {
	*h = append(*h, x.(ExpireItem))
}
func (h *ExpireHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}

type ExpireTask struct {
	BasicTimeExpire
	TaskID TaskHash
}
type ExpireSlot struct {
	BasicTimeExpire
	SlotHash    SlotHash
	PendingSlot *PendingCommitSlotTrack
}
