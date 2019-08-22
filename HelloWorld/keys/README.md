## 证书制作
* 制作私钥: 
    * 方式1: `openssl genrsa -out server.key 2048`
    * 方式2: `openssl ecparam -genkey -name secp384r1 -out server.key`

* 自签名公钥:    
    * `openssl req -new -x509 -sha256 -key server.key -out server.pem -days 3650`

    * 注意: 在 Common Name 输入的值要记住, 客户端(client.go)需要用.
