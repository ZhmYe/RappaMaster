package Collector

import (
	"BHLayer2Node/ErasureDecoder"
	"BHLayer2Node/LogWriter"
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
	// todo 还缺一个padding size
	paddingSize []int32 // row chunk的padding size
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
			LogWriter.Log("ERROR", fmt.Sprintf("recover %s row %d chunk failed, err: empty chunks", r.slotHash, row))
		} else {
			rowChunkRecoverOutput, err := r.recoverRowChunk(rowChunks, r.paddingSize[row])
			if err != nil {
				LogWriter.Log("ERROR", fmt.Sprintf("recover %s row %d chunk failed, err: %v", rowChunks[0].Hash, rowChunks[0].Row, err))
			}
			transformer := OutputTransformer{outputType: r.outputType} // 统一转化
			LogWriter.Log("COLLECT", fmt.Sprintf("Try to transform to original output type, type: %s", paradigm.ModelOutputTypeToString(r.outputType)))

			output, err := transformer.Transform(rowChunkRecoverOutput)
			if err != nil {
				LogWriter.Log("ERROR", fmt.Sprintf("Transform to %s Error: %v", paradigm.ModelOutputTypeToString(r.outputType), err))
			}
			recoverOutputs = append(recoverOutputs, output)
			//output = append(output, rowChunkRecoverOutput)
		}
	}
	// 到这里得到了所有的rawData: [][]byte，每一维是rowChunk
	// 下面要恢复成正常的输出
	return r.merge(recoverOutputs)
}

func (r *SlotRecover) recoverRowChunk(chunks []*pb.RecoverSlotChunk, padding int32) ([]byte, error) {
	// 按排列顺序，取前k个
	sort.Slice(chunks, func(i int, j int) bool {
		return chunks[i].Col < chunks[j].Col
	})
	LogWriter.Log("COLLECT", fmt.Sprintf("Start recover %s row %d chunk, padding Size = %d", chunks[0].Hash, chunks[0].Row, padding))
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

}

// merge 将若干个一样的内容合并成一个，比如dataframe的合并
// todo 这里应该有error
func (r *SlotRecover) merge(rowChunksOutputs []interface{}) interface{} {
	switch r.outputType {
	case paradigm.DATAFRAME:
		mergeDf := rowChunksOutputs[0].(dataframe.DataFrame)
		for i := 1; i < len(rowChunksOutputs); i++ {
			mergeDf = mergeDf.Concat(rowChunksOutputs[i].(dataframe.DataFrame))
		}
		return mergeDf
	default:
		panic("Unknown Output Type!!!")
	}
}
