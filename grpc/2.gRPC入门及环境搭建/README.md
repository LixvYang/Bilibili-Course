# GRPC教程 2- gRPC下载以及入门gRPC

本篇文章我们开始学习gRPC，我们之前学到了RPC(远程过程调用)，那么gRPC是什么呢？gRPC是Google开源的现代高性能RPC框架，能够运行在任何环境中，使用HTTP2作为传输协议。

gRPC与RPC一样，可以像调用本地方法一样去调用另一个进程上的服务，这可以帮助你很轻松的创建微服务程序。gRPC只是定义类型和远程服务带有的参数和返回类型，我们需要在gRPC服务端程序中定义服务的逻辑，在客户端调用和服务器端相同的方法。

## 安装gRPC

1. 安装Protocol  Buffers

```shell
https://github.com/google/protobuf/releases
```

下载

```
protoc-22.2-linux-x86_64.zip
```
2. 解压缩文件

```
unzip protoc-22.2-linux-x86_64.zip -d protoc
```

将 protoc下的 bin目录下的 `protoc`文件加入到`$GOPATH/bin`目录下， 然后将include目录 放到`$GOPATH` 目录下，以便于我们编写proto文件时，可以找到对应的文件。

3. 下载protobuf go语言的插件
进入 https://grpc.io , 点击 go 进入https://grpc.io/docs/languages/go/quickstart/

开始下载go的插件

```
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28 // 生成 .pb.go文件
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2 // 生成 _grpc.pb.go 文件
$ export PATH="$PATH:$(go env GOPATH)/bin"
```
go语言下的protobuf 环境下载好了

4. 开始编写代码

```
// protobuf编程的几个步骤
1. 编写.proto文件
2. 使用.proto文件生成对应语言的文件
3. 编写业务逻辑
```

首先随便选个目录 创建两个文件夹，client和server,这两个文件夹将成为我们调用gRPC的客户端和服务端。

我们在client和server两个文件夹中分别mod init一下，并且创建对应的调用函数文件

```
// /server
go mod int server

// /client
go mod int client

```
接下来在这两个目录下都创建一个proto文件夹。
首先然后在server/proto文件夹下创建一个hello.proto,在client/proto文件夹下创建一个hello.proto文件。

```
// server/proto/hello.proto
syntax = "proto3";

option go_package = "server/proto";

package proto;

// Hello Request
message HelloReq {
  string name = 1;
}

// Hello Response
message HelloResp {
  string msg = 1;
}

service Greetering {
  rpc Hello (HelloReq) returns (HelloResp);
}

// client/proto/hello.proto
syntax = "proto3";

option go_package = "client/proto";

package proto;

// Hello Request
message HelloReq {
  string name = 1;
}

// Hello Response
message HelloResp {
  string msg = 1;
}

service Greetering {
  rpc Hello (HelloReq) returns (HelloResp);
}
```

接着，你需要对这两个.proto 通过我们刚刚下载的的protoc工具和插件去生成相应的go代码。

```
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative server/proto/hello.proto client/proto.hello.proto
```

最后分别在两个目录下执行`go mod tidy`，然后目录的结构就如下所示

```
.
├── client
│   ├── client.go
│   ├── go.mod
│   ├── go.sum
│   └── proto
│       ├── hello_grpc.pb.go
│       ├── hello.pb.go
│       └── hello.proto
└── server
    ├── go.mod
    ├── go.sum
    ├── proto
    │   ├── hello_grpc.pb.go
    │   ├── hello.pb.go
    │   └── hello.proto
    └── server.go
```

随后，我们开始写对应的逻辑
```
// server/server.go
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"server/proto"

	"google.golang.org/grpc"
)

type server struct {
	proto.UnimplementedGreeteringServer
}

func (s *server) Hello(ctx context.Context, req *proto.HelloReq) (*proto.HelloResp, error) {
	msg := fmt.Sprintf("Hello, %s", req.GetName())
	return &proto.HelloResp{
		Msg: msg,
	}, nil
}

func main() {

	fmt.Println("os getpid: ", os.Getpid())

	l, err := net.Listen("tcp", ":7890")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	s := grpc.NewServer()
	proto.RegisterGreeteringServer(s, &server{})
	err = s.Serve(l)
	if err != nil {
		log.Fatal(err)
		return
	}
}
```

```
// client/client.go
package main

import (
	"client/proto"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {

	fmt.Println("os getpid: ", os.Getpid())

	conn, err := grpc.Dial("127.0.0.1:7890", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := proto.NewGreeteringClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := client.Hello(ctx, &proto.HelloReq{Name: "Cheng Long"})
	if err != nil {
		log.Fatal(err)
	}
	log.Println(resp.GetMsg())
}
```

我们在`server`文件夹下执行`go run server.go`在`client`文件夹下执行`go run client.go`就可以输出

```
Hello,Cheng Long
```
我们就成功执行了gRPC代码。

## 总结
下载安装压缩gRPC的环境
开始编写gRPC的相关代码




