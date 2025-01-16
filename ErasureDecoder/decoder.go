package ErasureDecoder

import (
	pb "BHLayer2Node/pb/service"
	"fmt"
	"github.com/klauspost/reedsolomon"
)

// ErasureDecoder 处理纠删码，将chunks还原，这里仅考虑一个块，行块由recover维护
type ErasureDecoder struct {
	k int // 数据块数量
	n int // 数据块 + 冗余块 数量, k + m = n
}

func (decoder *ErasureDecoder) Decode(chunks []*pb.RecoverSlotChunk) ([]byte, error) {
	processChunks := func(chunks []*pb.RecoverSlotChunk) ([][]byte, error) {
		if len(chunks) == 0 {
			return [][]byte{}, fmt.Errorf("empty Chunks")
		}
		if len(chunks) > decoder.n {
			return [][]byte{}, fmt.Errorf("too many chunks: len(chunks) = %d, expected <= %d", len(chunks), decoder.n)
		}
		dataChunks := make([][]byte, decoder.n)
		row := chunks[0].Row
		for i := 0; i < len(chunks); i++ {
			chunk := chunks[i]
			if chunk.Row != row {
				return [][]byte{}, fmt.Errorf("chunks should in the same row")
			}
			dataChunks[chunk.Col] = chunk.Chunk
		}
		return dataChunks, nil
	}
	dataChunks, err := processChunks(chunks)
	if err != nil {
		return []byte{}, err
	}
	rs, err := reedsolomon.New(decoder.k, decoder.n-decoder.k)
	if err != nil {
		return []byte{}, err
	}
	err = rs.Reconstruct(dataChunks)
	if err != nil {
		return []byte{}, fmt.Errorf("Failed to reconstruct data: %v\n", err)
	}
	toBytes := func(chunks [][]byte) []byte {
		var result []byte
		// 前k个是数据块
		for i := 0; i < decoder.k; i++ {
			chunk := chunks[i]
			result = append(result, chunk...)
		}
		return result
	}
	return toBytes(dataChunks), nil
}
func NewErasureDecoder(k int, n int) ErasureDecoder {
	return ErasureDecoder{k: k, n: n}
}
