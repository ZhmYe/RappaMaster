package paradigm

type TaskHash = string
type ScheduleHash = int32
type SlotHash = string // 用来表示一个slot(node,sign,slot) TODO

type SlotCommitment = []byte // 在这里写commitment的结构，后续好改一点

type TxHash = string // 交易哈希，这里暂定为string
