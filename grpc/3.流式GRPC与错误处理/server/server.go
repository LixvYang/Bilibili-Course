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
