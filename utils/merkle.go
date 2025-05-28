package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type MerkleProofItem struct {
	Position string `json:"position"` // "left" or "right"
	Hash     string `json:"hash"`     // hex string
}

// 构造 Merkle Tree，返回每层节点与 Merkle Root
func BuildMerkleTree(leaves [][]byte) ([][][]byte, []byte) {
	if len(leaves) == 0 {
		return nil, nil
	}
	tree := [][][]byte{leaves}
	currentLevel := leaves
	for len(currentLevel) > 1 {
		var nextLevel [][]byte
		for i := 0; i < len(currentLevel); i += 2 {
			var left = currentLevel[i]
			var right []byte
			if i+1 < len(currentLevel) {
				right = currentLevel[i+1]
			} else {
				right = left
			}
			h := sha256.Sum256(append(left, right...))
			nextLevel = append(nextLevel, h[:])
		}
		tree = append(tree, nextLevel)
		currentLevel = nextLevel
	}
	return tree, tree[len(tree)-1][0]
}

// 生成某个叶子节点的 Proof 路径
func GetMerkleProof(tree [][][]byte, index int) ([]MerkleProofItem, bool) {
	if len(tree) == 0 || index < 0 || index >= len(tree[0]) {
		return nil, false
	}
	var proof []MerkleProofItem
	for level := 0; level < len(tree)-1; level++ {
		layer := tree[level]
		siblingIndex := index ^ 1
		if siblingIndex >= len(layer) {
			index /= 2
			continue
		}
		position := "left"
		if index%2 == 0 {
			position = "right"
		}
		proof = append(proof, MerkleProofItem{
			Position: position,
			Hash:     fmt.Sprintf("%x", layer[siblingIndex]),
		})

		index /= 2
	}
	return proof, true
}

func VerifyMerkleProof(leaf []byte, proof []MerkleProofItem, root []byte) bool {
	hash := leaf
	for _, p := range proof {
		siblingHashBytes, err := hex.DecodeString(p.Hash)
		if err != nil {
			return false
		}
		if p.Position == "left" {
			h := sha256.Sum256(append(siblingHashBytes, hash...))
			hash = h[:]
		} else {
			h := sha256.Sum256(append(hash, siblingHashBytes...))
			hash = h[:]
		}
	}
	return bytes.Equal(hash, root)
}
