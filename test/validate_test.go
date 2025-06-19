package test

import (
	"BHLayer2Node/Validatable"
	"crypto/rand"
	"fmt"
	"github.com/goccy/go-json"
	"testing"
)

func generateRandomData(count int) [][]byte {
	data := make([][]byte, count)
	for i := 0; i < count; i++ {
		b := make([]byte, 32)
		rand.Read(b) // 填充随机字节
		data[i] = b
	}
	return data
}

func TestNewValidatable_Merkle(t *testing.T) {
	data := generateRandomData(20)

	v, err := Validatable.NewValidatable(data, Validatable.Merkle)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if v == nil {
		t.Fatal("expected non-nil Validatable for Merkle")
	}

	// 获取第一个叶子的 proof
	proof, ok := v.GetProof(0)
	if !ok {
		t.Fatalf("expected proof for index 0")
	}

	targetHash := data[0]

	fmt.Println(json.Marshal(proof))

	// 验证 proof
	if !v.Verify(targetHash, proof) {
		t.Errorf("expected valid proof for leaf 0")
	}
}

func TestNewValidatable_Verkle(t *testing.T) {
	data := generateRandomData(20)

	v, err := Validatable.NewValidatable(data, Validatable.Verkle)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if v == nil {
		t.Fatal("expected non-nil Validatable for Verkle")
	}

	// 同样进行 GetProof 和 Verify 测试
	proof, ok := v.GetProof(0)
	if !ok {
		t.Fatalf("expected proof for index 0")
	}

	str, _ := json.Marshal(proof)
	fmt.Println(string(str))

	targetHash := data[0]
	if !v.Verify(targetHash, proof) {
		t.Errorf("expected valid proof for leaf 0")
	}
}

func TestNewValidatable_UnsupportedType(t *testing.T) {
	data := [][]byte{
		[]byte("leaf1"),
	}

	invalidType := Validatable.StructType(99) // 不支持的类型
	_, err := Validatable.NewValidatable(data, invalidType)
	if err == nil {
		t.Errorf("expected error for unsupported type, got nil")
	}
}
