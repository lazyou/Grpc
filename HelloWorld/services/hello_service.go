// hello 服务实现
package services

import (
	"context"
	"log"
	"test/protobuf"
	"test/util"
	"time"
)

// HelloServer 服务的实现
type HelloServer struct {
	protobuf.HelloServer // PS: struct 包含(内嵌) interface 之后，并不需要实现 interface 的接口，也能成为 interface 接口类
}

func (s *HelloServer) SayHello(ctx context.Context, request *protobuf.HelloRequest) (*protobuf.HelloResponse, error) {
	log.Printf("Received: %v \n", request.Name)

	return &protobuf.HelloResponse{
		Result: &protobuf.BaseResponse{
			IsOk:    true,
			Code:    200,
			Message: "",
		},
		Message: "Hello! " + request.Name,
	}, nil
}

func (s *HelloServer) SayHelloServerStream(request *protobuf.HelloRequest, stream protobuf.Hello_SayHelloServerStreamServer) error {
	log.Printf("Received: %v \n", request.Name)

	// 每秒
	for {
		time.Sleep(time.Second)

		err := stream.Send(&protobuf.HelloResponse{
			Result: &protobuf.BaseResponse{
				IsOk:    true,
				Code:    200,
				Message: "",
			},
			Message: "Hello! " + util.NowTime(),
		})

		if err != nil {
			return err
		}
	}

	return nil
}
