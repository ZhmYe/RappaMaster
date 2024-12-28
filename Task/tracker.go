package Task

import (
	"BHLayer2Node/Config"
	"BHLayer2Node/paradigm"
	"sync"
)

// Tracker 追踪任务状态
type Tracker struct {
	config        *Config.BHLayer2NodeConfig
	taskBuckets   [][]string            // 二维切片，每个索引对应剩余时间，这里是task slot的过期时间，也就是过期要开启新的slot
	slotBuckets   [][]paradigm.SlotHash // 二维切片，每个索引对应剩余时间，这里是一个commit slot的过期时间，要求一个commit slot必须1. 携带zkp; 2. 通过vote
	maxEpochDelay int
	mu            sync.Mutex
}

// UpdateTask 更新一个任务，已有的任务会根据OutOfDate()不断更新
func (t *Tracker) UpdateTask(sign string) {
	remainingEpoch := t.maxEpochDelay
	//if remainingEpoch >= len(t.taskBuckets) {
	t.expandBuckets(remainingEpoch)
	//}
	t.taskBuckets[remainingEpoch] = append(t.taskBuckets[remainingEpoch], sign)
}

// UpdateSlot 更新一个slot
func (t *Tracker) UpdateSlot(hash paradigm.SlotHash) {
	remainingEpoch := t.maxEpochDelay
	//if remainingEpoch >= len(t.slotBuckets) {
	t.expandBuckets(remainingEpoch)
	//}
	t.slotBuckets[remainingEpoch] = append(t.slotBuckets[remainingEpoch], hash)
}

func (t *Tracker) OutOfDate() ([]string, []paradigm.SlotHash) {
	processExpireTask := func() []string {
		if len(t.taskBuckets) == 0 {
			return []string{}
		}
		outOfDateTaskSlot := t.taskBuckets[0]
		t.taskBuckets = t.taskBuckets[1:]
		return outOfDateTaskSlot
	}
	processExpireSlot := func() []paradigm.SlotHash {
		if len(t.slotBuckets) == 0 {
			return []paradigm.SlotHash{}
		}
		outOfDateCommitSlot := t.slotBuckets[0]
		t.slotBuckets = t.slotBuckets[1:]
		return outOfDateCommitSlot
	}
	return processExpireTask(), processExpireSlot()

}

// expandBuckets 扩展桶的数量
func (t *Tracker) expandBuckets(required int) {
	for len(t.taskBuckets) <= required {
		t.taskBuckets = append(t.taskBuckets, []string{})
	}
	for len(t.slotBuckets) <= required {
		t.slotBuckets = append(t.slotBuckets, []paradigm.SlotHash{})
	}
}

// Setup 初始化 Tracker
func (t *Tracker) Setup(config *Config.BHLayer2NodeConfig) {
	t.config = config
	t.taskBuckets = make([][]string, 0)
	t.maxEpochDelay = config.MaxEpochDelay
}

// NewTracker 创建新的 Tracker
func NewTracker(config *Config.BHLayer2NodeConfig) *Tracker {
	tracker := &Tracker{}
	tracker.Setup(config)
	return tracker
}
