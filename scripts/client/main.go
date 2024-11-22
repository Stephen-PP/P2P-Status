package main

import (
	"context"
	"fmt"
	"log"
	"time"

	clients "github.com/stephen-pp/p2p-status/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RelayClient struct {
	client    clients.RelayServiceClient
	conn      *grpc.ClientConn
	clientKey string
	clientId  string
}

func NewRelayClient(serverAddr string) (*RelayClient, error) {
	// Set up connection to the server
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}

	return &RelayClient{
		client: clients.NewRelayServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *RelayClient) Register(host string, port int32) error {
	// Register the client
	resp, err := c.client.RegisterClient(context.Background(), &clients.RegisterClientRequest{
		Host: host,
		Port: port,
	})
	if err != nil {
		return fmt.Errorf("failed to register: %v", err)
	}

	c.clientKey = resp.ClientKey
	c.clientId = resp.Id
	log.Printf("Registered with ID: %s and Key: %s", resp.Id, resp.ClientKey)
	return nil
}

func (c *RelayClient) StartListeningForClients(ctx context.Context) error {
	stream, err := c.client.GetClients(ctx, &clients.ClientKeyRequest{
		ClientKey: c.clientKey,
	})
	if err != nil {
		return fmt.Errorf("failed to start client stream: %v", err)
	}

	go func() {
		for {
			client, err := stream.Recv()
			if err != nil {
				log.Printf("Stream ended: %v", err)
				return
			}
			// Skip our own client
			if client.Id == c.clientId {
				continue
			}
			log.Printf("Received client update: ID=%s Host=%s Port=%d",
				client.Id, client.Host, client.Port)
		}
	}()

	return nil
}

func (c *RelayClient) Close() error {
	// Cleanup: delete our registration and close connection
	if c.clientKey != "" {
		_, err := c.client.DeleteClient(context.Background(), &clients.ClientKeyRequest{
			ClientKey: c.clientKey,
		})
		if err != nil {
			log.Printf("Failed to delete client: %v", err)
		}
	}
	return c.conn.Close()
}

func main() {
	client, err := NewRelayClient("localhost:9000")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Register with some example values
	err = client.Register("localhost", 8080)
	if err != nil {
		log.Fatalf("Failed to register: %v", err)
	}

	// Create a cancellable context for the client stream
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start listening for other clients
	err = client.StartListeningForClients(ctx)
	if err != nil {
		log.Fatalf("Failed to start listening: %v", err)
	}

	// Keep the client running for a while
	time.Sleep(time.Minute)
}
