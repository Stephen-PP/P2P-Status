package service

import (
	"sync"

	"github.com/google/uuid"
	"github.com/stephen-pp/p2p-status/proto"
)

type PeerService struct {
	proto.UnimplementedPeerServiceServer
	peers      map[string]*Peer
	peersMutex sync.RWMutex
	self       *Peer
	heartbeat  *HeartbeatService
}

type Peer struct {
	ID   string
	Host string
	Port int16
}

func NewPeerServer(host string, port int16) *PeerService {
	peerService := &PeerService{
		peers:      make(map[string]*Peer),
		peersMutex: sync.RWMutex{},
		self: &Peer{
			Host: host,
			Port: port,
			ID:   uuid.NewString(),
		},
		heartbeat: &HeartbeatService{
			peerChannel: make(chan *Peer),
			peers:       make(map[string]bool),
			peersMutex:  sync.RWMutex{},
		},
	}

	peerService.heartbeat.peerService = peerService

	return peerService
}
