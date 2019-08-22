package util

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GetAppPath 获取app运行路径
func GetAppPath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	index := strings.LastIndex(path, string(os.PathSeparator))

	return path[:index]
}
