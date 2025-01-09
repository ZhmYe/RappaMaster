package paradigm

type DevQueryType int

const (
	EpochQuery          = iota // 查询一个epoch内发生了什么
	TaskSlotQuery              // 查询一个task在某个slot内的完成情况
	EpochRangeQuery            // 查询epoch_i ~ epoch_j
	TaskSlotRangeQuery         // 查询某个task在slot_i ~ slot_j内的完成情况
	TxReceiptQuery             // 查询某个tx的receipt
	BatchTxReceiptQuery        // 查询某些tx的receipt
	// todo
)
