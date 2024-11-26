package Monitor

import "BHCoordinator/Config"

// Monitor 监视节点状态
type Monitor struct {
	config Config.BHCoordinatorConfig
}

// Advice todo 这里传入nIDs以及数据量，考虑如何分配一个slot里的数据
func (m *Monitor) Advice(nIDs []int, size int) []int {
	result := make([]int, 0)
	for len(result) < len(nIDs) {
		result = append(result, size/len(nIDs))
	}
	result[0] += size % len(nIDs)
	return result
}
