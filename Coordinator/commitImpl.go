package Coordinator

import (
	"BHLayer2Node/LogWriter"
	"BHLayer2Node/paradigm"
	pb "BHLayer2Node/pb/service"
	"context"
	"fmt"
	"math/rand"
)

// CommitSlot 服务端方法，用于处理提交的slot
func (c *Coordinator) CommitSlot(ctx context.Context, req *pb.SlotCommitRequest) (*pb.SlotCommitResponse, error) {
	//nodeId, _ := strconv.Atoi(req.NodeId)
	//slot, _ := strconv.Atoi(req.Slot)
	LogWriter.Log("DEBUG", "successfully receive commit slot") // TODO

	item := paradigm.NewCommitSlotItem(&pb.JustifiedSlot{
		Nid:        req.NodeId,
		Process:    req.Size,
		Sign:       req.Sign,
		Slot:       req.Slot,
		Epoch:      -1, // 这里先不加真正的epoch,等待TaskManager
		Padding:    req.Padding,
		Store:      req.Store,
		Commitment: req.Commitment,
	})
	item.SetHash(req.Hash) // 设置slotHash
	//TODO  @YZM 将验证后的结果放入commitSlot 这里目前没想好验什么
	c.channel.CommitSlots <- item
	LogWriter.Log("COORDINATOR", fmt.Sprintf("successfully receive commit slot: {%s}", item.SlotHash()))
	generateRandomSeed := func() []byte {
		size := 256 // 暂定
		randomBytes := make([]byte, size)

		// 使用 crypto/rand 生成安全随机字节
		rand.Read(randomBytes)
		//if err != nil {
		//	return nil, err
		//}

		return randomBytes
	}

	// 这里就简单的回复即可，后续所有的东西都由heartbeat、chainupper来给定
	return &pb.SlotCommitResponse{
		Seed:    generateRandomSeed(),
		Timeout: int32(c.maxEpochDelay),
		Hash:    item.SlotHash(), // TODO: Hash是用来让节点在收到heartbeat后判断： 1. 是否本地有一些未确认的数据不需要保存了; 2. 是否自己的任务被拒绝了，是否太慢了
	}, nil
}
