package merkle

import (
	"crypto/sha256"
	"fmt"
)

type MerkleTree struct {
	tree [][][]byte
}

func (m *MerkleTree) GetRoot() []byte {
	if m.tree == nil {
		return nil
	}
	return m.tree[len(m.tree)-1][0]
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
	// 提取每层用于构建验证的两个节点（当前节点 + 兄弟节点）
	proofPairs := []map[string]interface{}{}
	index := targetIndex
	for level := 0; level < len(m.tree)-1; level++ {
		currentLevel := m.tree[level]
		siblingIndex := index ^ 1
		if siblingIndex >= len(currentLevel) {
			// 没有兄弟节点，不构建该层 proofPair
			index /= 2
			continue
		}

		currentNode := currentLevel[index]
		siblingNode := currentLevel[siblingIndex]

		proofPairs = append(proofPairs, map[string]interface{}{
			"level": level,
			"current": map[string]interface{}{
				"index": index,
				"hash":  fmt.Sprintf("0x%x", currentNode),
			},
			"sibling": map[string]interface{}{
				"index": siblingIndex,
				"hash":  fmt.Sprintf("0x%x", siblingNode),
			},
		})

		index /= 2
	}
	// 反转 proofPairs，使其自顶向下排序
	for i, j := 0, len(proofPairs)-1; i < j; i, j = i+1, j-1 {
		proofPairs[i], proofPairs[j] = proofPairs[j], proofPairs[i]
	}

	return proofPairs, true
}

// todo 这里没有实现
func (m *MerkleTree) Verify(targetHash []byte, proof interface{}) bool {
	return true
}
