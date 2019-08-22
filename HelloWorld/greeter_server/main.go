//go:generate protoc -I ../helloworld --go_out=plugins=grpc:../helloworld ../helloworld/helloworld.proto

// Package main implements a server for Greeter service.
package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	pb "test/helloworld"
)

const (
	port = ":50051"
)

/* 服务实现 start */
// server is used to implement helloworld.GreeterServer.
type HelloServer struct {
	pb.GreeterServer // PS: struct 包含(内嵌) interface 之后，并不需要实现 interface 的接口，也能成为 interface 接口类
}

// SayHello implements helloworld.GreeterServer
// 实现 GreeterServer 服务: 接收 HelloRequest, 返回 HelloReply.
func (s *HelloServer) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.Name)
	return &pb.HelloReply{
		Message: "Hello " + in.Name,
	}, nil
}

/* 服务实现 end */

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
	pb.RegisterGreeterServer(s, &HelloServer{}) // 注册服务

	if err := s.Serve(listen); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

/* gRPC 服务注册, 监听服务 end */
