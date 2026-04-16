package Query

import (
	"BHLayer2Node/paradigm"

	bcostypes "github.com/FISCO-BCOS/go-sdk/v3/types"
)

func safeReceiptTxHash(receipt *bcostypes.Receipt) string {
	if receipt == nil {
		return ""
	}
	return receipt.TransactionHash
}

func safeReceiptBlockHeight(receipt *bcostypes.Receipt) interface{} {
	if receipt == nil {
		return nil
	}
	return receipt.BlockNumber
}

func safeReceiptContract(receipt *bcostypes.Receipt) interface{} {
	if receipt == nil {
		return nil
	}
	return receipt.To
}

func safeReceiptMerkleRoot(receipt *bcostypes.Receipt) string {
	if receipt == nil {
		return ""
	}
	return paradigm.CalculateMerkleRoot(receipt)
}

func safeReceiptProof(receipt *bcostypes.Receipt) interface{} {
	if receipt == nil {
		return nil
	}
	return receipt.ReceiptProof
}
