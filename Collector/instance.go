package Collector

import (
	"BHLayer2Node/PKI"
	"BHLayer2Node/paradigm"
	pb "BHLayer2Node/pb/service"
	"fmt"
)

// CollectSlotInstance 收集实例
type CollectSlotInstance struct {
	OutputType paradigm.ModelOutputType
	//Mission         string
	Slots []paradigm.CollectSlotItem // 所有要收集的slot
	//Transfer        chan interface{}           // 传给http，用于返回给Backend
	ResponseChannel chan pb.RecoverResponse // 这里是grpc收到response以后通过这个channel传回collector
	//Connection      chan paradigm.RecoverConnection // 这里是传递给grpc的channel
	Channel *paradigm.RappaChannel
	Manager *PKI.PKIManager
}

func (i *CollectSlotInstance) Collect() (interface{}, error) {
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
			outputType:  i.OutputType,
			paddingSize: slot.PaddingSize,
			storeMethod: slot.StoreMethod,
			dataHash:    "",
			sign:        slot.Sign,
			nodeId:      slot.NodeId,
			nodeSign:    slot.NodeSign,
			//output:     make([]byte, 0),
		}
	}

	conn := paradigm.RecoverConnection{
		//Mission:         i.Mission,
		Hashs:           slotHashs,
		ResponseChannel: i.ResponseChannel,
	}
	i.Channel.SlotCollectChannel <- conn // 将request发给grpc

	// 等待response
	for response := range i.ResponseChannel {
		//这里收到的是某个节点的关于slotHash的chunk todo 这里的data还要改，要包含row_index, col_index等，恢复过程也要改
		//更新对应的recover
		paradigm.Log("COLLECT", fmt.Sprintf("Receive Recover Response, Len(Chunks) = %d", len(response.Chunks)))
		for _, chunk := range response.Chunks {
			if _, exist := recovers[chunk.Hash]; !exist {
				paradigm.Error(paradigm.RuntimeError, fmt.Sprintf("Recover %s does not exist", chunk.Hash))
				continue
			}
			slotRecover := recovers[chunk.Hash]
			slotRecover.Add(chunk)
		}
	}
	// 在grpc完成通信后，关闭channel
	// 此时这里可运行
	outputs := make([]interface{}, 0)
	isError := false
	errorList := make([]string, 0)
	for _, r := range recovers {
		//先检查
		if i.Manager.VertifyNodeSign(r.nodeId, r.dataHash, r.nodeSign) {
			recoverOutput := r.Recover()
			outputs = append(outputs, recoverOutput)
		} else {
			paradigm.Error(paradigm.RuntimeError, fmt.Sprintf("Node %d sign verify failed", r.nodeId))
			isError = true
			errorList = append(errorList, fmt.Sprintf("Node %d sign verify failed", r.nodeId))
		}
		//i.Transfer <- recoverOutput
	}
	if isError {
		return nil, fmt.Errorf("Collect failed, error list: %v", errorList)
	}
	finalRecover := SlotRecover{
		slotHash:    "",
		commitment:  nil,
		chunks:      nil,
		k:           0,
		n:           0,
		outputType:  i.OutputType,
		paddingSize: nil,
		storeMethod: 0,
		nodeSign:    "",
	}
	output := finalRecover.merge(outputs)
	//close(i.Transfer)
	return output, nil

}
