package main

import (
	"fmt"
	"log"

	"github.com/xuhe2/go-fs/p2p"
)

func OnPeer(p2p.Peer) error {
	fmt.Println("new peer connected")
	// do something
	return nil
}

func main() {
	opts := p2p.TCPTransportOpts{
		ListenAddress: ":3000",
		Decoder:       p2p.DefaultDecoder{},
		ShakeHands:    p2p.NOPHandshake,
		OnPeer:        OnPeer,
	}
	tr := p2p.NewTCPTransport(opts)

	f := func() {
		for {
			msg := <-tr.Consume()
			fmt.Printf("received message: %s\n", msg)
		}
	}
	go f()

	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	select {}
}
