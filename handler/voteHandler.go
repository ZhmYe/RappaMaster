package handler

import (
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
			// 找到对应的slot
			instance, exist := handler.voteInstance[fmt.Sprintf("%s", vote.Hash)]
			if exist {
				if vote.State {
					// 如果同意
					instance.Accept(int(vote.NodeId))
				} else {
					// 如果拒绝
					instance.Reject(int(vote.NodeId), vote.Desp)
				}
			} else {
				// 这里暂时先不报错
				//paradigm.Error(paradigm.RuntimeError, "Vote Instance does not exist")
				continue
			}

		}
	}
	// 当所有的response处理完毕,channel被关闭后
	// 开始判断所有的instance是否通过投票，如果通过，则将对应的slot作为finalize
	for _, instance := range handler.voteInstance {
		if instance.Check() {
			// 如果通过投票了，那么就finalize TODO @YZM
			//slot := instance.Slot
			//slot.SetFinalize()
			paradigm.Log("VOTE", fmt.Sprintf("%s CommitSlot pass the Vote...", instance.Hash))
			// TODO @YZM 这里简单构造了一个假的commitSlot，因为taskManager只需要hash，可以分成两个channel
			handler.accepts <- paradigm.NewFakeCommitSlotItem(instance.Hash)
		} else {
			paradigm.Error(paradigm.SlotLifeError, fmt.Sprintf("%s CommitSlot does not pass the Vote", instance.Hash))
		}
	}
}

func NewVoteHandler(heartbeat *pb.HeartbeatRequest, accepts chan paradigm.CommitSlotItem, responses chan *pb.HeartbeatResponse) *VoteHandler {
	instances := make(map[string]*paradigm.SlotVote)
	for hash, commitment := range heartbeat.Commits {
		instances[fmt.Sprintf("%s", hash)] = paradigm.NewSlotVote(hash, commitment)
	}
	return &VoteHandler{
		voteInstance: instances,
		responses:    responses,
		epoch:        int(heartbeat.Epoch),
		accepts:      accepts,
	}
}
