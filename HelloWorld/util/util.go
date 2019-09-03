package util

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// GetAppPath 获取app运行路径
func GetAppPath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	index := strings.LastIndex(path, string(os.PathSeparator))

	return path[:index]
}

// NowTime 获取当前时间
func NowTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
