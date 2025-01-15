package Collector

import (
	"BHLayer2Node/LogWriter"
	"BHLayer2Node/paradigm"
	pb "BHLayer2Node/pb/service"
	"fmt"
)

// CollectSlotInstance 收集实例
type CollectSlotInstance struct {
	Mission         string
	Slots           []paradigm.CollectSlotItem // 所有要收集的slot
	Transfer        chan interface{}           // 传给http，用于返回给Backend
	ResponseChannel chan pb.RecoverResponse    // 这里是grpc收到response以后通过这个channel传回collector
	//Connection      chan paradigm.RecoverConnection // 这里是传递给grpc的channel
	Channel *paradigm.RappaChannel
}

func (i *CollectSlotInstance) Collect() {
	// 这里的逻辑是，将要collect的内容发给grpc client，然后通过grpc发送到节点，节点返回ec chunk
	// 恢复后通过channel返回给http，进而给backend
	// 启动SlotRecover
	recovers := make(map[paradigm.SlotHash]*SlotRecover)
	slotHashs := make([]paradigm.SlotHash, 0)
	for _, slot := range i.Slots {
		slotHash := slot.Hash
		slotHashs = append(slotHashs, slotHash)
		recovers[slotHash] = &SlotRecover{
			slotHash:    slotHash,
			commitment:  nil, // todo
			chunks:      make([][]*pb.RecoverSlotChunk, 0),
			n:           i.Channel.Config.ErasureCodeParamN,
			k:           i.Channel.Config.ErasureCodeParamK,
			outputType:  slot.OutputType,
			paddingSize: slot.PaddingSize,
			//output:     make([]byte, 0),
		}
	}

	conn := paradigm.RecoverConnection{
		Mission:         i.Mission,
		Hashs:           slotHashs,
		ResponseChannel: i.ResponseChannel,
	}
	i.Channel.SlotCollectChannel <- conn // 将request发给grpc

	// 等待response
	for response := range i.ResponseChannel {
		//这里收到的是某个节点的关于slotHash的chunk todo 这里的data还要改，要包含row_index, col_index等，恢复过程也要改
		//更新对应的recover
		LogWriter.Log("COLLECT", fmt.Sprintf("Receive Recover Response, Len(Chunks) = %d", len(response.Chunks)))
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
