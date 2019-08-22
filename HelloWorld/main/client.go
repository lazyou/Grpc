// Package main implements a client for Greeter service.
package main

import (
	"context"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"test/protobuf"
)

const (
	address     = "localhost:50051"
	defaultName = "world"
)

func main() {
	// TLS连接
	credential, err := credentials.NewClientTLSFromFile("../keys/server.pem", "lin")
	if err != nil {
		grpclog.Fatalf("Failed to create TLS credentials %v", err)
	}

	// Set up a connection to the server.
	// 连接服务: 在 gRPC Go 是使用一个特殊的 Dial() 方法来创建频道
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(credential))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// 与 RPC 服务器连接的客户端
	c := protobuf.NewHelloClient(conn)

	// Contact the server and print out its response.
	name := defaultName
	if len(os.Args) > 1 {
		name = os.Args[1]
	}

	// TODO: context 这部分看不懂
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// 客户端调用 RPC 方法, 接收服务端返回值
	r, err := c.SayHello(ctx, &protobuf.HelloRequest{Name: name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}

	log.Printf("Greeting: %s", r.Message)
}
