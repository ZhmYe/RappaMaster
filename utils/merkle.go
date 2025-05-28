package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func BuildMerkleTree(leaves [][]byte) ([][][]byte, []byte) {
	if len(leaves) == 0 {
		return nil, nil
	}

	tree := [][][]byte{leaves}
	currentLevel := leaves

	for len(currentLevel) > 1 {
		var nextLevel [][]byte
		for i := 0; i < len(currentLevel); i += 2 {
			if i+1 < len(currentLevel) {
				combined := append(currentLevel[i], currentLevel[i+1]...)
				hash := sha256.Sum256(combined)
				nextLevel = append(nextLevel, hash[:])
			} else {
				// odd node gets promoted
				nextLevel = append(nextLevel, currentLevel[i])
			}
		}
		tree = append(tree, nextLevel)
		currentLevel = nextLevel
	}
	return tree, tree[len(tree)-1][0]
}

type MerkleProofItem struct {
	Hash      string
	Position  string
	Level     int
	NodeIndex int
}

func GetMerkleProof(tree [][][]byte, targetIndex int) ([]MerkleProofItem, bool) {
	if len(tree) == 0 {
		return nil, false
	}

	proof := []MerkleProofItem{}
	index := targetIndex

	for level := 0; level < len(tree)-1; level++ {
		currentLevel := tree[level]
		siblingIndex := index ^ 1

		if siblingIndex < len(currentLevel) {
			position := "left"
			if index%2 == 0 {
				position = "right"
			}
			proof = append([]MerkleProofItem{{
				Hash:      fmt.Sprintf("%x", currentLevel[siblingIndex]),
				Position:  position,
				Level:     level,
				NodeIndex: siblingIndex,
			}}, proof...)
		}
		index /= 2
	}

	return proof, true
}

func VerifyMerkleProof(leaf []byte, proof []MerkleProofItem, root []byte) bool {
	computedHash := leaf
	for _, p := range proof {
		sibling, err := hex.DecodeString(p.Hash)
		if err != nil {
			return false
		}
		if p.Position == "left" {
			computed := append(sibling, computedHash...)
			sum := sha256.Sum256(computed)
			computedHash = sum[:]
		} else if p.Position == "right" {
			computed := append(computedHash, sibling...)
			sum := sha256.Sum256(computed)
			computedHash = sum[:]
		}
	}
	return bytes.Equal(computedHash, root)
}
