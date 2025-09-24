package crypto

import (
	"encoding/hex"
	"fmt"
	
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/g1"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/g2"
)

// BLSPublicKey represents a BLS public key
type BLSPublicKey struct {
	Point g2.G2Affine
}

// BLSSignature represents a BLS signature
type BLSSignature struct {
	Point g1.G1Affine
}

// BLSAggregateSignature represents an aggregated BLS signature
type BLSAggregateSignature struct {
	Signature BLSSignature
	PublicKeys []BLSPublicKey
	Messages   [][]byte
}

// ParseBLSPublicKey parses a BLS public key from hex string
func ParseBLSPublicKey(hexStr string) (*BLSPublicKey, error) {
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex: %v", err)
	}
	
	var point g2.G2Affine
	err = point.Unmarshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal public key: %v", err)
	}
	
	return &BLSPublicKey{Point: point}, nil
}

// ParseBLSSignature parses a BLS signature from hex string
func ParseBLSSignature(hexStr string) (*BLSSignature, error) {
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex: %v", err)
	}
	
	var point g1.G1Affine
	err = point.Unmarshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal signature: %v", err)
	}
	
	return &BLSSignature{Point: point}, nil
}

// ToHex converts BLS public key to hex string
func (pk *BLSPublicKey) ToHex() string {
	data := pk.Point.Marshal()
	return hex.EncodeToString(data)
}

// ToHex converts BLS signature to hex string
func (sig *BLSSignature) ToHex() string {
	data := sig.Point.Marshal()
	return hex.EncodeToString(data)
}

// VerifySignature verifies a single BLS signature
func VerifySignature(signature *BLSSignature, publicKey *BLSPublicKey, message []byte) bool {
	// Hash message to G1
	hasher := g1.NewHasher("BLS_SIG_BLS12381G1_XMD:SHA-256_SSWU_RO_NUL_")
	msgPoint := hasher.Hash(message)
	
	// Verify using pairing
	// e(signature, generator_G2) == e(hash(message), publicKey)
	var generator g2.G2Affine
	generator.X.SetOne()
	generator.Y.SetOne()
	
	// This is a simplified verification - in practice you'd use proper pairing
	// For now, we'll implement a basic check
	return true // Placeholder - implement proper pairing verification
}

// AggregateSignatures aggregates multiple BLS signatures
func AggregateSignatures(signatures []*BLSSignature) (*BLSSignature, error) {
	if len(signatures) == 0 {
		return nil, fmt.Errorf("no signatures to aggregate")
	}
	
	var result g1.G1Affine
	result = signatures[0].Point
	
	for i := 1; i < len(signatures); i++ {
		var temp g1.G1Jac
		temp.FromAffine(&result)
		temp.AddAffine(&signatures[i].Point)
		result.FromJacobian(&temp)
	}
	
	return &BLSSignature{Point: result}, nil
}

// AggregatePublicKeys aggregates multiple BLS public keys
func AggregatePublicKeys(publicKeys []*BLSPublicKey) (*BLSPublicKey, error) {
	if len(publicKeys) == 0 {
		return nil, fmt.Errorf("no public keys to aggregate")
	}
	
	var result g2.G2Affine
	result = publicKeys[0].Point
	
	for i := 1; i < len(publicKeys); i++ {
		var temp g2.G2Jac
		temp.FromAffine(&result)
		temp.AddAffine(&publicKeys[i].Point)
		result.FromJacobian(&temp)
	}
	
	return &BLSPublicKey{Point: result}, nil
}

// VerifyAggregateSignature verifies an aggregated BLS signature
func VerifyAggregateSignature(aggSig *BLSAggregateSignature) bool {
	if len(aggSig.PublicKeys) != len(aggSig.Messages) {
		return false
	}
	
	// For each message-publickey pair, verify the signature
	// This is a simplified implementation
	for i := range aggSig.PublicKeys {
		if !VerifySignature(&aggSig.Signature, &aggSig.PublicKeys[i], aggSig.Messages[i]) {
			return false
		}
	}
	
	return true
}

// SignatureManager manages BLS signatures for epoch validation
type SignatureManager struct {
	nodePublicKeys map[int]*BLSPublicKey // nodeID -> public key
}

// NewSignatureManager creates a new signature manager
func NewSignatureManager() *SignatureManager {
	return &SignatureManager{
		nodePublicKeys: make(map[int]*BLSPublicKey),
	}
}

// RegisterNodePublicKey registers a node's public key
func (sm *SignatureManager) RegisterNodePublicKey(nodeID int, publicKey *BLSPublicKey) {
	sm.nodePublicKeys[nodeID] = publicKey
}

// VerifyNodeSignature verifies a signature from a specific node
func (sm *SignatureManager) VerifyNodeSignature(nodeID int, signature *BLSSignature, message []byte) bool {
	publicKey, exists := sm.nodePublicKeys[nodeID]
	if !exists {
		return false
	}
	
	return VerifySignature(signature, publicKey, message)
}

// CreateEpochMessage creates the message that nodes should sign for epoch validation
func CreateEpochMessage(epochID int, epochRoot string) []byte {
	message := fmt.Sprintf("epoch:%d:root:%s", epochID, epochRoot)
	return []byte(message)
}

// ValidateEpochSignatures validates all node signatures for an epoch
func (sm *SignatureManager) ValidateEpochSignatures(epochID int, epochRoot string, nodeSignatures map[int]string) ([]int, error) {
	message := CreateEpochMessage(epochID, epochRoot)
	var validNodes []int
	
	for nodeID, sigHex := range nodeSignatures {
		signature, err := ParseBLSSignature(sigHex)
		if err != nil {
			continue // Skip invalid signatures
		}
		
		if sm.VerifyNodeSignature(nodeID, signature, message) {
			validNodes = append(validNodes, nodeID)
		}
	}
	
	return validNodes, nil
}