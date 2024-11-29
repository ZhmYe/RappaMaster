package handler

//
//// VoteHandler 每当一个epoch结束就将在这个epoch内提交的所有内容全部上链
//// 目前先考虑上链的是节点的commitSlot
//type VoteHandler struct {
//	commitSlots         []paradigm.CommitSlotItem
//	slotToVotes         chan paradigm.CommitSlotItem
//	pendingTransactions chan []paradigm.Transaction
//	epochChangeEvent    chan bool // 外部触发的 epoch 更新信号
//	unprocessedIndex    int       // 还未处理的index开头，包括这一index
//	currentEpoch        int
//}
//
//func (h *VoteHandler) ProcessEpochVote(epoch int, commitSlots []paradigm.CommitSlotItem) []paradigm.Transaction {
//	// 这里先不写投票过程，就直接给出结果
//	// todo 完成投票的grpc，然后将投票结果不够的剔除
//	transactions := make([]paradigm.Transaction, 0)
//	for _, slot := range commitSlots {
//		transaction := paradigm.CommitSlotTransaction{
//			CommitSlotItem: slot,
//			//Votes:          make([]int, 0),
//		}
//		transactions = append(transactions, &transaction)
//	}
//	return transactions
//}
//func (h *VoteHandler) UpdateEpoch() {
//	// 更新epoch，并打包一部分commitSlot出来作为上一个epoch的上链内容
//	// todo 这里先简单写成有多少给多少
//	LogWriter.Log("VOTE", fmt.Sprintf("Start Epoch %d Commit Slot Vote...", h.currentEpoch))
//	//for index := h.unprocessedIndex; index < len(h.slotToVotes); index++ {
//	//	transaction := paradigm.CommitSlotTransaction{
//	//		CommitSlotItem: h.commitSlots[index],
//	//		Votes:          nil,
//	//	}
//	//	h.pendingTransactions <- &transaction
//	//}
//	transactions := h.ProcessEpochVote(h.currentEpoch, h.commitSlots[h.unprocessedIndex:])
//	h.unprocessedIndex = len(h.commitSlots)
//	go func() { h.pendingTransactions <- transactions }()
//	h.currentEpoch++
//}
//func (h *VoteHandler) Collect(commitSlot paradigm.CommitSlotItem) {
//	// 这里就是把slot放进去
//	h.commitSlots = append(h.commitSlots, commitSlot)
//	// todo
//	//transaction := paradigm.CommitSlotTransaction{CommitSlotItem: commitSlot, Votes: make([]int, 0)}
//	//h.pendingTransactions <- &transaction
//
//}
//func (h *VoteHandler) Start() {
//	for {
//		select {
//		case <-h.epochChangeEvent:
//			h.UpdateEpoch()
//		case commitSlot := <-h.slotToVotes:
//			h.Collect(commitSlot)
//
//		}
//	}
//
//}
