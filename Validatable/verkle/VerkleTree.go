package verkle

import "BHLayer2Node/paradigm"

var baseValue = []byte("0123456789abcdef0123456789abcdef")

type VerkleTree struct {
	cfg  *Config
	root VerkleNode
	data [][]byte
}

// 这里定义便于序列化
type VerkleSerializedProof struct {
	Vp   *VerkleProof
	Diff StateDiff
}

func (v VerkleTree) Build(leaves [][]byte) error {
	v.root = New()
	for _, leaf := range leaves {
		err := v.root.Insert(leaf, baseValue, nil)
		if err != nil {
			return err
		}
	}
	v.data = leaves
	v.cfg = GetConfig()
	v.root.Commit()
	return nil
}

// 这里我们假定只去验证一个叶子节点
func (v VerkleTree) GetProof(targetIndex int) (interface{}, bool) {
	dataHash := v.data[targetIndex]
	proof, _, _, _, err := MakeVerkleMultiProof(v.root, nil, [][]byte{dataHash}, nil)
	if err != nil {
		paradigm.Error(paradigm.RuntimeError, err.Error())
		return nil, false
	}
	vp, statediff, err := SerializeProof(proof)
	if err != nil {
		paradigm.Error(paradigm.RuntimeError, err.Error())
		return nil, false
	}
	return VerkleSerializedProof{
		Vp:   vp,
		Diff: statediff,
	}, true
}

// 这里的Proof传入的是VerkleProof
func (v VerkleTree) Verify(targetIndex int, proof interface{}) bool {
	serializedProof := proof.(VerkleSerializedProof)
	// 这一步进行反序列化
	verkleProof, err := DeserializeProof(serializedProof.Vp, serializedProof.Diff)
	dataHash := v.data[targetIndex]
	pe, _, _, err := GetCommitmentsForMultiproof(v.root, [][]byte{dataHash}, nil)
	if err != nil {
		paradigm.Error(paradigm.RuntimeError, err.Error())
		return false
	}
	if ok, _ := verifyVerkleProof(verkleProof, pe.Cis, pe.Zis, pe.Yis, v.cfg); ok {
		return true
	}
	return false
}
