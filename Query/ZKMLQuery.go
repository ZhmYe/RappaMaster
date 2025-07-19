package Query

// CommitProofRequest 定义了从外部Prover提交证明时HTTP请求体的结构。
type CommitProofRequest struct {
	SlotHash string `json:"slotHash" binding:"required"`
	Proof    string `json:"proof" binding:"required"` // 接收 Base64 编码的字符串
}
