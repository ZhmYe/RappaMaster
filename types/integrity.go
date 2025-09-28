package types

import (
	"RappaMaster/crypto"
	pb "RappaMaster/pb/service"
	"encoding/hex"
	"slices"
	"sort"
)

// EpochIntegrityEvidence 某个节点在当前epoch需要验证的内容
type EpochIntegrityEvidence struct {
	nodeID          NodeID
	epochRoot       []byte               // 整个epoch的root，节点最终需要对这个root进行bls12381聚合签名
	tasks           []string             // 传入的sign，已排序
	taskMerkleProof []crypto.MerkleProof // 和tasks一一对应，默克尔证明
	slots           [][]string           // 和tasks一一对应，该节点在每个task中的slot哈希
}

func (eie *EpochIntegrityEvidence) NodeID() NodeID { return eie.nodeID }
func (eie *EpochIntegrityEvidence) Slots() []string {
	res := make([]string, 0)
	for _, s := range eie.slots {
		res = append(res, s...)
	}
	return res
}

type EpochTree struct {
	nodes      map[string]TaskTree
	merkleTree *crypto.MerkleTree[*EpochTree] // 内置Merkle树
}

// NewEpochTree 初始化EpochTree
func NewEpochTree() *EpochTree {
	return &EpochTree{
		nodes:      make(map[string]TaskTree),
		merkleTree: nil,
	}
}

func (et *EpochTree) keys() []string {
	keys := make([]string, 0, len(et.nodes))
	for k := range et.nodes {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	return keys
}

func (et *EpochTree) Update(slot *pb.SlotCommitRequest) error {
	sign := slot.Sign
	if _, ok := et.nodes[sign]; !ok {
		et.nodes[sign] = NewTaskTree()
	}
	tt := et.nodes[sign]
	if err := tt.update(slot); err != nil {
		return err
	}
	et.nodes[sign] = tt
	return nil
}

// Prune 简直，出现在et无法justified的时候
func (et *EpochTree) Prune(nodeID NodeID) {
	et.merkleTree = nil // 要重新计算
	for taskID, tt := range et.nodes {
		tt.delete(nodeID)
		et.nodes[taskID] = tt
	}
}

// delete 调用这个函数的时候
func (et *EpochTree) delete(nodeID NodeID) {

}

func (et *EpochTree) HashableNodes() [][]byte {
	ids := et.keys()
	res := make([][]byte, 0, len(ids))
	for _, taskID := range ids {
		tt := et.nodes[taskID]
		res = append(res, tt.Root())
	}
	return res
}

func (et *EpochTree) build() {
	if et.merkleTree == nil {
		et.merkleTree = crypto.NewMerkleTree[*EpochTree](et)
	}
}

func (et *EpochTree) Root() []byte {
	et.build()
	return et.merkleTree.Root()
}

// MerkleProof 生成指定任务的默克尔证明
func (et *EpochTree) MerkleProof(taskIndex int) crypto.MerkleProof {
	et.build()
	return et.merkleTree.MerkleProof(taskIndex)
}

// Evidences 为每个节点生成其在当前epoch需要验证的内容
func (et *EpochTree) Evidences() []EpochIntegrityEvidence {
	// 确保Merkle树已构建,taskTree和nodeTree也已经构建好了
	et.build()

	// 准备任务ID和对应的签名
	tasks := et.keys()

	epochRoot := et.Root()

	// 收集所有节点ID
	nodeIDs := make(map[NodeID]struct{})
	for _, task := range et.nodes {
		for nodeID := range task.nodes {
			nodeIDs[nodeID] = struct{}{}
		}
	}

	// 为每个节点构建证据
	evidences := make([]EpochIntegrityEvidence, 0)
	for nodeID := range nodeIDs {

		proofs := make([]crypto.MerkleProof, 0)
		slots := make([][]string, 0)

		// 为每个任务构建证明和收集slot哈希
		for _, sign := range tasks {
			tt := et.nodes[sign]
			leaves := tt.keys()
			index := slices.Index(leaves, nodeID)
			if index != -1 {
				// 收集该节点在该任务中的所有slot哈希
				nt := tt.nodes[nodeID]
				proofs = append(proofs, nt.MerkleProof(index))
				slots = append(slots, nt.slotHashes())
			} else {
				// 节点未参与该任务，添加空列表
				proofs = append(proofs, crypto.MerkleProof{})
				slots = append(slots, []string{})
			}
		}

		evidences = append(evidences, EpochIntegrityEvidence{
			nodeID:          nodeID,
			epochRoot:       epochRoot,
			tasks:           tasks,
			taskMerkleProof: proofs,
			slots:           slots,
		})
	}

	return evidences
}

func (et *EpochTree) Clear() {
	et.nodes = make(map[string]TaskTree)
	et.merkleTree = nil
}

type TaskTree struct {
	nodes      map[NodeID]NodeTree
	merkleTree *crypto.MerkleTree[*TaskTree] // 内置Merkle树
}

func NewTaskTree() TaskTree {
	return TaskTree{
		nodes:      make(map[NodeID]NodeTree),
		merkleTree: nil,
	}
}

func (tt *TaskTree) keys() []NodeID {
	keys := make([]NodeID, 0, len(tt.nodes))
	for k := range tt.nodes {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	return keys
}

func (tt *TaskTree) update(slot *pb.SlotCommitRequest) error {
	nodeID := NodeID(slot.NodeID)
	if _, ok := tt.nodes[nodeID]; !ok {
		tt.nodes[nodeID] = NewNodeTree()
	}
	nt := tt.nodes[nodeID]
	if err := nt.update(slot); err != nil {
		return err
	}
	tt.nodes[nodeID] = nt
	return nil
}
func (tt *TaskTree) delete(nodeID NodeID) {
	if _, ok := tt.nodes[nodeID]; !ok {
		return
	}
	delete(tt.nodes, nodeID)
}
func (tt *TaskTree) HashableNodes() [][]byte {
	ids := tt.keys()
	res := make([][]byte, 0, len(ids))
	for _, nodeID := range ids {
		nt := tt.nodes[nodeID]
		res = append(res, nt.Root())
	}
	return res
}

// 确保Merkle树已初始化并更新
func (tt *TaskTree) build() {
	if tt.merkleTree == nil {
		tt.merkleTree = crypto.NewMerkleTree[*TaskTree](tt)
	}
}

func (tt *TaskTree) Root() []byte {
	tt.build()
	return tt.merkleTree.Root()
}

// MerkleProof 生成指定节点的默克尔证明
func (tt *TaskTree) MerkleProof(nodeIndex int) crypto.MerkleProof {
	tt.build()
	return tt.merkleTree.MerkleProof(nodeIndex)
}

type NodeTree struct {
	slots      []*pb.SlotCommitRequest
	merkleTree *crypto.MerkleTree[*NodeTree] // 内置Merkle树
}

func NewNodeTree() NodeTree {
	return NodeTree{
		slots:      make([]*pb.SlotCommitRequest, 0),
		merkleTree: nil,
	}
}

func (nt *NodeTree) update(slot *pb.SlotCommitRequest) error {
	// 检查是否已存在相同的slot
	for _, s := range nt.slots {
		if hex.EncodeToString(s.Commitment) == hex.EncodeToString(slot.Commitment) {
			return nil // 已存在，无需重复添加
		}
	}
	nt.slots = append(nt.slots, slot)
	return nil
}

func (nt *NodeTree) HashableNodes() [][]byte {
	// 按Commitment排序
	sort.Slice(nt.slots, func(i, j int) bool {
		return hex.EncodeToString(nt.slots[i].Commitment) < hex.EncodeToString(nt.slots[j].Commitment)
	})

	res := make([][]byte, len(nt.slots))
	for i := 0; i < len(nt.slots); i++ {
		res[i] = nt.slots[i].Commitment
	}
	return res
}

func (nt *NodeTree) build() {
	if nt.merkleTree == nil {
		nt.merkleTree = crypto.NewMerkleTree[*NodeTree](nt)
	}
}

func (nt *NodeTree) Root() []byte {
	nt.build()
	return nt.merkleTree.Root()
}

// MerkleProof 生成指定slot的默克尔证明
func (nt *NodeTree) MerkleProof(slotIndex int) crypto.MerkleProof {
	nt.build()
	return nt.merkleTree.MerkleProof(slotIndex)
}

func (nt *NodeTree) slotHashes() []string {
	res := make([]string, len(nt.slots))
	for i := 0; i < len(nt.slots); i++ {
		res[i] = nt.slots[i].SlotHash
	}
	return res
}
