package main

import (
	"fmt"
	"net"

	"github.com/ethereum/go-ethereum/rpc"
)

func main() {
	server := rpc.NewServer()
	defer server.Stop()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Println("can't listen:", err)
	}
	defer listener.Close()
	server.ServeListener(listener)
}
