package merkle

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
)

// MerkleNode represents a node in the merkle tree
type MerkleNode struct {
	Hash      string
	Left      *MerkleNode
	Right     *MerkleNode
	IsLeaf    bool
	Data      interface{} // original data for leaf nodes
}

// MerkleTree represents the complete merkle tree
type MerkleTree struct {
	Root   *MerkleNode
	Leaves []*MerkleNode
}

// MerkleProof represents a merkle proof path
type MerkleProof struct {
	Path      []string `json:"path"`      // hashes along the path
	Positions []bool   `json:"positions"` // true for right, false for left
	Root      string   `json:"root"`      // root hash
}

// TaskData represents task data for merkle tree construction
type TaskData struct {
	TaskSign string
	TaskRoot string
}

// Hash computes SHA256 hash of input string
func Hash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// CombineHashes combines two hashes into one
func CombineHashes(left, right string) string {
	return Hash(left + right)
}

// BuildSlotMerkleTree builds merkle tree for slots within a task
// slots parameter is a slice of slotHash strings
func BuildSlotMerkleTree(slotHashes []string) *MerkleTree {
	if len(slotHashes) == 0 {
		return nil
	}

	// Sort slot hashes for deterministic tree
	sort.Strings(slotHashes)

	// Create leaf nodes using slotHash directly
	var leaves []*MerkleNode
	for _, slotHash := range slotHashes {
		leaf := &MerkleNode{
			Hash:   slotHash, // Use slotHash directly as leaf hash
			IsLeaf: true,
			Data:   slotHash,
		}
		leaves = append(leaves, leaf)
	}

	// Build tree bottom-up
	root := buildTreeFromLeaves(leaves)
	
	return &MerkleTree{
		Root:   root,
		Leaves: leaves,
	}
}

// BuildTaskMerkleTree builds merkle tree for tasks within an epoch
func BuildTaskMerkleTree(tasks []TaskData) *MerkleTree {
	if len(tasks) == 0 {
		return nil
	}

	// Sort tasks by taskSign for deterministic tree
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].TaskSign < tasks[j].TaskSign
	})

	// Create leaf nodes
	var leaves []*MerkleNode
	for _, task := range tasks {
		leafData := fmt.Sprintf("%s:%s", task.TaskSign, task.TaskRoot)
		leafHash := Hash(leafData)
		leaf := &MerkleNode{
			Hash:   leafHash,
			IsLeaf: true,
			Data:   task,
		}
		leaves = append(leaves, leaf)
	}

	// Build tree bottom-up
	root := buildTreeFromLeaves(leaves)
	
	return &MerkleTree{
		Root:   root,
		Leaves: leaves,
	}
}

// buildTreeFromLeaves builds merkle tree from leaf nodes
func buildTreeFromLeaves(leaves []*MerkleNode) *MerkleNode {
	if len(leaves) == 0 {
		return nil
	}
	
	if len(leaves) == 1 {
		return leaves[0]
	}

	var nextLevel []*MerkleNode
	
	// Process pairs of nodes
	for i := 0; i < len(leaves); i += 2 {
		left := leaves[i]
		var right *MerkleNode
		
		if i+1 < len(leaves) {
			right = leaves[i+1]
		} else {
			// Odd number of nodes, duplicate the last one
			right = left
		}
		
		// Create parent node
		parentHash := CombineHashes(left.Hash, right.Hash)
		parent := &MerkleNode{
			Hash:   parentHash,
			Left:   left,
			Right:  right,
			IsLeaf: false,
		}
		
		nextLevel = append(nextLevel, parent)
	}
	
	// Recursively build upper levels
	return buildTreeFromLeaves(nextLevel)
}

// GenerateProof generates merkle proof for a specific leaf
func (mt *MerkleTree) GenerateProof(leafHash string) (*MerkleProof, error) {
	if mt.Root == nil {
		return nil, fmt.Errorf("empty tree")
	}

	// Find the leaf node
	var targetLeaf *MerkleNode
	for _, leaf := range mt.Leaves {
		if leaf.Hash == leafHash {
			targetLeaf = leaf
			break
		}
	}
	
	if targetLeaf == nil {
		return nil, fmt.Errorf("leaf not found in tree")
	}

	// Generate proof path
	var path []string
	var positions []bool
	
	current := targetLeaf
	err := mt.generateProofPath(mt.Root, current, &path, &positions)
	if err != nil {
		return nil, err
	}

	return &MerkleProof{
		Path:      path,
		Positions: positions,
		Root:      mt.Root.Hash,
	}, nil
}

// generateProofPath recursively generates proof path
func (mt *MerkleTree) generateProofPath(node, target *MerkleNode, path *[]string, positions *[]bool) error {
	if node == nil {
		return fmt.Errorf("node not found")
	}
	
	if node == target {
		return nil // Found target, stop recursion
	}
	
	if node.IsLeaf {
		return fmt.Errorf("target not found in this subtree")
	}
	
	// Check left subtree
	if mt.containsNode(node.Left, target) {
		// Target is in left subtree, add right sibling to proof
		*path = append(*path, node.Right.Hash)
		*positions = append(*positions, true) // right sibling
		return mt.generateProofPath(node.Left, target, path, positions)
	}
	
	// Check right subtree
	if mt.containsNode(node.Right, target) {
		// Target is in right subtree, add left sibling to proof
		*path = append(*path, node.Left.Hash)
		*positions = append(*positions, false) // left sibling
		return mt.generateProofPath(node.Right, target, path, positions)
	}
	
	return fmt.Errorf("target not found in subtree")
}

// containsNode checks if a subtree contains the target node
func (mt *MerkleTree) containsNode(root, target *MerkleNode) bool {
	if root == nil {
		return false
	}
	
	if root == target {
		return true
	}
	
	if root.IsLeaf {
		return false
	}
	
	return mt.containsNode(root.Left, target) || mt.containsNode(root.Right, target)
}

// VerifyProof verifies a merkle proof
func VerifyProof(leafHash string, proof *MerkleProof) bool {
	if proof == nil || len(proof.Path) != len(proof.Positions) {
		return false
	}
	
	currentHash := leafHash
	
	// Traverse proof path
	for i, siblingHash := range proof.Path {
		if proof.Positions[i] {
			// Sibling is on the right
			currentHash = CombineHashes(currentHash, siblingHash)
		} else {
			// Sibling is on the left
			currentHash = CombineHashes(siblingHash, currentHash)
		}
	}
	
	return currentHash == proof.Root
}

// GetRootHash returns the root hash of the tree
func (mt *MerkleTree) GetRootHash() string {
	if mt.Root == nil {
		return ""
	}
	return mt.Root.Hash
}