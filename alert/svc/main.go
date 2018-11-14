package main

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
)

func main() {

	l, _ := net.Listen("tcp", fmt.Sprintf("localhost:%d", 9000))
	server := grpc.NewServer()
	server.Serve(l)
}
