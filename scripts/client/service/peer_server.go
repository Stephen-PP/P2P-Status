package service

import (
	"sync"

	"github.com/google/uuid"
)

type PeerService struct {
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
