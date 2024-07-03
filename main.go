package main

import (
	"fmt"

	"github.com/xuhe2/go-fs/p2p"
)

func main() {
	// create a TCP transport
	opts := p2p.TCPTransportOpts{
		ListenAddress: ":3000",
		Decoder:       p2p.DefaultDecoder{},
		ShakeHands:    p2p.NOPHandshake,
	}
	tCPTransport := p2p.NewTCPTransport(opts)

	// create a file server
	fileServerOpts := FileServerOpts{
		StorageRootFileName: "",
		PathTransformFunc:   SHA1PathTransformFunc,
		Transport:           tCPTransport,
	}

	fileServer := NewFileServer(fileServerOpts)
	//start the file server service
	if err := fileServer.Start(); err != nil {
		panic(err)
	}

	for {
		msg := <-fileServer.Transport.Consume()
		fmt.Printf("Received message: %v\n", msg.Payload)
	}

	select {}
}
