package service

import (
	"context"
	"fmt"

	"github.com/stephen-pp/p2p-status/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (p *PeerService) RegisterNewPeer(ctx context.Context, req *proto.NewPeerRequest) (*proto.Empty, error) {
	p.peersMutex.Lock()
	defer p.peersMutex.Unlock()

	// Create our peer
	p.peers[req.Id] = &Peer{
		ID:   req.Id,
		Host: req.Host,
		Port: int16(req.Port),
	}

	// Send peer to the heartbeat channel
	p.heartbeat.peerChannel <- p.peers[req.Id]

	// Send the new peer out to all our connected peers
	for _, peer := range p.peers {
		if peer.ID != req.Id {
			client, err := grpc.NewClient(fmt.Sprintf("%s:%d", peer.Host, peer.Port),
				grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				fmt.Printf("Error creating client for peer %s: %v\n", peer.ID, err)
				continue
			}

			peerClient := proto.NewPeerServiceClient(client)
			_, err = peerClient.RegisterNewPeer(context.Background(), &proto.NewPeerRequest{
				Id:   req.Id,
				Host: req.Host,
				Port: int32(req.Port),
			})
			if err != nil {
				fmt.Printf("Error sending new peer to peer %s: %v\n", peer.ID, err)
			}
		}
	}
	return &proto.Empty{}, nil
}
