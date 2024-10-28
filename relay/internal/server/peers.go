package server

import (
	"encoding/json"
	"net/http"
)

// Peer addition route (takes a host and a port)
func (s *RelayServer) AddPeer(w http.ResponseWriter, r *http.Request) {
	// Import host and port from JSON body
	type Peer struct {
		Host string `json:"host"`
		Port string `json:"port"`
	}

	// Read JSON body
	decoder := json.NewDecoder(r.Body)
	var p Peer
	err := decoder.Decode(&p)
	if err != nil {
		w.Write([]byte("Could not decode JSON body"))
	}

	protocol := "http"
	err = attemptPeerConnection(protocol, p.Host, p.Port)
	if err != nil {
		protocol = "https"
		err = attemptPeerConnection(protocol, p.Host, p.Port)
		if err != nil {
			w.Write([]byte("Could not connect to peer"))
		}
	}

	_, err = s.Db.Exec("INSERT INTO peers (host, port, protocol) VALUES (?, ?)", p.Host, p.Port, protocol)
	if err != nil {
		w.Write([]byte("Could not insert peer into database"))
	}
}

func attemptPeerConnection(protocol string, host string, port string) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", protocol+"://"+host+":"+port+"/health", nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return err
	}
	if resp.Header.Get("X-Peer") != "true" {
		return err
	}

	return nil
}
