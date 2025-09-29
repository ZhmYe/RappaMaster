package paradigm

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	ecdsa_secp "github.com/consensys/gnark-crypto/ecc/secp256k1/ecdsa"
	"golang.org/x/crypto/sha3"
)

// CA 表示一个数字信任证书
// 由可信第三方生成
// 可信第三方有一个公开的公钥pk，和一个私有的私钥sk
// 公钥要上链
// 签名：对sha256(PublicKey, EpochLowerBound, EpochUpperBound)
// 这里没有包含bls12381，因为对于用户来说没有用
type CA struct {
	PublicKey       ecdsa_secp.PublicKey // 签名的公钥 todo 其实这个pk不用加，因为证书是配合公钥发过去，这里为了方便展示证书，就写成这样吧
	EpochLowerBound int32                // epoch的范围下界
	EpochUpperBound int32                // epoch的范围上界
	Signature       []byte               // 签名结果
	DecryptKey      ecdsa_secp.PublicKey // 第三方的pk，用于解密签名得到明文（即PublicKey, EpochLowerBound, EpochUpperBound)
}

// SerializableCA 用于JSON序列化的中间结构体
// 将 ecdsa.PublicKey 替换为 []byte
type SerializableCA struct {
	PublicKey       []byte `json:"public_key"`
	EpochLowerBound int32  `json:"epoch_lower_bound"`
	EpochUpperBound int32  `json:"epoch_upper_bound"`
	Signature       []byte `json:"signature"`
	DecryptKey      []byte `json:"decrypt_key"`
}

func (ca *CA) Marshal() SerializableCA {
	pubkeyBytes := ca.PublicKey.Bytes() // [64]位的bytes，x和y各256位

	decryptKeyBytes := ca.DecryptKey.Bytes()
	return SerializableCA{
		// 使用 elliptic.Marshal 将公钥转换为标准的字节格式
		PublicKey:       []byte(base64.StdEncoding.EncodeToString(pubkeyBytes[:])),
		EpochLowerBound: ca.EpochLowerBound,
		EpochUpperBound: ca.EpochUpperBound,
		Signature:       []byte(base64.StdEncoding.EncodeToString(ca.Signature)),
		DecryptKey:      []byte(base64.StdEncoding.EncodeToString(decryptKeyBytes[:])),
	}
}

func (ca *CA) Unmarshal(sca SerializableCA) error {
	var publicKey ecdsa_secp.PublicKey
	// 解码base64字符串到字节数组
	publicKeyBytes, err := base64.StdEncoding.DecodeString(string(sca.PublicKey))
	if err != nil {
		return err
	}
	_, err = publicKey.SetBytes(publicKeyBytes)
	if err != nil {
		return err
	}
	ca.PublicKey = publicKey

	var decryptKey ecdsa_secp.PublicKey
	// 解码base64字符串到字节数组
	decryptKeyBytes, err := base64.StdEncoding.DecodeString(string(sca.DecryptKey))
	if err != nil {
		return err
	}
	_, err = decryptKey.SetBytes(decryptKeyBytes)
	if err != nil {
		return err
	}
	ca.DecryptKey = decryptKey
	ca.EpochLowerBound = sca.EpochLowerBound
	ca.EpochUpperBound = sca.EpochUpperBound
	ca.Signature, err = base64.StdEncoding.DecodeString(string(sca.Signature))
	if err != nil {
		return err
	}
	return nil

}

// Verify 用来验证这个ca内容
func (ca *CA) Verify() error {

	hasher := sha3.New256()

	pubkeyBytes := ca.PublicKey.Bytes() // [64]位的bytes，x和y各256位
	hasher.Write(pubkeyBytes[:])
	e1Bytes, err := ConvertIntToBytes(ca.EpochLowerBound)
	if err != nil {
		return err
	}
	e2Bytes, err := ConvertIntToBytes(ca.EpochUpperBound)
	if err != nil {
		return err
	}
	hasher.Write(e1Bytes)
	hasher.Write(e2Bytes)
	plainText := hasher.Sum(nil) // 要签名的明文

	if flag, err := ca.DecryptKey.Verify(ca.Signature, plainText, sha3.New256()); err != nil {
		return err
	} else {
		if !flag {
			return fmt.Errorf("invalid signature")
		}
	}
	return nil
}

// 验证是否过期
func (ca *CA) VerifyExpire(epoch int32) error {
	// 验证是否过期
	if epoch < ca.EpochLowerBound || epoch > ca.EpochUpperBound {
		return fmt.Errorf("expire ca")
	}
	return nil
}

func ConvertIntToBytes(num int32) ([]byte, error) {
	buf := new(bytes.Buffer)                        // 使用bytes.Buffer来收集写入的字节
	err := binary.Write(buf, binary.BigEndian, num) // 使用big endian格式写入
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
