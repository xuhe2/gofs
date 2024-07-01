package main

import (
	"log"

	"github.com/xuhe2/go-fs/p2p"
)

func main() {
	opts := p2p.TCPTransportOpts{
		ListenAddress: ":3000",
		Decoder:       p2p.DefaultDecoder{},
		ShakeHands:    p2p.NOPHandshake,
	}
	tr := p2p.NewTCPTransport(opts)

	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	select {}
}
