package paradigm

type BlockInfo struct {
	//BlockHash:      block.Hash,
	//"parentHash":     parentHash,
	//"blockHeight":    block.Number,
	//"merkleRoot":     block.StateRoot,
	//"nbTransactions": len(block.Transactions),
	//"transactions":   txDetails,
	BlockHash       string
	ParentHash      string
	BlockHeight     int32
	TransactionRoot string
	Txs             []TransactionInfo
	Invalid         bool
	ErrorMessage    string
}

func (b *BlockInfo) IsValid() bool {
	return b.Invalid
}
func (b *BlockInfo) Error() string {
	return b.ErrorMessage
}
func NewInvalidBlockInfo(e string) BlockInfo {
	return BlockInfo{
		BlockHash:       "",
		ParentHash:      "",
		BlockHeight:     -1,
		TransactionRoot: "",
		Txs:             make([]TransactionInfo, 0),
		Invalid:         false,
		ErrorMessage:    e,
	}
}

func NewMockerBlockInfo() BlockInfo {
	return BlockInfo{
		BlockHash:       "0x123456",
		ParentHash:      "0x123455",
		BlockHeight:     1,
		TransactionRoot: "0x111111",
		Txs:             []TransactionInfo{NewMockerTransactionInfo(), NewMockerTransactionInfo(), NewMockerTransactionInfo()},
		Invalid:         true,
		ErrorMessage:    "",
	}
}

type TransactionInfo struct {
	TxHash    string
	Contract  string
	Abi       string
	BlockHash string
	//UpchainTime  time.Time
	Invalid      bool
	ErrorMessage string
	// TODO
}

func NewMockerTransactionInfo() TransactionInfo {
	return TransactionInfo{
		TxHash:    "",
		Contract:  "",
		Abi:       "STORE",
		BlockHash: "",
		//UpchainTime:  time.Now(),
		Invalid:      true,
		ErrorMessage: "",
	}
}

func NewInvalidTransactionInfo(e string) TransactionInfo {
	return TransactionInfo{
		TxHash:    "",
		Contract:  "",
		Abi:       "",
		BlockHash: "",
		//UpchainTime:  time.Now(),
		Invalid:      false,
		ErrorMessage: e,
	}
}
