package utils

import (
	"encoding/json"
	"math/big"
)

func StringToBytes32(s string) [32]byte {
	var b [32]byte
	copy(b[:], s)
	return b
}

// serializeParams 将 map[string]interface{} 转化为[32]byte
func SerializeParams(params map[string]interface{}) [32]byte {
	bytes, _ := json.Marshal(params)
	return StringToBytes32(string(bytes))
}

// BigIntToBytes32 将 *big.Int 转换为32字节的切片（左侧填充0）
func BigIntToBytes32(n *big.Int) []byte {
	b := n.Bytes() // 返回大端序的最小字节切片
	if len(b) > 32 {
		// 如果超出32字节，取最后32字节（通常不应该发生）
		return b[len(b)-32:]
	}
	// 否则左侧填充0
	padded := make([]byte, 32)
	copy(padded[32-len(b):], b)
	return padded
}

func FlattenByte32Slice(arr [][32]byte) []byte {
	// 每个元素都是 32 字节
	totalBytes := len(arr) * 32
	// 提前分配足够容量，减少内存拷贝
	result := make([]byte, 0, totalBytes)
	for _, v := range arr {
		// v[:] 把 [32]byte 转成 []byte，然后 ... 将其展开
		result = append(result, v[:]...)
	}
	return result
}

//// BinarySearch 二分查找
//func BinarySearch(arr []interface{}, target interface{}, compare func(element interface{}, target interface{}) bool) int {
//	left, right := 0, len(arr)-1
//	result := -1
//	// 迭代直到范围为空
//	for left <= right {
//		mid := (left + right) / 2
//		if compare(arr[mid], target) {
//			// 说明mid符合范围
//			result = mid
//		}
//
//		if comparisonResult == 0 {
//			// 找到目标值
//			return mid
//		} else if comparisonResult < 0 {
//			// 目标值在右边
//			left = mid + 1
//		} else {
//			// 目标值在左边
//			right = mid - 1
//		}
//	}
//
//	// 如果没有找到，返回 -1
//	return -1
//}
