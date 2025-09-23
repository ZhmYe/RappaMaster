package schedule

import (
	"RappaMaster/types"
	"context"
	"slices"
	"sort"
	"sync"
)

const (
	MIN_SPEED_CAN_SCHEDULE int64 = 1000 // 1KB
)

// Monitor process the commit/justified slots and maintains a global status of all nodes
type Monitor struct {
	mu     sync.Mutex
	status []types.NodeStatus
}

func (m *Monitor) Start(ctx context.Context) {
}
func (m *Monitor) TopNodes() ([]types.NodeStatus, int64) {
	m.mu.Lock()
	toSort := slices.Clone(m.status) // so we can have a Unchanged order of m.nodeStatus
	m.mu.Unlock()
	MAX_NODE_NUMBER_TO_SCHEDULE := len(toSort) / 2 // todo
	// Inert sorting
	sort.Slice(toSort, func(i, j int) bool {
		if toSort[i].Check() && toSort[j].Check() {
			return toSort[i].Speed() >= toSort[j].Speed()
		} else {
			return toSort[i].Check()
		}
	})
	totalSpeed := int64(0)
	res := make([]types.NodeStatus, 0)
	for i := 0; i < MAX_NODE_NUMBER_TO_SCHEDULE; i++ {
		if !toSort[i].Check() || toSort[i].Speed() < MIN_SPEED_CAN_SCHEDULE {
			break
		}
		res = append(res, toSort[i])
		totalSpeed += toSort[i].Speed()
	}
	return res, totalSpeed
}
