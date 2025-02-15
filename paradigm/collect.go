package paradigm

import (
	pb "BHLayer2Node/pb/service"
)

// HttpCollectRequest 表示一个来自前端的收集请求
// 对某个合成任务sign的收集，一共要下载多少size的数据
type HttpCollectRequest struct {
	Sign string
	Size int32
	//Mission string // 标识
	// 数据传递通道
	TransferChannel chan interface{}
}
type RappaCollector interface {
	ProcessSlotUpdate(slot CollectSlotItem)
	ProcessCollect(collectRequest HttpCollectRequest) (interface{}, error)
}

// CollectSlotItem 这里的Slot已经经过了finalized，无需记录其他的状态
// 考虑到用户可能不是一次性下载所有数据，更常见的应该是download多少数据
// 所以要做的其实是按序遍历下来，要注意存储有序
type CollectSlotItem struct {
	Sign string   // 这里其实可以不记录sign
	Hash SlotHash // 主要是以这个作为标识
	Size int32    // 表示这个slot包含了多少的数据
	//OutputType  ModelOutputType // 模型输出格式，用于collector恢复
	PaddingSize []int32 // 所有的padding size
	StoreMethod int32   // 存储方式
}

type RecoverConnection struct {
	//Mission         string
	Hashs           []SlotHash              // 多个slotHash todo 考虑分批?
	ResponseChannel chan pb.RecoverResponse // 这里是grpc收到response以后通过这个channel传回collector
}
