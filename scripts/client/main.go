// server.go
package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/stephen-pp/p2p-status/proto"
	"github.com/stephen-pp/p2p-status/scripts/client/service"
	"google.golang.org/grpc"
)

func main() {
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Fatalf("failed to convert port to int: %v", err)
	}

	var host string = os.Getenv("HOST")
	if strings.TrimSpace(host) == "" {
		host = strings.TrimSpace(host)
	}

	peerService := service.NewPeerServer(host, int16(port))

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	proto.RegisterPeerServiceServer(grpcServer, peerService)

	log.Printf("Starting gRPC server on %s:%d", host, port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
