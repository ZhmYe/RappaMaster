package PKI

import (
	"BHLayer2Node/paradigm"
	ecdsa_secp "github.com/consensys/gnark-crypto/ecc/secp256k1/ecdsa"
)

type PKIManager struct {
	// 节点秘钥管理
	nodeCert map[int]*paradigm.BHNodeKey
	// 主机秘钥管理
	sk ecdsa_secp.PrivateKey

	period int32
}

func NewPKIManager(config *paradigm.BHLayer2NodeConfig) *PKIManager {
	return &PKIManager{
		nodeCert: config.BHNodeKeyMap,
		sk:       config.HostPrivateKey,
		period:   100,
	}
}
