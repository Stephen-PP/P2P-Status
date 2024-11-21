package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"

	clients "github.com/stephen-pp/p2p-status/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type relayServer struct {
	clients.UnimplementedRelayServiceServer
	createdClients map[string]*clients.Client
}

func main() {
	// Create a TCP listener
	lis, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Create a gRPC server
	s := grpc.NewServer()
	clients.RegisterRelayServiceServer(s, &relayServer{})

	// Serve the gRPC server
	fmt.Printf("Server listening on port %v\n", lis.Addr().String())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// RegisterClient is a gRPC method that registers a client
func (s *relayServer) RegisterClient(ctx context.Context, req *clients.RegisterClientRequest) (*clients.RegisterClientResponse, error) {
	// Generate a random integer for the ID and client key
	id := rand.Intn(1000000000)
	clientKey := rand.Intn(1000000000)

	// Create a new client
	client := &clients.Client{
		Host: req.Host,
		Port: req.Port,
		Id:   fmt.Sprintf("%d", id),
	}

	// Append the client to the list of created clients
	s.createdClients[fmt.Sprintf("%d", clientKey)] = client

	// Return the client
	return &clients.RegisterClientResponse{
		Host:      client.Host,
		Port:      client.Port,
		Id:        client.Id,
		ClientKey: fmt.Sprintf("%d", clientKey),
	}, nil
}

// DeleteClient is a gRPC method that deletes a client
func (s *relayServer) DeleteClient(ctx context.Context, req *clients.DeleteClientRequest) (*clients.Client, error) {
	// Get the client key
	clientKey := req.ClientKey

	// Get the client from the list of created clients
	client := s.createdClients[clientKey]

	// Delete the client from the list of created clients
	delete(s.createdClients, clientKey)

	// Return the client
	return client, nil
}

// GetClients is a gRPC method that streams clients
func (s *relayServer) GetClients(req *emptypb.Empty, stream clients.RelayService_GetClientsServer) error {
	// Stream all clients
	for _, client := range s.createdClients {
		if err := stream.Send(client); err != nil {
			return err
		}
	}

	return nil
}
