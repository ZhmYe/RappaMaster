package paradigm

import pb "BHLayer2Node/pb/service"

// CollectSlotItem 这里的Slot已经经过了finalized，无需记录其他的状态
// 考虑到用户可能不是一次性下载所有数据，更常见的应该是download多少数据
// 所以要做的其实是按序遍历下来，要注意存储有序
type CollectSlotItem struct {
	Sign string   // 这里其实可以不记录sign
	Hash SlotHash // 主要是以这个作为标识
	Size int32    // 表示这个slot包含了多少的数据
}

// CollectRequest 表示一个来自前端的收集请求
// 对某个合成任务sign的收集，一共要下载多少size的数据
type CollectRequest struct {
	Sign    string
	Size    int32
	Mission string // 标识
	// 数据传递通道
	TransferChannel chan []byte // 这里暂时写成[]byte
}

type RecoverConnection struct {
	Mission         string
	Hashs           []SlotHash              // 多个slotHash todo 考虑分批?
	ResponseChannel chan pb.RecoverResponse // 这里是grpc收到response以后通过这个channel传回collector
}

// CollectSlotInstance 收集实例
type CollectSlotInstance struct {
	Mission         string
	SlotHashs       []SlotHash              // 所有要收集的哈希
	Transfer        chan []byte             // 传给http，用于返回给Backend
	ResponseChannel chan pb.RecoverResponse // 这里是grpc收到response以后通过这个channel传回collector
	Connection      chan RecoverConnection  // 这里是传递给grpc的channel
}

func (i *CollectSlotInstance) Collect() {
	// 这里的逻辑是，将要collect的内容发给grpc client，然后通过grpc发送到节点，节点返回ec chunk
	// 恢复后通过channel返回给http，进而给backend
	// 启动SlotRecover
	recovers := make(map[SlotHash]*SlotRecover)
	for _, slotHash := range i.SlotHashs {
		recovers[slotHash] = &SlotRecover{
			slotHash:   slotHash,
			commitment: nil, // todo
			chunks:     make([][]*pb.RecoverSlotChunk, 0),
			//output:     make([]byte, 0),
		}
	}

	conn := RecoverConnection{
		Mission:         i.Mission,
		Hashs:           i.SlotHashs,
		ResponseChannel: i.ResponseChannel,
	}
	i.Connection <- conn // 将request发给grpc

	// 等待response
	for response := range i.ResponseChannel {
		//这里收到的是某个节点的关于slotHash的chunk todo 这里的data还要改，要包含row_index, col_index等，恢复过程也要改
		//更新对应的recover
		for _, chunk := range response.Chunks {
			if _, exist := recovers[chunk.Hash]; !exist {
				panic("No such Recover...Channel Transfer Error or Runtime Error!!!")
			}
			slotRecover := recovers[chunk.Hash]
			slotRecover.Add(chunk)
		}
	}
	// 在grpc完成通信后，关闭channel
	// 此时这里可运行
	for _, r := range recovers {
		recoverOutput := r.Recover()
		i.Transfer <- recoverOutput
	}
	close(i.Transfer)

}
