package paradigm

// Proof 这里统一定义可信证明，后续可接入gnark
type Proof interface {
}

type ProofReceipt struct {
    SlotHash SlotHash
    Proof    []byte
}