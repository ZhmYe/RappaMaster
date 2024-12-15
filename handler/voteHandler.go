package handler

import (
	"BHLayer2Node/LogWriter"
	"BHLayer2Node/paradigm"
	pb "BHLayer2Node/pb/service"
	"fmt"
)

type VoteHandler struct {
	voteInstance map[string]*paradigm.SlotVote // 要处理的所有投票
	responses    chan *pb.HeartbeatResponse    // 节点对心跳的回复
	epoch        int                           // 第几个epoch的投票处理
	accepts      chan paradigm.CommitSlotItem  // 通过投票，用于更新task,taskManager更新完后将其传递到chainUpper
}

func (handler *VoteHandler) Process() {
	for response := range handler.responses {
		votes := response.Votes
		// 遍历回复中的所有投票
		for _, vote := range votes {
			slot := vote.Slot
			// 找到对应的slot
			instance, exist := handler.voteInstance[fmt.Sprintf("%s_%d_%d", slot.Sign, slot.Slot, slot.Nid)]
			if exist {
				if vote.State {
					// 如果同意
					instance.Accept(int(slot.Nid))
				} else {
					// 如果拒绝
					instance.Reject(int(slot.Nid), vote.Desp)
				}
			} else {
				// 这里暂时先不报错
				LogWriter.Log("ERROR", "Vote Instance does not exist")
				continue
			}

		}
	}
	// 当所有的response处理完毕,channel被关闭后
	// 开始判断所有的instance是否通过投票，如果通过，则将对应的slot作为finalize
	for _, instance := range handler.voteInstance {
		if instance.Check() {
			// 如果通过投票了，那么就finalize
			slot := instance.Slot
			slot.SetFinalize()
			LogWriter.Log("VOTE", fmt.Sprintf("%d CommitSlot in Task %s Slot %d pass the Vote...", instance.Slot.Nid, instance.Slot.Sign, instance.Slot.Slot))
			handler.accepts <- slot
		} else {
			LogWriter.Log("VOTE", fmt.Sprintf("%d CommitSlot in Task %s Slot %d does not pass the Vote!!!", instance.Slot.Nid, instance.Slot.Sign, instance.Slot.Slot))
		}
	}
}

func NewVoteHandler(heartbeat *pb.HeartbeatRequest, accepts chan paradigm.CommitSlotItem, responses chan *pb.HeartbeatResponse) *VoteHandler {
	instances := make(map[string]*paradigm.SlotVote)
	for _, slot := range heartbeat.Commits {
		instances[fmt.Sprintf("%s_%d_%d", slot.Sign, slot.Slot, slot.Nid)] = &paradigm.SlotVote{
			Slot: paradigm.NewCommitSlotItem(slot),
			//Slot: paradigm.CommitSlotItem{
			//	Nid:     int(slot.Nid),
			//	Process: int(slot.Process),
			//	Sign:    slot.Sign,
			//	Slot:    int(slot.Slot),
			//
			//},
			Total:   0,
			Vote:    0,
			Message: make(map[int]string),
		}
	}
	return &VoteHandler{
		voteInstance: instances,
		responses:    responses,
		epoch:        int(heartbeat.Epoch),
		accepts:      accepts,
	}
}
