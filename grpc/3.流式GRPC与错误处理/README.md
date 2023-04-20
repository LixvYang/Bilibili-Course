# GRPC教程 3- 流式GRPC与错误处理

- 解决vscode中，一个目录下多个mod文件的问题，怎么解决呢？
如何 在一个目录下正常的有多个.mod 文件

使用 go work init 命令

本篇文章开始，我们将要开始学习流式GRPC与GRPC的错误处理

## 流式GRPC

什么是流式GRPC呢？

和之前我们写的普通的RPC服务写入直接返回不同，流式GRPC允许我们在一个RPC请求中建立一个Stream(流)，客户端和服务器端都可以向这个流中写入数据，当客户端写入数据时，服务器端只需要不断监听这个流就可以不断获取客户端发送的消息，直到关闭。

首先我们先说说HTTP/2,GRPC的底层就是HTTP/2协议，HTTP2支持服务器端主动向客户端去发送流数据。

举例：

```
// proto文件，在Greetering服务中加一行
  rpc StreamHello (HelloReq) returns (stream HelloResp);
```

然后proto文件就变成了
```
// 
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
  rpc StreamHello (HelloReq) returns (stream HelloResp);
}
```
接下来用我们上一篇文章的方式，借助插件，自动生成go文件。
```
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative
```

然后我们在server/server.go中实现StreamHello函数
```

func (s *server) StreamHello(in *proto.HelloReq, stream proto.Greetering_StreamHelloServer) error {
	for i := 10; i > 0; i-- {
		data := &proto.HelloResp{
			Msg: fmt.Sprintf("This is %d Msg", i),
		}
		if err := stream.Send(data); err != nil {
			return err
		}
	}
	return nil
}
```
在client/client.go中加入
```
stream, err := client.StreamHello(context.Background(), &proto.HelloReq{Name: "Lixin"})
	if err != nil {
		log.Fatal(err)
	}
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(res.GetMsg())
	}
```
运行server.go代码，client.go代码就可以得到以下结果，我们的返回结果就是这样
```
os getpid:  330496
This is 10 Msg
This is 9 Msg
This is 8 Msg
This is 7 Msg
This is 6 Msg
This is 5 Msg
This is 4 Msg
This is 3 Msg
This is 2 Msg
This is 1 Msg
```
接下来我们来改一改代码，尝试去使用客户端的流式GRPC
```
//proto文件
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
  rpc StreamHello (stream HelloReq) returns (HelloResp);
}
```
go的代码
```go
// server.go
func (s *server) StreamHello(stream proto.Greetering_StreamHelloServer) error {
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			stream.SendAndClose(&proto.HelloResp{
				Msg: "server end.",
			})
			fmt.Println("Message end.")
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(res.GetName())
	}
	return nil
}
```
client.go
```go
stream, err := client.StreamHello(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < 10; i++ {
		data := &proto.HelloReq{
			Name: fmt.Sprintf("This is %d msg from client.", i),
		}
		err = stream.Send(data)
		if err != nil {
			log.Fatal(err)
		}
	}
	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("c failed: %v", err)
	}
	log.Printf("got reply: %v", res.GetMsg())
```


双向流式GRPC
```
修改proto文件加入
rpc StreamHello (stream HelloReq) returns (stream HelloResp);
```

```go
// server.go

func (s *server) StreamHello(stream proto.Greetering_StreamHelloServer) error {
	signalch := make(chan os.Signal, 1)
	signal.Notify(signalch, os.Interrupt, syscall.SIGTERM)
	msg := ""
	go func() {
		for {
			fmt.Scanln(&msg)
			stream.Send(&proto.HelloResp{
				Msg: fmt.Sprint(msg),
			})
			msg = ""
		}
	}()

	go func() {
		for {
			// 接收流式请求
			in, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				return
			}
			fmt.Println("client: ", in.GetName())
		}
	}()
	signalType := <-signalch
	signal.Stop(signalch)
	fmt.Printf("Os Signal: <%s>", signalType)
	fmt.Println("Exit....")
	return nil
}
```

```
// client.go

	msg := ""
	go func() {
		for {
			fmt.Scanln(&msg)
			stream.Send(&proto.HelloReq{
				Name: fmt.Sprint(msg),
			})
			msg = ""
		}
	}()
	go func() {
		for {
			// 接收流式请求
			in, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				return
			}
			fmt.Println("server: ", in.GetMsg())
		}
	}()

	signalType := <-signalch
	signal.Stop(signalch)
	//cleanup before exit
	log.Printf("On Signal <%s>", signalType)
	log.Println("Exit command received. Exiting...")
```

GRPC对比WebSocket

WebSocket是HTML5新增的协议，它的目的是在浏览器和服务器之间建立一个不受限的双向通信的通道，比如说，服务器可以在任意时刻发送消息给浏览器。

为什么传统的HTTP协议不能做到WebSocket实现的功能？这是因为HTTP协议是一个请求－响应协议，请求必须先由浏览器发给服务器，服务器才能响应这个请求，再把数据发送给浏览器。换句话说，浏览器不主动请求，服务器是没法主动发数据给浏览器的。

这样一来，要在浏览器中搞一个实时聊天，在线炒股（不鼓励），或者在线多人游戏的话就没法实现了，只能借助Flash这些插件。

也有人说，HTTP协议其实也能实现啊，比如用轮询或者Comet。轮询是指浏览器通过JavaScript启动一个定时器，然后以固定的间隔给服务器发请求，询问服务器有没有新消息。这个机制的缺点一是实时性不够，二是频繁的请求会给服务器带来极大的压力。

Comet本质上也是轮询，但是在没有消息的情况下，服务器先拖一段时间，等到有消息了再回复。这个机制暂时地解决了实时性问题，但是它带来了新的问题：以多线程模式运行的服务器会让大部分线程大部分时间都处于挂起状态，极大地浪费服务器资源。另外，一个HTTP连接在长时间没有数据传输的情况下，链路上的任何一个网关都可能关闭这个连接，而网关是我们不可控的，这就要求Comet连接必须定期发一些ping数据表示连接“正常工作”。

以上两种机制都治标不治本，所以，HTML5推出了WebSocket标准，让浏览器和服务器之间可以建立无限制的全双工通信，任何一方都可以主动发消息给对方。

```
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// We'll need to define an Upgrader
// this will require a Read and Write buffer size
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	// upgrade this connection to a WebSocket
	// connection
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	// helpful log statement to show connections
	log.Println("Client Connected")

	reader(ws)
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Home Page")
}

// define a reader which will listen for
// new messages being sent to our WebSocket
// endpoint
func reader(conn *websocket.Conn) {
	for {
		// read in a message
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		// print out that message for clarity
		reply := fmt.Sprintf("server reply: %s", p)
		fmt.Println(reply)

		if err := conn.WriteMessage(messageType, []byte(reply)); err != nil {
			log.Println(err)
			return
		}

	}
}

func setupRoutes() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/ws", wsEndpoint)
}

func main() {
	fmt.Println("Hello World")
	setupRoutes()
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

流式 gRPC 和 WebSockets 都是用于实现客户端和服务器之间的双向通信，但它们有以下几个区别：

协议：gRPC 是基于 HTTP/2 协议的，而 WebSocket 是一种独立的协议。HTTP/2 是一个二进制协议，可提供更好的性能和安全性。

语言支持：gRPC 支持多种语言，包括 Java、Python、Go 等，而 WebSocket 主要支持 Web 技术栈，如 JavaScript。

应用场景：gRPC 通常用于在微服务架构中进行服务间通信，而 WebSocket 更多地用于实时通信应用程序，如在线游戏或聊天应用程序。

通信方式：在流式 gRPC 中，客户端和服务器之间的通信是通过流来完成的，客户端可以发送多个请求，服务器也可以发送多个响应。而在 WebSocket 中，客户端和服务器之间的通信是通过消息来完成的，消息可以是文本或二进制数据。

总之，gRPC 和 WebSocket 都有其各自的优势和适用场景。选择哪种技术应该根据应用程序的需求和设计来决定。


我们再来学一下GRPC的错误处理

GRPC自己定义了一些常见的错误码，和我们可以在[codes](https://pkg.go.dev/google.golang.org/grpc/codes)找到。
需要使用时，需要引入codes包

```
https://pkg.go.dev/google.golang.org/grpc/codes
```

使用codes时，需要配合status使用
```
import "google.golang.org/grpc/status"
```

GRPC的方法，一般是返回err或者status类型的错误，然后调用GRPC的一方若`err!=nil`我们可以通过status.Convert方法读取对应的错误。
```
if err != nil {
s := status.Convert(err)        // 将err转为status
	for _, d := range s.Details() { // 获取details
	switch info := d.(type) {
	case *errdetails.QuotaFailure:
		fmt.Printf("Quota failure: %s\n", info)
	default:
		fmt.Printf("Unexpected type: %s\n", info)
	}
}
```

```
//  代码中的例子
// server.go
func (s *server) StreamHello(stream proto.Greetering_StreamHelloServer) error {
	st := status.New(codes.Aborted, "error!!!!!!")
	ds, err := st.WithDetails(
		&errdetails.BadRequest{
			FieldViolations: []*errdetails.BadRequest_FieldViolation{
				{
					Description: "Bad Request",
					Field: "bad",
				},
			},
		},
	)
	if err != nil {
		return st.Err()
	}
	return ds.Err()
}
```

```
// client.go

	for {
		// 接收流式请求
		in, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("braeak")
			return
		}
		if err != nil {
			s := status.Convert(err)        // 将err转为status
			for _, d := range s.Details() { // 获取details
				fmt.Println(d)
				switch info := d.(type) {
				case *errdetails.BadRequest:
					fmt.Printf("BadRequest failure: %s\n", info)
				default:
					fmt.Println(info)
				}
			}
			return
		}
		fmt.Println("server: ", in.GetMsg())
	}
```
## 总结 

- 在vscode下同一个目录下多个go.mod文件解决

- 服务端流式GRPC

- 客户端流式GRPC

- 双端流式GRPC

- GRPC对比WebSocket

- GRPC的错误处理





