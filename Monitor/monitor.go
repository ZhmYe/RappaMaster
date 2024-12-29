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
		// todo @YZM 这里先简单定下范围
		adviceSize := size / int32(len(nIDs))
		if adviceSize > 10 {
			adviceSize = 10
		}
		if adviceSize == 0 {
			adviceSize = 1
		}
		result = append(result, adviceSize)
	}
	result[0] += size % int32(len(nIDs))
	return result
}
