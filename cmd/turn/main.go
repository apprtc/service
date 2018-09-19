package main

import (
	"log"
	"os"
	"strconv"

	"github.com/pions/pkg/stun"
	"github.com/pions/turn"
)

type MyTurnServer struct {
}

func (m *MyTurnServer) AuthenticateRequest(username string, srcAddr *stun.TransportAddr) (password string, ok bool) {
	return "password", true
}

func main() {
	m := &MyTurnServer{}

	realm := os.Getenv("REALM")
	if realm == "" {
		log.Panic("REALM is a required environment variable")
	}

	udpPortStr := os.Getenv("UDP_PORT")
	if udpPortStr == "" {
		log.Panic("UDP_PORT is a required environment variable")
	}
	udpPort, err := strconv.Atoi(udpPortStr)
	if err != nil {
		log.Panic(err)
	}

	turn.Start(turn.StartArguments{
		Server:  m,
		Realm:   realm,
		UDPPort: udpPort,
	})
}
