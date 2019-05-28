1. 定义服务:
    * 定义一个服务：一个 RPC 服务通过参数和返回类型来指定可以远程调用的方法;
    
    * gRPC 通过 protocol buffers 来实现;
        * 使用 protocol buffers 接口定义语言来定义服务方法;
        * 使用 protocol buffer 来定义参数和返回类型;
        * 客户端和服务端均使用服务定义生成的接口代码。
        
    * 案例: helloworld\helloworld.proto        


2. 生成 gRPC 代码:
    * 定义好服务，可以使用 protocol buffer 编译器 protoc 来生成创建应用所需的特定客户端和服务端的代码 
    
    * `protoc -I helloworld helloworld/helloworld.proto --go_out=plugins=grpc:helloworld` 生成客户端和服务端接口 helloworld\helloworld.pb.go
        * 注意需要先 `go get -u github.com/golang/protobuf/protoc-gen-go`


3. 写一个服务器
    * 案例: greeter_server/main.go
    * 3.1 服务实现 (实现 GreeterServer 服务)
    * 3.2 服务端实现 (提供一个 gRPC 服务的另一个主要功能是让这个服务实在在网络上可用)
    

4. 写一个客户端
    * 案例: greeter_client/main.go
    * 4.1 连接服务
    * 4.2 调用 RPC


5. 运行: 必须进入 Test
    * 服务端: go run greeter_server/main.go
    * 客户端: go run greeter_client/main.go
