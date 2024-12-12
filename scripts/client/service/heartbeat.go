package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/stephen-pp/p2p-status/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (p *PeerService) RequestHeartbeat(ctx context.Context, req *proto.Empty) (*proto.HeartbeatResponse, error) {
	return &proto.HeartbeatResponse{
		Success: true,
	}, nil
}

type HeartbeatService struct {
	peerChannel chan *Peer
	peers       map[string]bool
	peersMutex  sync.RWMutex
	peerService *PeerService
}

func (h *HeartbeatService) AddNewPeer(peer *Peer) {
	h.peerChannel <- peer
}

// The heartbeat service will send heartbeat requests every minute to all connected peers.
func (h *HeartbeatService) StartHeartbeatService() {
	// Goroutines to keep us updated with the peer service
	go func() {
		for {
			select {
			case peer := <-h.peerChannel:
				h.peersMutex.Lock()
				h.peers[peer.ID] = true
				h.peersMutex.Unlock()
				go h.heartbeatHandler(peer.ID)
			}
		}
	}()

	// Initial peer sync
	h.peerService.peersMutex.RLock()
	for _, peer := range h.peerService.peers {
		h.peerChannel <- peer
	}
	h.peerService.peersMutex.RUnlock()
}

func (h *HeartbeatService) heartbeatHandler(peerID string) {
	for {
		// Read peer from service
		h.peerService.peersMutex.RLock()
		peer := h.peerService.peers[peerID]
		h.peerService.peersMutex.RUnlock()

		// If peer no longer exists, stop heartbeating
		if peer == nil {
			h.peersMutex.Lock()
			delete(h.peers, peerID)
			h.peersMutex.Unlock()
			return
		}

		// Send heartbeat
		client, err := grpc.NewClient(fmt.Sprintf("%s:%d", peer.Host, peer.Port),
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Printf("Error creating client for peer %s: %v\n", peerID, err)
			continue
		}

		peerClient := proto.NewPeerServiceClient(client)
		_, err = peerClient.RequestHeartbeat(context.Background(), &proto.Empty{})
		if err != nil {
			fmt.Printf("Error sending heartbeat to peer %s: %v\n", peerID, err)
		}

		time.Sleep(time.Minute) // Wait before next heartbeat
	}
}
