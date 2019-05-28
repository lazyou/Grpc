//go:generate protoc -I ../routeguide --go_out=plugins=grpc:../routeguide ../routeguide/RouteGuide.proto

// Package main implements a simple gRPC server that demonstrates how to use gRPC-Go libraries
// to perform unary, client streaming, server streaming and full duplex RPCs.
// gRPC 允许定义 四类服务方法 演示: 1.单项 RPC; 2.服务端流式 RPC; 3.客户端流式 RPC; 4.双向流式 RPC

// It implements the route guide service whose definition can be found in routeguide/RouteGuide.proto.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/testdata"

	pb "RouteGuide/routeguide"
	"RouteGuide/util"
)

var (
	tls        = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile   = flag.String("cert_file", "", "The TLS cert file")
	keyFile    = flag.String("key_file", "", "The TLS key file")
	jsonDBFile = flag.String("json_db_file", "", "A json file containing a list of features")
	port       = flag.Int("port", 40000, "The server port")
)

type routeGuideServer struct {
	savedFeatures []*pb.Feature // read-only after initialized

	mu         sync.Mutex // protects routeNotes
	routeNotes map[string][]*pb.RouteNote
}

// GetFeature returns the feature at the given point.
// 1. 简单 RPC: 从客户端拿到一个 Point 对象，然后从返回包含从数据库拿到的feature信息的 Feature
// 该方法传入了 RPC 的上下文对象，以及客户端的 Point protocol buffer请求。
// 它返回了一个包含响应信息和error 的 Feature protocol buffer对象。
func (s *routeGuideServer) GetFeature(ctx context.Context, point *pb.Point) (*pb.Feature, error) {
	log.Printf("【1. 简单 RPC -- 接收】 %v", point)

	for _, feature := range s.savedFeatures {
		if proto.Equal(feature.Location, point) {
			return feature, nil
		}
	}

	// No feature was found, return an unnamed feature
	return &pb.Feature{Location: point}, nil
}

// ListFeatures lists all features contained within the given bounding Rectangle.
// 2. 服务器端流式 RPC
// 请求对象是一个 Rectangle，客户端期望从中找到 Feature
// 这次我们得到了一个请求对象和一个特殊的 RouteGuide_ListFeaturesServer 来写入我们的响应，__而不是__ 得到方法参数中的简单请求和响应对象。
func (s *routeGuideServer) ListFeatures(rect *pb.Rectangle, stream pb.RouteGuide_ListFeaturesServer) error {
	log.Printf("【2. 服务器端流式 RPC -- 接收】 %v", rect)

	for _, feature := range s.savedFeatures {
		log.Printf("【2. 服务器端流式 RPC -- 发送中】 %v", feature)
		//time.Sleep(1 * time.Second) // 可模拟失败的情况

		if util.InRange(feature.Location, rect) {
			// 我们需要将多个 Feature 发回给客户端
			if err := stream.Send(feature); err != nil {
				// 如果在调用过程中发生任何错误，我们会返回一个非 nil 的错误；
				// gRPC 层会将其转化为合适的 RPC 状态通过线路发送
				return err
			}
		}
	}

	log.Printf("【2. 服务器端流式 RPC -- 流发送结束】")

	// 返回了一个 nil 错误告诉 gRPC 响应的写入已经完成
	return nil
}

// RecordRoute records a route composited of a sequence of points.
//
// It gets a stream of points, and responds with statistics about the "trip":
// number of points,  number of known features visited, total distance traveled, and
// total time spent.
// 从客户端拿到一个 Point 的流，其中包括它们路径的信息.
// 如你所见，这次这个方法没有请求参数。相反的，它拿到了一个 RouteGuide_RecordRouteServer 流:
// 服务器可以用它来同时读 和 写消息 —— 它可以用自己的 Recv() 方法接收客户端消息并且用 SendAndClose() 方法返回它的单个响应。

// 我们使用 RouteGuide_RecordRouteServer 的 Recv() 方法去反复读取客户端的请求到一个请求对象（在这个场景下是 Point），直到没有更多的消息：服务器需要在每次调用后检查 Read() 返回的错误。
// 如果返回值为 nil，流依然完好，可以继续读取；
// 如果返回值为 io.EOF，消息流结束，服务器可以返回它的 RouteSummary。
// 如果它还有其它值，我们原样返回错误，gRPC 层会把它转换为 RPC 状态。
func (s *routeGuideServer) RecordRoute(stream pb.RouteGuide_RecordRouteServer) error {
	var pointCount, featureCount, distance int32
	var lastPoint *pb.Point
	startTime := time.Now()

	for {
		// Recv() 方法接收客户端消息
		point, err := stream.Recv()
		if err == io.EOF {
			endTime := time.Now()
			// SendAndClose() 方法返回它的单个响应
			return stream.SendAndClose(&pb.RouteSummary{
				PointCount:   pointCount,
				FeatureCount: featureCount,
				Distance:     distance,
				ElapsedTime:  int32(endTime.Sub(startTime).Seconds()),
			})
		}

		if err != nil {
			return err
		}

		pointCount++
		for _, feature := range s.savedFeatures {
			if proto.Equal(feature.Location, point) {
				featureCount++
			}
		}

		if lastPoint != nil {
			distance += util.CalcDistance(lastPoint, point)
		}

		lastPoint = point
	}
}

// RouteChat receives a stream of message/location pairs, and responds with a stream of all
// previous messages at each of those locations.
func (s *routeGuideServer) RouteChat(stream pb.RouteGuide_RouteChatServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		key := util.Serialize(in.Location)

		s.mu.Lock()
		s.routeNotes[key] = append(s.routeNotes[key], in)
		// Note: this copy prevents blocking other clients while serving this one.
		// We don't need to do a deep copy, because elements in the slice are
		// insert-only and never modified.
		rn := make([]*pb.RouteNote, len(s.routeNotes[key]))
		copy(rn, s.routeNotes[key])
		s.mu.Unlock()

		for _, note := range rn {
			if err := stream.Send(note); err != nil {
				return err
			}
		}
	}
}

// loadFeatures loads features from a JSON file.
func (s *routeGuideServer) loadFeatures(filePath string) {
	var data []byte
	var err error

	if filePath != "" {
		data, err = ioutil.ReadFile(filePath)
		if err != nil {
			log.Fatalf("Failed to load default features: %v", err)
		}
	} else {
		data = util.GetExampleData()
	}

	if err := json.Unmarshal(data, &s.savedFeatures); err != nil {
		log.Fatalf("Failed to load default features: %v", err)
	}
}

func newServer() *routeGuideServer {
	s := &routeGuideServer{routeNotes: make(map[string][]*pb.RouteNote)}
	s.loadFeatures(*jsonDBFile)
	return s
}

// 构建和启动服务器
func main() {
	flag.Parse()

	// 客户端请求的监听端口
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	var opts []grpc.ServerOption
	if *tls {
		if *certFile == "" {
			*certFile = testdata.Path("server1.pem")
		}
		if *keyFile == "" {
			*keyFile = testdata.Path("server1.key")
		}
		creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
		if err != nil {
			log.Fatalf("Failed to generate credentials %v", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}

	// 创建 gRPC 服务器的一个实例
	grpcServer := grpc.NewServer(opts...)

	// 在 gRPC 服务器注册我们的服务实现
	pb.RegisterRouteGuideServer(grpcServer, newServer())

	// 用服务器 Serve() 方法以及我们的端口信息区实现阻塞等待，直到进程被杀死或者 Stop() 被调用
	grpcServer.Serve(lis)
}
