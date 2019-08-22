## 注意
* protobuf 服务 和 数据定义

* service 和 message 分开写

* protobuf 文件编译结果也存放此处
 

## .proto文件编译结果 (进入此目录)
```sh
cd protobuf
protoc common_messages.proto --go_out=plugins=grpc:.
protoc hello_messages.proto --go_out=plugins=grpc:.
protoc hello_services.proto --go_out=plugins=grpc:.
```
