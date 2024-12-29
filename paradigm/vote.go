package paradigm

// SlotVote 针对某一个slot的投票
type SlotVote struct {
	//Slot    CommitSlotItem
	Hash       SlotHash
	Commitment SlotCommitment
	Total      int            // 节点总数
	Vote       int            // 收到的投票数
	Message    map[int]string // 这个简单记录所有投票
}

func (v *SlotVote) Accept(id int) {
	v.Vote++
	v.Total++
	v.Message[id] = "OK" // 简单写一下,也可以是从vote里拿出来
}
func (v *SlotVote) Reject(id int, desp string) {
	v.Total++
	v.Message[id] = desp
}
func (v *SlotVote) Check() bool {
	check := func() bool {
		// 这里简单写一下投票规则，初步想法是收到的同意需要超过k票，保证纠删码可恢复 todo
		return 2*v.Vote >= v.Total
	}
	return check()
}
func NewSlotVote(hash SlotHash, commitment SlotCommitment) *SlotVote {
	return &SlotVote{
		Hash:       hash,
		Commitment: commitment,
		Total:      0,
		Vote:       0,
		Message:    make(map[int]string),
	}
}
