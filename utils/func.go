package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
