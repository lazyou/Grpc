package main

import (
	"log"
	"net"

	"google.golang.org/grpc"
	"test/protobuf"
	"test/services"
)

const (
	port = ":50051"
)

/* gRPC 服务注册, 监听服务 start */
func main() {
	// 监听本地网络端口
	listen, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	} else {
		log.Printf("grpc server start %s", port)
	}

	// 服务端实现: 提供一个 gRPC 服务的另一个主要功能是让这个服务实在在网络上可用
	s := grpc.NewServer()
	protobuf.RegisterHelloServer(s, &services.HelloServer{}) // 注册服务

	if err := s.Serve(listen); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

/* gRPC 服务注册, 监听服务 end */
