package PKI

import (
	"BHLayer2Node/paradigm"
	"crypto/sha256"
	"encoding/base64"
	ecdsa_secp "github.com/consensys/gnark-crypto/ecc/secp256k1/ecdsa"
	"github.com/goccy/go-json"
	"golang.org/x/crypto/sha3"
)

// 密钥管理
type PKIManager struct {
	// 节点秘钥管理
	nodeCert map[int]*paradigm.BHNodeKey
	// 主机秘钥管理
	sk ecdsa_secp.PrivateKey
	// 节点证书
	nodeCAMap map[int]string
	// 证书过期时间周期
	period int32
}

func NewPKIManager(config *paradigm.BHLayer2NodeConfig) *PKIManager {
	return &PKIManager{
		nodeCert:  config.BHNodeKeyMap,
		sk:        config.HostPrivateKey,
		nodeCAMap: make(map[int]string),
		period:    100,
	}
}

func (pm *PKIManager) VerifyNodeCA(nodeId int32, currentEpoch int32, caBase64 string) bool {
	//检查证书是否存在,如果相等则直接返回true
	if pm.nodeCAMap[int(nodeId)] == caBase64 {
		return true
	}
	//caBytes反序列化
	caBytes, err := base64.StdEncoding.DecodeString(caBase64)
	if err != nil {
		return false
	}
	serialCA := paradigm.SerializableCA{}
	err = json.Unmarshal(caBytes, &serialCA)
	if err != nil {
		return false
	}
	ca := &paradigm.CA{}
	err = ca.Unmarshal(serialCA)
	if err != nil {
		return false
	}
	// 验证ca是否合法
	err = ca.Verify()
	if err != nil {
		return false
	}
	// 验证ca是否是本机和节点公钥
	if ca.DecryptKey != pm.sk.PublicKey && ca.PublicKey != pm.nodeCert[int(nodeId)].SecpKey {
		return false
	}
	// 验证ca是否过期
	err = ca.VerifyExpire(currentEpoch)
	if err != nil {
		return false
	}
	// 更新证书
	pm.nodeCAMap[int(nodeId)] = caBase64
	return true
}

// 验证节点签名是否正确
func (pm *PKIManager) VertifyNodeSign(nodeId int, data string, signBase64 string) bool {
	//signBytes反序列化
	signBytes, err := base64.StdEncoding.DecodeString(signBase64)
	if err != nil {
		return false
	}
	// executor 侧当前签名对象为 sha256(data) 的 32 字节摘要，这里保持同一语义
	dataHash := sha256.Sum256([]byte(data))
	// 验证签名
	nodePK := pm.nodeCert[nodeId].SecpKey
	flag, _ := nodePK.Verify(signBytes, dataHash[:], sha3.New256())
	// 验证签名是否正确
	return flag
}
