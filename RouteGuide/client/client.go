// 调用服务方法:
// 注意，在 gRPC-Go 中，RPC以阻塞/同步模式操作，这意味着 RPC 调用等待服务器响应，同时要么返回响应，要么返回错误。
package main

import (
	"context"
	"flag"
	"io"
	"log"
	"math/rand"
	"time"

	pb "RouteGuide/routeguide"
	"RouteGuide/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/testdata"
)

var (
	tls                = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	caFile             = flag.String("ca_file", "", "The file containning the CA root cert file")
	serverAddr         = flag.String("server_addr", "127.0.0.1:40000", "The server address in the format of host:port")
	serverHostOverride = flag.String("server_host_override", "x.test.youtube.com", "The server name use to verify the hostname returned by TLS handshake")
	runFunName         = flag.String("funcName", "printFeature", "要执行的方法, 方便代码阅读查看效果")
)

// printFeature gets the feature for the given point.
// 1. 简单 RPC
func printFeature(client pb.RouteGuideClient, point *pb.Point) {
	log.Printf("【1. 简单 RPC -- 请求】 Getting feature for point (%d, %d)", point.Latitude, point.Longitude)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 调用简单 RPC GetFeature 几乎是和调用一个本地方法一样直观
	feature, err := client.GetFeature(ctx, point)
	if err != nil {
		log.Fatalf("%v.GetFeatures(_) = _, %v: ", client, err)
	}

	log.Printf("【1. 简单 RPC -- 返回值】 %v", feature)
}

// printFeatures lists all the features within the given bounding Rectangle.
// 服务器端流式 RPC
func printFeatures(client pb.RouteGuideClient, rect *pb.Rectangle) {
	log.Printf("Looking for features within %v", rect)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 服务器端流方法，它会返回地理的Feature 流
	stream, err := client.ListFeatures(ctx, rect)
	if err != nil {
		log.Fatalf("%v.ListFeatures(_) = _, %v", client, err)
	}

	// Recv() 方法去反复读取服务器的响应到一个响应 protocol buffer 对象（在这个场景下是Feature）直到消息读取完毕：
	// 每次调用完成时，客户端都要检查从 Recv() 返回的错误 err。
	// 如果返回为 nil，流依然完好并且可以继续读取；
	// 如果返回为 io.EOF，则说明消息流已经结束；否则就一定是一个通过 err 传过来的 RPC 错误。
	for {
		feature, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("%v.ListFeatures(_) = _, %v", client, err)
		}

		log.Println(feature)
	}
}

// runRecordRoute sends a sequence of points to server and expects to get a RouteSummary from server.
// 需要给方法传入一个上下文而后返回 RouteGuide_RecordRouteClient 流以外，
// 客户端流方法 RecordRoute 和服务器端方法类似，它可以用来读 和 写消息
func runRecordRoute(client pb.RouteGuideClient) {
	// Create a random number of random points
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	pointCount := int(r.Int31n(100)) + 2 // Traverse at least two points
	var points []*pb.Point

	for i := 0; i < pointCount; i++ {
		points = append(points, util.RandomPoint(r))
	}

	log.Printf("Traversing %d points.", len(points))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := client.RecordRoute(ctx)
	if err != nil {
		log.Fatalf("%v.RecordRoute(_) = _, %v", client, err)
	}

	for _, point := range points {
		// Send() 方法用来给服务器发送请求
		if err := stream.Send(point); err != nil {
			log.Fatalf("%v.Send(%v) = %v", stream, point, err)
		}
	}

	// CloseAndRecv()方法 让 gRPC 知道我们已经完成了写入同时期待返回应答
	// CloseAndRecv() 返回的 err 中获得 RPC 的状态。
	// 如果状态为nil，那么CloseAndRecv()的第一个返回值将会是合法的服务器应答
	reply, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
	}

	log.Printf("Route summary: %v", reply)
}

// runRouteChat receives a sequence of route notes, while sending notes for various locations.
// 和 RecordRoute 的场景类似，只给函数传 入一个上下文对象，拿到可以用来读写的流。
// 但是，当服务器依然在往 他们 的消息流写入消息时，通过方法流返回值
func runRouteChat(client pb.RouteGuideClient) {
	notes := []*pb.RouteNote{
		{Location: &pb.Point{Latitude: 0, Longitude: 1}, Message: "First message"},
		{Location: &pb.Point{Latitude: 0, Longitude: 2}, Message: "Second message"},
		{Location: &pb.Point{Latitude: 0, Longitude: 3}, Message: "Third message"},
		{Location: &pb.Point{Latitude: 0, Longitude: 1}, Message: "Fourth message"},
		{Location: &pb.Point{Latitude: 0, Longitude: 2}, Message: "Fifth message"},
		{Location: &pb.Point{Latitude: 0, Longitude: 3}, Message: "Sixth message"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := client.RouteChat(ctx)
	if err != nil {
		log.Fatalf("%v.RouteChat(_) = _, %v", client, err)
	}

	waitc := make(chan struct{})

	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				// read done.
				close(waitc)
				return
			}

			if err != nil {
				log.Fatalf("Failed to receive a note : %v", err)
			}

			log.Printf("Got message %s at point(%d, %d)", in.Message, in.Location.Latitude, in.Location.Longitude)
		}
	}()

	for _, note := range notes {
		if err := stream.Send(note); err != nil {
			log.Fatalf("Failed to send a note: %v", err)
		}
	}

	stream.CloseSend()

	<-waitc
}

func main() {
	flag.Parse()

	// 可以使用 DialOptions 在 grpc.Dial 中设置授权认证（如， TLS，GCE认证，JWT认证）
	var opts []grpc.DialOption
	if *tls {
		if *caFile == "" {
			*caFile = testdata.Path("ca.pem")
		}

		creds, err := credentials.NewClientTLSFromFile(*caFile, *serverHostOverride)
		if err != nil {
			log.Fatalf("Failed to create TLS credentials %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	// 为了调用服务方法，我们首先创建一个 gRPC channel 和服务器交互.
	// 通过给 grpc.Dial() 传入服务器地址和端口号做到这点
	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()

	// 一旦 gRPC channel 建立起来，需要一个客户端 存根 去执行 RPC。
	// 通过 .proto 生成的 pb 包提供的 NewRouteGuideClient 方法来完成
	client := pb.NewRouteGuideClient(conn)

	// 运行示例: go run client.go -funcName printFeature

	if *runFunName == "printFeature" {
		// Looking for a valid feature
		var pointExist = &pb.Point{
			Latitude:  409146138,
			Longitude: -746188906,
		}
		printFeature(client, pointExist)

		// Feature missing.
		var pointNoExist = &pb.Point{
			Latitude:  0,
			Longitude: 0,
		}
		printFeature(client, pointNoExist)
	}

	if *runFunName == "printFeatures" {
		// Looking for features between 40, -75 and 42, -73.
		printFeatures(client, &pb.Rectangle{
			Lo: &pb.Point{Latitude: 400000000, Longitude: -750000000},
			Hi: &pb.Point{Latitude: 420000000, Longitude: -730000000},
		})
	}

	if *runFunName == "runRecordRoute" {
		// RecordRoute
		runRecordRoute(client)
	}

	if *runFunName == "runRouteChat" {
		// RouteChat
		runRouteChat(client)
	}
}
