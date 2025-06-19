package Validatable

import (
	"BHLayer2Node/Validatable/merkle"
	"BHLayer2Node/Validatable/verkle"
	"BHLayer2Node/paradigm"
	"fmt"
)

type Validatable interface {
	Build(leaves [][]byte) error
	GetProof(targetIndex int) (interface{}, bool)
	Verify(targetHash []byte, proof interface{}) bool
}

type StructType int

const (
	Merkle StructType = iota
	Verkle
)

func NewValidatable(data [][]byte, st StructType) (Validatable, error) {
	var vStruct Validatable
	switch st {
	case Merkle:
		vStruct = &merkle.MerkleTree{}
	case Verkle:
		vStruct = &verkle.VerkleTree{}
	default:
		paradigm.Error(paradigm.RuntimeError, "unsupported validatable type to create")
		return nil, fmt.Errorf("buid failed")
	}
	err := vStruct.Build(data)
	if err != nil {
		return nil, err
	}
	return vStruct, nil
}
