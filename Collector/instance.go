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
	ResponseChannel chan pb.RecoverResponse // 这里是grpc收到response以后通过这个channel传回collector
	//Connection      chan paradigm.RecoverConnection // 这里是传递给grpc的channel
	Channel *paradigm.RappaChannel
	Manager *PKI.PKIManager
}

func (i *CollectSlotInstance) Collect() *io.PipeReader {
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
		connections[j] = paradigm.SlotRecoverConnection{
			Hash:            slotHashs[j],
			ResponseChannel: i.ResponseChannel, // 这里可以共用一个，因为我们接收也是一个个来的
		}
	}
	//conn := paradigm.RecoverConnection{
	//	//Mission:         i.Mission,
	//	Hashs:           slotHashs,
	//	ResponseChannel: i.ResponseChannel,
	//}
	//i.Channel.SlotCollectChannel <- conn // 将request发给grpc
	r, w := io.Pipe() // 这里创建流式内存
	go func() {
		defer func() {
			err := w.Close()
			if err != nil {
				paradigm.Error(paradigm.RuntimeError, fmt.Sprintf("close io pipe writer error: %v", err))
			}
		}()
		for _, conn := range connections {
			i.Channel.SlotCollectChannel <- conn
			switch recovers[conn.Hash].storeMethod {
			case 1:
				// 这里只有一个chunk，所以应该只会收到一个 todo
				response := <-i.ResponseChannel
				paradigm.Log("COLLECT", fmt.Sprintf("Receive Recover Response, Len(Chunks) = %d", len(response.Chunks)))
				if len(response.Chunks) != 1 {
					paradigm.Error(paradigm.ChunkRecoverError, "More than 1 chunk in local store")
					continue
				}
				chunk := response.Chunks[0]
				if _, exist := recovers[chunk.Hash]; !exist || chunk.Hash != conn.Hash {
					paradigm.Error(paradigm.ChunkRecoverError, "Error chunk Hash")
					continue
				}
				slotRecover := recovers[chunk.Hash]
				slotRecover.Add(chunk)
				// 此时recover已经收好了
				recoverOutput := slotRecover.Recover() // 还原
				// 这里不调用merge了，应该[]byte就是可以直接拼的，测试一下@SD
				// 然后直接写到写管道里，此时会阻塞
				fileByte, _, err := paradigm.DataToFile(recoverOutput)
				if err != nil {
					paradigm.Error(paradigm.ChunkRecoverError, "Error chunk Hash")
					continue
				}
				_, err = w.Write(fileByte)
				if err != nil {
					paradigm.Error(paradigm.ChunkRecoverError, "Error file byte")
					continue
				}
				// 写完这里应该会阻塞下一个write，这样我们至多在内存里有两个connection
			default:
				paradigm.Error(paradigm.ChunkRecoverError, "Slot store in EC low-memory recovery not impl yet")
				continue // todo 这里先不panic了
			}

		}
	}()
	return r // 返回读管道
	// 等待response
	//for response := range i.ResponseChannel {
	//	//这里收到的是某个节点的关于slotHash的chunk todo 这里的data还要改，要包含row_index, col_index等，恢复过程也要改
	//	//更新对应的recover
	//	paradigm.Log("COLLECT", fmt.Sprintf("Receive Recover Response, Len(Chunks) = %d", len(response.Chunks)))
	//	for _, chunk := range response.Chunks {
	//		if _, exist := recovers[chunk.Hash]; !exist {
	//			paradigm.Error(paradigm.RuntimeError, fmt.Sprintf("Recover %s does not exist", chunk.Hash))
	//			continue
	//		}
	//		slotRecover := recovers[chunk.Hash]
	//		slotRecover.Add(chunk)
	//	}
	//}
	//// 在grpc完成通信后，关闭channel
	//// 此时这里可运行
	//outputs := make([]interface{}, 0)
	//for _, r := range recovers {
	//	recoverOutput := r.Recover()
	//	outputs = append(outputs, recoverOutput)
	//	//i.Transfer <- recoverOutput
	//}
	//finalRecover := SlotRecover{
	//	slotHash:    "",
	//	commitment:  nil,
	//	chunks:      nil,
	//	k:           0,
	//	n:           0,
	//	outputType:  i.OutputType,
	//	paddingSize: nil,
	//	storeMethod: 0,
	//}
	//output := finalRecover.merge(outputs)
	////close(i.Transfer)
	//return output

}
