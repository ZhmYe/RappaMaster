package paradigm

import (
	"BHLayer2Node/LogWriter"
	pb "BHLayer2Node/pb/service"
	"fmt"
	"sort"
)

// SlotRecover 收集发过来的slotHash对应的ec chunk，然后还原回一个数据
type SlotRecover struct {
	slotHash   SlotHash
	commitment SlotCommitment           // 对应还原后executor提交的承诺，还原后要验证
	chunks     [][]*pb.RecoverSlotChunk // todo 这里暂时就写成byte
	//output     []byte                   // 这里暂时先写成byte
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
func (r *SlotRecover) Recover() []byte {
	// todo 这里还需要check，以及最好是及时的recover这样可以并行
	// 针对每一个行块进行恢复
	output := []byte{}
	for _, rowChunks := range r.chunks {
		// todo 这里要判断纠删码k
		check := func() bool {
			return true
		}
		if check() {
			rowChunkRecoverOuput := r.recoverRowChunk(rowChunks)
			output = append(output, rowChunkRecoverOuput...)
		}
	}
	return output
}

func (r *SlotRecover) recoverRowChunk(chunks []*pb.RecoverSlotChunk) []byte {
	// 按排列顺序，取前k个
	sort.Slice(chunks, func(i int, j int) bool {
		return chunks[i].Col < chunks[j].Col
	})
	LogWriter.Log("COLLECT", fmt.Sprintf("Start recover %s row %d chunk", chunks[0].Hash, chunks[0].Row))
	for _, chunk := range chunks {
		fmt.Println(chunk.Row, chunk.Col, chunk.Chunk)
	}
	return []byte(fmt.Sprintf("%s_%d", chunks[0].Hash, chunks[0].Row)) // todo 这里接入ec decoder

}
