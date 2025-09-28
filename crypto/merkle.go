package crypto

import (
	"golang.org/x/crypto/sha3"
)

type Hashable interface {
	HashableNodes() [][]byte
}
type MerkleProof struct {
	root []byte
	path [][]byte
}

// MerkleTree 用泛型来统一写叶子节点，包含缓存的树结构
type MerkleTree[T Hashable] struct {
	data   T
	root   []byte     // 缓存的根哈希
	levels [][][]byte // 缓存的树层级，levels[0]是叶子节点，levels[len-1]是根节点
}

// build 构建Merkle树并缓存层级信息
func (mt *MerkleTree[T]) build() {
	// 如果已经构建过树，直接返回
	if len(mt.levels) > 0 {
		return
	}

	// 从叶子节点开始构建
	leaves := mt.data.HashableNodes()
	if len(leaves) == 0 {
		mt.levels = [][][]byte{}
		mt.root = nil
		return
	}

	// 初始化层级，第一层是叶子节点
	mt.levels = [][][]byte{leaves}
	currentLevel := leaves

	// 逐层构建直到根节点
	for len(currentLevel) > 1 {
		nextLevel := make([][]byte, 0, (len(currentLevel)+1)/2)
		for i := 0; i < len(currentLevel); i += 2 {
			left := currentLevel[i]
			right := left // 处理奇数情况
			if i+1 < len(currentLevel) {
				right = currentLevel[i+1]
			}

			// 计算父节点哈希
			hasher := sha3.New256()
			hasher.Write(left)
			hasher.Write(right)
			nextLevel = append(nextLevel, hasher.Sum(nil))
		}

		currentLevel = nextLevel
		mt.levels = append(mt.levels, currentLevel)
	}

	// 缓存根节点
	if len(mt.levels) > 0 {
		mt.root = mt.levels[len(mt.levels)-1][0]
	}
}

// Root 返回Merkle树的根哈希
func (mt *MerkleTree[T]) Root() []byte {
	mt.build()
	return mt.root
}

// MerkleProof 生成指定索引叶子节点的默克尔证明
// 参数index为叶子节点的索引（从0开始）
// 返回值为证明路径，包含从叶子到根过程中所需的所有兄弟节点哈希
func (mt *MerkleTree[T]) MerkleProof(index int) MerkleProof {
	// 确保树已经构建
	mt.build()

	// 检查叶子节点是否存在
	if len(mt.levels) == 0 || index < 0 || index >= len(mt.levels[0]) {
		return MerkleProof{
			root: nil,
			path: nil,
		} // 索引无效，返回空证明
	}

	proof := make([][]byte, 0, len(mt.levels)-1)
	currentIndex := index

	// 从叶子节点层开始向上收集证明
	for i := 0; i < len(mt.levels)-1; i++ {
		currentLevel := mt.levels[i]
		var siblingIndex int

		// 计算兄弟节点索引
		if currentIndex%2 == 0 {
			// 当前节点是左孩子，兄弟是右孩子
			siblingIndex = currentIndex + 1
			// 检查兄弟节点是否存在（处理奇数个节点的情况）
			if siblingIndex >= len(currentLevel) {
				siblingIndex = currentIndex // 没有右兄弟，使用自身作为兄弟
			}
		} else {
			// 当前节点是右孩子，兄弟是左孩子
			siblingIndex = currentIndex - 1
		}

		// 将兄弟节点的哈希添加到证明中
		proof = append(proof, currentLevel[siblingIndex])

		// 计算父节点索引，准备上移到下一层
		currentIndex = currentIndex / 2
	}

	return MerkleProof{root: mt.Root(), path: proof}
}

// NewMerkleTree 创建一个新的Merkle树实例
func NewMerkleTree[T Hashable](data T) *MerkleTree[T] {
	return &MerkleTree[T]{data: data}
}
