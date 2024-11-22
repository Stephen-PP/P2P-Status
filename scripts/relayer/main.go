package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/google/uuid"
	clients "github.com/stephen-pp/p2p-status/proto"
	"google.golang.org/grpc"
)

type relayServer struct {
	clients.UnimplementedRelayServiceServer
	createdClients      map[string]*clients.Client
	createdClientsMutex sync.RWMutex
	clientStreams       map[string]clients.RelayService_GetClientsServer
	clientStreamsMutex  sync.RWMutex
}

func main() {
	// Create a TCP listener
	lis, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Create a gRPC server
	s := grpc.NewServer()
	clients.RegisterRelayServiceServer(s, &relayServer{
		createdClients: make(map[string]*clients.Client),
		clientStreams:  make(map[string]clients.RelayService_GetClientsServer),
	})

	// Serve the gRPC server
	fmt.Printf("Server listening on port %v\n", lis.Addr().String())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// RegisterClient is a gRPC method that registers a client
func (s *relayServer) RegisterClient(ctx context.Context, req *clients.RegisterClientRequest) (*clients.RegisterClientResponse, error) {
	fmt.Println("register client called")
	// Generate a random integer for the ID and client key
	id := uuid.New().String()
	clientKey := uuid.New().String()

	// Create a new client
	client := &clients.Client{
		Host: req.Host,
		Port: req.Port,
		Id:   id,
	}

	// Append the client to the list of created clients
	s.createdClientsMutex.Lock()
	s.createdClients[clientKey] = client
	s.createdClientsMutex.Unlock()

	// Now, let's stream this client to all clients
	s.clientStreamsMutex.RLock()
	for _, stream := range s.clientStreams {
		// We don't actually care about results here, so we can just fire and forget
		go func(stream clients.RelayService_GetClientsServer, streamId string) {
			if err := stream.Send(client); err != nil {
				log.Printf("Error streaming client to client: %v", err)
				return
			}
		}(stream, id)
	}
	s.clientStreamsMutex.RUnlock()

	// Return the client
	return &clients.RegisterClientResponse{
		Host:      client.Host,
		Port:      client.Port,
		Id:        client.Id,
		ClientKey: clientKey,
	}, nil
}

// DeleteClient is a gRPC method that deletes a client
func (s *relayServer) DeleteClient(ctx context.Context, req *clients.ClientKeyRequest) (*clients.Client, error) {
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
func (s *relayServer) GetClients(req *clients.ClientKeyRequest, stream clients.RelayService_GetClientsServer) error {
	// Stream all clients
	s.clientStreamsMutex.RLock()
	for _, client := range s.createdClients {
		if err := stream.Send(client); err != nil {
			s.clientStreamsMutex.RUnlock()
			return err
		}
	}
	s.clientStreamsMutex.RUnlock()

	// Generate a unique ID for this stream
	id := uuid.New().String()

	// Now, let's add this stream to the list of client streams
	s.clientStreamsMutex.Lock()
	s.clientStreams[id] = stream
	s.clientStreamsMutex.Unlock()

	// Now let's queue up the deletion of this stream when it closes
	defer func() {
		fmt.Println("Deleting client stream due to closed")
		s.clientStreamsMutex.Lock()
		delete(s.clientStreams, id)
		s.clientStreamsMutex.Unlock()

		s.createdClientsMutex.Lock()
		delete(s.createdClients, req.ClientKey)
		s.createdClientsMutex.Unlock()
	}()

	<-stream.Context().Done()
	return nil
}
