// server.go
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/stephen-pp/p2p-status/proto"
	"google.golang.org/grpc"
)

type peerServer struct {
	proto.UnimplementedPeerServiceServer
	peers map[string]*Peer // Store peer info
}

type Peer struct {
	ID   string
	Host string
	Port string
}

func NewPeerServer() *peerServer {
	return &peerServer{
		peers: make(map[string]*Peer),
	}
}

func (s *peerServer) RegisterNewStatusCheck(ctx context.Context, req *proto.RegisterNewStatusCheckRequest) (*proto.Empty, error) {
	log.Printf("Received status check registration request for ID: %s", req.Id)

	switch check := req.Data.(type) {
	case *proto.RegisterNewStatusCheckRequest_Http:
		// Handle HTTP check
		log.Printf("HTTP check for URL: %s", check.Http.Url)
	case *proto.RegisterNewStatusCheckRequest_Ping:
		// Handle Ping check
		log.Printf("Ping check for host: %s", check.Ping.Host)
	}

	return &proto.Empty{}, nil
}

func (s *peerServer) RegisterNewPeer(ctx context.Context, req *proto.NewPeerRequest) (*proto.Empty, error) {
	log.Printf("Registering new peer with ID: %s", req.Id)
	s.peers[req.Id] = &Peer{
		ID:   req.Id,
		Host: req.Host,
		Port: req.Port,
	}
	return &proto.Empty{}, nil
}

func (s *peerServer) RequestHeartbeat(ctx context.Context, req *proto.Empty) (*proto.HeartbeatResponse, error) {
	fmt.Println("Received heartbeat response. ")
	return &proto.HeartbeatResponse{
		Success: true,
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", os.Getenv("PORT"))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	server := NewPeerServer()
	server.peers["abc"] = &Peer{
		ID:   "abc",
		Host: "localhost",
		Port: os.Getenv("PEER_PORT"),
	}
	proto.RegisterPeerServiceServer(grpcServer, server)

	// Function that runs every 5 seconds
	go func() {
		<-time.After(3 * time.Second)
		// Create client connection to the peer ID
		client, err := grpc.NewClient(server.peers["abc"].Host+":"+server.peers["abc"].Port, grpc.WithInsecure())
		if err != nil {
			panic(err)
		}

		// Create a new client
		peerClient := proto.NewPeerServiceClient(client)

		for {
			fmt.Println("Requesting heartbeat from :%s", server.peers["abc"].Port)
			res, err := peerClient.RequestHeartbeat(context.Background(), &proto.Empty{})
			if err != nil {
				log.Printf("Error requesting heartbeat: %v", err)
			} else {
				log.Printf("Heartbeat response: %v", res)
			}
			<-time.After(5 * time.Second)
		}
	}()

	log.Printf("Starting gRPC server on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
