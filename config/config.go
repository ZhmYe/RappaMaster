package config

import (
	"os"
	"path/filepath"
)

var (
	ProjectRootPath string = ""
)

func init() {
	// 获取当前工作目录
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	for {
		// 构造标识文件的完整路径
		markerPath := filepath.Join(dir, ".project_root")
		// 检查标识文件是否存在
		if _, err := os.Stat(markerPath); err == nil {
			// 找到标识文件，返回当前目录
			ProjectRootPath = dir
			return
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
	panic(os.ErrNotExist)
}
