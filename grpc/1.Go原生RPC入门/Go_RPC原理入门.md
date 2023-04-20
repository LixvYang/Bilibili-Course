# Go 原生RPC原理及入门

GRPC教程 1- Go语言原生RPC原理 

本篇文章介绍一下RPC的概念以及在Go语言如何使用标准库中的RPC.

RPC是全称叫Remote Procedure Call，远程过程调用，它允许像调用本地服务一样去调用远程服务，相对应的就是本地调用。

本地调用的例子

```
package main

import (
	"fmt"
	"os"
	"time"
)

type Args struct {
	A int
	B int
}

func Add(args *Args) int {
	return args.A + args.B 
}

func main() {
	fmt.Println(os.Getpid())
	fmt.Println(Add(&Args{10, 20}))
	time.Sleep(100 * time.Second)
}

```

可以看到结果是

```
107088
30
```

那么与此相对应的RPC调用即，不是同一进程下的服务调用。

```go
// server.go
package main

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
)

type Args struct {
	A int
	B int
}

type ServiceA struct{}

func (s ServiceA) Add(a *Args, reply *int) error {
	*reply = a.A + a.B
	return nil
}

type ServiceB struct{}

func (s ServiceB) Sub(a *Args, reply *int) error {
	*reply = a.A - a.B
	return nil
}

func main() {
	serviceA := new(ServiceA)
	serviceB := new(ServiceB)
	rpc.Register(serviceA)
	rpc.Register(serviceB)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", ":7890")
	if err != nil {
		log.Fatal("listen error:", err)
	}
	http.Serve(l, nil)
}
```

```go
// client.go
package main

import (
	"fmt"
	"log"
	"net/rpc"
)

type Args struct {
	A, B int
}

func main() {
	// 建立HTTP连接
	client, err := rpc.DialHTTP("tcp", "127.0.0.1:7890")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer client.Close()

	// 同步调用
	args := &Args{10, 20}
	var reply int
	err = client.Call("ServiceA.Add", args, &reply)
	if err != nil {
		log.Fatal("ServiceA.Add error:", err)
	}
	fmt.Printf("ServiceA.Add: %d+%d=%d\n", args.A, args.B, reply)

	// 异步调用
	var reply2 int
	divCall := client.Go("ServiceB.Sub", args, &reply2, nil)
	replyCall := <-divCall.Done // 接收调用结果
	fmt.Println("ServiceB.Sub error:", replyCall.Error)
	fmt.Printf("ServiceB.Sub: %d+%d=%d\n", args.A, args.B, reply2)
}
```
在这里，client调用的是server下的服务，即不是同一个进程下的服务，这样的好处有很多，比如如果一个项目过大，那么维护起来会有很大的成本，可以将服务给分离成很多服务，来降低服务之间的耦合度，这也是我们常说的微服务。




