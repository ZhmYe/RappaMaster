package Collector

import (
	"BHLayer2Node/ErasureDecoder"
	"BHLayer2Node/paradigm"
	pb "BHLayer2Node/pb/service"
	"fmt"
	"github.com/go-gota/gota/dataframe"
	"sort"
)

// SlotRecover 收集发过来的slotHash对应的ec chunk，然后还原回一个数据
type SlotRecover struct {
	slotHash   paradigm.SlotHash
	commitment paradigm.SlotCommitment  // 对应还原后executor提交的承诺，还原后要验证
	chunks     [][]*pb.RecoverSlotChunk // todo 这里暂时就写成byte
	k          int
	n          int
	outputType paradigm.ModelOutputType
	//output     []byte                   // 这里暂时先写成byte
	paddingSize []int32 // row chunk的padding size
	storeMethod int32   // 存储方式，也是恢复方式, 一个slot只会有一种方式
}

func (r *SlotRecover) Add(chunk *pb.RecoverSlotChunk) {
	// todo
	row := chunk.Row
	if chunk.Hash != r.slotHash {
		panic("Add Error Chunk to recover!!!")
	}
	for row >= int32(len(r.chunks)) {
		r.chunks = append(r.chunks, []*pb.RecoverSlotChunk{})
	}
	r.chunks[row] = append(r.chunks[row], chunk)
	//r.chunks = append(r.chunks, chunk)
}
func (r *SlotRecover) Recover() interface{} {
	// todo 这里还需要check，以及最好是及时的recover这样可以并行
	// 针对每一个行块进行恢复
	//output := [][]byte{}
	var recoverOutputs []interface{}
	for row, rowChunks := range r.chunks {
		if len(rowChunks) == 0 {

			paradigm.Error(paradigm.ChunkRecoverError, fmt.Sprintf("Recover %s row %d chunk failed: Empty Chunks", r.slotHash, row))
		} else {
			rowChunkRecoverOutput, err := r.recoverRowChunk(rowChunks, row)
			if err != nil {
				paradigm.Error(paradigm.ChunkRecoverError, fmt.Sprintf("Recover %s row %d chunk failed: %v", rowChunks[0].Hash, rowChunks[0].Row, err))
			}
			transformer := OutputTransformer{outputType: r.outputType} // 统一转化
			paradigm.Log("COLLECT", fmt.Sprintf("Try to transform to original output type, type: %s", paradigm.ModelOutputTypeToString(r.outputType)))

			output, err := transformer.Transform(rowChunkRecoverOutput)
			if err != nil {
				paradigm.Error(paradigm.DataTransformError, fmt.Sprintf("Transform to %s Error: %v", paradigm.ModelOutputTypeToString(r.outputType), err))
			}
			recoverOutputs = append(recoverOutputs, output)
			//output = append(output, rowChunkRecoverOutput)
		}
	}
	// 到这里得到了所有的rawData: [][]byte，每一维是rowChunk
	// 下面要恢复成正常的输出
	return r.merge(recoverOutputs)
}

func (r *SlotRecover) recoverRowChunk(chunks []*pb.RecoverSlotChunk, row int) ([]byte, error) {
	// 按排列顺序，取前k个
	switch r.storeMethod {
	case 1:
		// 本地存储，那么只有可能有一个块
		if len(chunks) != 1 {
			e := paradigm.Error(paradigm.RuntimeError, fmt.Sprintf("Local only have one chunk in each chunk, but receive %d", len(chunks)))
			return []byte{}, fmt.Errorf(e.Error())
		}
		paradigm.Log("COLLECT", fmt.Sprintf("Local Store Chunk Receive Success..."))
		return chunks[0].Chunk, nil
	case 2:
		e := paradigm.Error(paradigm.NotImplError, "Replicas Not impl")
		panic(e.Error())
		//panic("Replicas has not been impl...")
	case 3:
		// 纠删码
		padding := r.paddingSize[row]
		sort.Slice(chunks, func(i int, j int) bool {
			return chunks[i].Col < chunks[j].Col
		})
		paradigm.Log("COLLECT", fmt.Sprintf("Start recover %s row %d chunk with EC, padding Size = %d", chunks[0].Hash, chunks[0].Row, padding))
		//for _, chunk := range chunks {
		//	fmt.Println(chunk.Row, chunk.Col, chunk.Chunk)
		//}
		decoder := ErasureDecoder.NewErasureDecoder(r.k, r.n)
		decodedData, err := decoder.Decode(chunks)
		if err != nil {
			return []byte{}, err
		}
		// 对decoded_data去掉padding
		decodedData = decodedData[:len(decodedData)-int(padding)]
		return decodedData, nil
	default:
		e := paradigm.Error(paradigm.RuntimeError, "Unknown Store Method")
		panic(e.Error())
	}

}

// merge 将若干个一样的内容合并成一个，比如dataframe的合并
// todo 这里应该有error
func (r *SlotRecover) merge(rowChunksOutputs []interface{}) interface{} {
	switch r.outputType {
	case paradigm.DATAFRAME:
		mergeDf, ok := rowChunksOutputs[0].(dataframe.DataFrame)
		if !ok {
			paradigm.Error(paradigm.ChunkRecoverError, "output type error")
			return nil
		}
		for i := 1; i < len(rowChunksOutputs); i++ {
			nextDf, ok := rowChunksOutputs[i].(dataframe.DataFrame)
			if !ok {
				paradigm.Error(paradigm.ChunkRecoverError, "output type error")
				continue
			}
			mergeDf = mergeDf.Concat(nextDf)
		}
		return mergeDf
	case paradigm.NETWORK:
		// 这里直接归并数组就ok了
		netList, ok := rowChunksOutputs[0].([]paradigm.Graph)
		if !ok {
			paradigm.Error(paradigm.ChunkRecoverError, "output type error")
			return nil
		}
		for i := 1; i < len(rowChunksOutputs); i++ {
			nextNetList, ok := rowChunksOutputs[i].([]paradigm.Graph)
			if !ok {
				paradigm.Error(paradigm.ChunkRecoverError, "output type error")
				continue
			}
			netList = append(netList, nextNetList...)
		}
		return netList

	default:
		e := paradigm.Error(paradigm.RuntimeError, "Unknown Output Type")
		panic(e.Error())
		//panic("Unknown Output Type!!!")
	}
}
