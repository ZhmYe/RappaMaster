package utils

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
)

// GetProjectRoot 查找包含特定标识文件的项目根目录
func GetProjectRoot() (string, error) {
	// 获取当前工作目录
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		// 构造标识文件的完整路径
		markerPath := filepath.Join(dir, ".project_root")
		fmt.Println(markerPath)
		// 检查标识文件是否存在
		if _, err := os.Stat(markerPath); err == nil {
			// 找到标识文件，返回当前目录
			return dir, nil
		}

		// 获取父目录
		parentDir := filepath.Dir(dir)

		// 如果已经到根目录，退出循环
		if parentDir == dir {
			break
		}

		// 继续搜索父目录
		dir = parentDir
	}

	// 未找到标识文件，返回错误
	return "", os.ErrNotExist
}

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

// 把一个小于 256 的数字放在 bytes32 最低位
func IntToBytes32(i int) [32]byte {
	var b [32]byte
	b[31] = byte(i)
	return b
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

// trimZero 去掉右侧零字节，并转成 string
func TrimZero(b []byte) string {
	return strings.TrimRight(string(b), "\x00")
}

// hexList 把一组 byte-slice 转成 hex 字符串列表
func HexList(bsList [][]byte) []string {
	out := make([]string, len(bsList))
	for i, bs := range bsList {
		out[i] = "0x" + hex.EncodeToString(bs)
	}
	return out
}

// bytes32ListToStrings 把 [][32]byte 当成字符串，去掉尾部 \x00
func Bytes32ListToStrings(in [][32]byte) []string {
	out := make([]string, len(in))
	for i, v := range in {
		// 直接把字节切片转成 string 然后去尾部 \x00
		out[i] = strings.TrimRight(string(v[:]), "\x00")
	}
	return out
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
