package Collector

import (
	"BHLayer2Node/PKI"
	"BHLayer2Node/paradigm"
	pb "BHLayer2Node/pb/service"
	"fmt"
	"io"
)

// CollectSlotInstance 收集实例
type CollectSlotInstance struct {
	OutputType paradigm.ModelOutputType
	//Mission         string
	Slots []paradigm.CollectSlotItem // 所有要收集的slot
	//Transfer        chan interface{}           // 传给http，用于返回给Backend
	ResponseChannel chan *pb.RecoverResponse // 这里是grpc收到response以后通过这个channel传回collector
	//Connection      chan paradigm.RecoverConnection // 这里是传递给grpc的channel
	Channel *paradigm.RappaChannel
	Manager *PKI.PKIManager
}

func (i *CollectSlotInstance) Collect() (*io.PipeReader, error) {
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
	// 这里我们改成一个个slot来
	connections := make([]paradigm.SlotRecoverConnection, len(slotHashs))
	for j := 0; j < len(connections); j++ {
		nodeId := recovers[slotHashs[j]].nodeId
		if recovers[slotHashs[j]].storeMethod != 1 {
			nodeId = -1 // broadcast for non-local store methods
		}
		connections[j] = paradigm.SlotRecoverConnection{
			Hash:            slotHashs[j],
			NodeId:          nodeId,
			ResponseChannel: i.ResponseChannel, // 这里可以共用一个，因为我们接收也是一个个来的
		}
	}

	r, w := io.Pipe() // 这里创建流式内存
	go func() {
		defer func() {
			err := w.Close()
			if err != nil {
				paradigm.Error(paradigm.RuntimeError, fmt.Sprintf("close io pipe writer error: %v", err))
			}
		}()
		numNodes := len(i.Channel.Config.BHNodeAddressMap)
		for _, conn := range connections {
			i.Channel.SlotCollectChannel <- conn
			// 根据 conn.NodeId 决定等待多少个 response
			numExpectedResponses := 1
			if conn.NodeId == -1 {
				numExpectedResponses = numNodes
			}

			for k := 0; k < numExpectedResponses; k++ {
				response := <-i.ResponseChannel
				paradigm.Log("COLLECT", fmt.Sprintf("Receive Recover Response for %s, Len(Chunks) = %d", conn.Hash, len(response.Chunks)))
				for _, chunk := range response.Chunks {
					if _, exist := recovers[chunk.Hash]; !exist {
						paradigm.Error(paradigm.RuntimeError, fmt.Sprintf("Recover %s does not exist", chunk.Hash))
						continue
					}
					slotRecover := recovers[chunk.Hash]
					slotRecover.Add(chunk)
				}
			}
		}

		// 所有slot收集完成，开始恢复和合并
		outputs := make([]interface{}, 0)
		for _, slotHash := range slotHashs {
			sr := recovers[slotHash]
			// 验证签名
			if !i.Manager.VertifyNodeSign(sr.nodeId, sr.dataHash, sr.nodeSign) {
				paradigm.Error(paradigm.SignatureVerifyError, fmt.Sprintf("Verify node %d sign for slot %s failed", sr.nodeId, slotHash))
				continue
			} else {
				paradigm.Log("COLLECT", fmt.Sprintf("Verify node %d sign for slot %s success", sr.nodeId, slotHash))
			}
			data := sr.Recover()
			outputs = append(outputs, data)
		}

		if len(outputs) == 0 {
			paradigm.Error(paradigm.ChunkRecoverError, "No outputs recovered")
			return
		}

		finalRecover := SlotRecover{
			outputType: i.OutputType,
		}
		output := finalRecover.merge(outputs)
		if output == nil {
			paradigm.Error(paradigm.ChunkRecoverError, "Merge failed")
			return
		}

		fileByte, _, err := paradigm.DataToFile(output)
		if err != nil {
			paradigm.Error(paradigm.ChunkRecoverError, fmt.Sprintf("DataToFile error: %v", err))
			return
		}

		_, err = w.Write(fileByte)
		if err != nil {
			paradigm.Error(paradigm.ChunkRecoverError, fmt.Sprintf("Writer Write error: %v", err))
		}
	}()
	return r, nil // 返回读管道
}
