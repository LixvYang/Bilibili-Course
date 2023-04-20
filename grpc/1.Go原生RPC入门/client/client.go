// client.go
package main

import (
	"fmt"
	"log"
	"net/rpc"
	"os"
)

type Args struct {
	A, B int
}

func main() {

	fmt.Println(os.Getpid())

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
