package Tracker

import "BHLayer2node/Config"

// Tracker 收到来自合成节点的心跳后，grpcEngine会将需要的信息传递给tracker
// tracker会统计所有节点的历史信息，用于向scheduler说明节点权重
// todo
type Tracker struct {
	config Config.BHLayer2NodeConfig
}

// todo 这里传入nIDs以及数据量，考虑如何给出一个slot的数据量
func (t *Tracker) Advice(nIDs []int, size int) int {
	if size < t.config.DefaultSlotSize {
		return size
	}
	return t.config.DefaultSlotSize
}

func (t *Tracker) Setup(config Config.BHLayer2NodeConfig) {
	t.config = config
}
func NewTracker(config Config.BHLayer2NodeConfig) *Tracker {
	return &Tracker{config: config}
}
