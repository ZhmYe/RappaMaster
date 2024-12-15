package Monitor

import "BHLayer2Node/Config"

// Monitor 监视节点状态
type Monitor struct {
	config Config.BHLayer2NodeConfig
}

// Advice todo 这里传入nIDs以及数据量，考虑如何分配一个slot里的数据
func (m *Monitor) Advice(nIDs []int, size int32) []int32 {
	result := make([]int32, 0)
	for len(result) < len(nIDs) {
		result = append(result, size/int32(len(nIDs)))
	}
	result[0] += size % int32(len(nIDs))
	return result
}
