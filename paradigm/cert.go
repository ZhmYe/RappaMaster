package paradigm

import (
	"BHLayer2Node/pb/service"
	"encoding/base64"
	"fmt"

	ecdsa_bls12381 "github.com/consensys/gnark-crypto/ecc/bls12-381/ecdsa"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fp"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	ecdsa_secp "github.com/consensys/gnark-crypto/ecc/secp256k1/ecdsa"
)

// BLS12381的bls签名和聚合方式
// 这里密钥的格式是一样的但是不能用ecdsa的算法

const (
	sizeFr         = fr.Bytes
	sizeFrBits     = fr.Bits
	sizeFp         = fp.Bytes
	sizePublicKey  = sizeFp
	sizePrivateKey = sizeFr + sizePublicKey
	sizeSignature  = 2 * sizeFr
)

type BLS12381PublicKey = ecdsa_bls12381.PublicKey

// BLS12381PrivateKey 私钥
type BLS12381PrivateKey struct {
	PublicKey BLS12381PublicKey // 公钥，[s]g1，一个base field上的点
	scalar    [sizeFr]byte      // [s], scalar filed上的点，一个big.int
}

type NodeCertification struct {
	NodeID int32
	SpKey  ecdsa_secp.PublicKey
	Bpkey  ecdsa_bls12381.PublicKey
}

type SignedCommitSlotItem struct {
	*CommitSlotItem
	//签名，需要验签
	signature string
	//节点证书
	ca string
}

func (s *SignedCommitSlotItem) GetSlotSignature() string {
	return s.signature
}

func (s *SignedCommitSlotItem) GetNodeCA() string {
	return s.ca
}

func NewSignedCommitSlot(slot *service.JustifiedSlot, sign string, ca string) SignedCommitSlotItem {
	cs := NewCommitSlotItem(slot)
	//s.hash = s.Hash // todo 这里简单这样写一下
	//s.computeHash()
	return SignedCommitSlotItem{
		CommitSlotItem: &cs,
		signature:      sign,
		ca:             ca,
	}
}

func NewSignedCommitSlotFrom(slot *CommitSlotItem, sign string, ca string) SignedCommitSlotItem {
	//s.hash = s.Hash // todo 这里简单这样写一下
	//s.computeHash()
	return SignedCommitSlotItem{
		CommitSlotItem: slot,
		signature:      sign,
		ca:             ca,
	}
}

func DecodeBLS12381PublicKey(key string) (BLS12381PublicKey, error) {
	blsKeyBytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return BLS12381PublicKey{}, fmt.Errorf("failed to decode bls12381 key: %w", err)
	}
	var blsKey BLS12381PublicKey
	if _, err := blsKey.SetBytes(blsKeyBytes); err != nil {
		return BLS12381PublicKey{}, fmt.Errorf("failed to set bls12381 key bytes: %w", err)
	}
	return blsKey, nil
}

func DecodeSecpPublicKey(key string) (ecdsa_secp.PublicKey, error) {
	secpKeyBytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return ecdsa_secp.PublicKey{}, fmt.Errorf("failed to decode secp256k1 key: %w", err)
	}

	var secpKey ecdsa_secp.PublicKey
	if _, err := secpKey.SetBytes(secpKeyBytes); err != nil {
		return ecdsa_secp.PublicKey{}, fmt.Errorf("failed to set secp256k1 key bytes: %w", err)
	}
	return secpKey, nil
}
