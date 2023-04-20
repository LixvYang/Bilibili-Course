// server.go
package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
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
	// 进程id
	fmt.Println(os.Getpid())

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