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