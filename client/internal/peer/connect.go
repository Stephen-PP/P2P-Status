package peer

import (
	"fmt"
	"net"
	"os"
)

type PeerService struct {
	UDPConn *net.UDPConn
}

func StartUDPServer() {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%s", os.Getenv("PEER_PORT")))
	if err != nil {
		panic(err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	exitChan := make(chan error, 1)
	buffer := make([]byte, 1024)
	go func() {
		for {
			n, remoteAddr, err := conn.ReadFromUDP(buffer)
			if err != nil {
				fmt.Printf("Error reading from UDP: %v\n", err)
				return
			}

			message := string(buffer[:n])
			fmt.Printf("Received message from %v: %s\n", remoteAddr, message)

			// Echo the message back to the sender
			_, err = conn.WriteToUDP(buffer[:n], remoteAddr)
			if err != nil {
				fmt.Printf("Error sending response: %v\n", err)
			}
		}
	}()

	<-exitChan
}
