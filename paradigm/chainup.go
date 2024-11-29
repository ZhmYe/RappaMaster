package paradigm

type Transaction interface {
	Call() string                   // 调用合约哪个函数
	Params() map[string]interface{} // 调用合约函数时候的参数
}

// CommitSlotTransaction 就是对节点完成情况进行上链
type CommitSlotTransaction struct {
	CommitSlotItem
	//Votes []int // 这里就先简单用[]int表示下，这里是所有节点的投票
	Epoch int
}

func (t *CommitSlotTransaction) Call() string {
	return "Commit" // 简单先写一下，后面具体和合约对齐
}
func (t *CommitSlotTransaction) Params() map[string]interface{} {
	result := make(map[string]interface{})
	result["Sign"] = t.Sign
	result["Slot"] = t.Slot
	result["Process"] = t.Process
	result["ID"] = t.Nid
	//result["Vote"] = t.Votes
	result["Epoch"] = t.Epoch
	return result
}
