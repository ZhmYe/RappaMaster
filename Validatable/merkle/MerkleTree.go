package merkle

import (
	"crypto/sha256"
	"fmt"
)

type MerkleTree struct {
	tree [][][]byte
}

type MerkleProofItem struct {
	Hash      string
	Position  string
	Level     int
	NodeIndex int
}

func (m *MerkleTree) Build(leaves [][]byte) error {
	if len(leaves) == 0 {
		return fmt.Errorf("no leaves")
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
	m.tree = tree
	return nil
}

func (m *MerkleTree) GetProof(targetIndex int) (interface{}, bool) {
	if len(m.tree) == 0 {
		return nil, false
	}

	proof := []MerkleProofItem{}
	index := targetIndex

	for level := 0; level < len(m.tree)-1; level++ {
		currentLevel := m.tree[level]
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

// todo 这里没有实现
func (m *MerkleTree) Verify(targetHash []byte, proof interface{}) bool {
	return true
}
