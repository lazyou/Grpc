package util

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/grpc/peer"
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

// ClientIP 获取请求客户端的远程地址 (通过从metadata中获取远程地址信息)
func ClientIP(ctx context.Context) string {
	pr, ok := peer.FromContext(ctx)
	if !ok {
		fmt.Errorf("[getClinetIP] invoke FromContext() failed \n")
		return ""
	}

	if pr.Addr == net.Addr(nil) {
		fmt.Errorf("[getClientIP] peer.Addr is nil \n")
		return ""
	}

	addStr := pr.Addr.String()
	fmt.Printf("ClientIP: %s \n", addStr)

	addSlice := strings.Split(addStr, ":")
	// 本机地址. TODO: 就算是本机地址, IP 一样, 但是端口号可能不一样, 所以如果获取 IP 这样返回是没错的, 如果要区分同一个 IP 下的请求, 还要再加上请求的 端口号
	if addSlice[0] == "[" {
		return "localhost"
	}

	return addSlice[0]
}
