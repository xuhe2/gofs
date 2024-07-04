package main

import (
	"bytes"
	"time"

	"github.com/xuhe2/go-fs/p2p"
)

func createFileServer(ListenAddress string, StorageRootFileName string, BootstrapNodes []string) *FileServer {
	// create a TCP transport
	opts := p2p.TCPTransportOpts{
		ListenAddress: ListenAddress,
		Decoder:       p2p.DefaultDecoder{},
		ShakeHands:    p2p.NOPHandshake,
	}
	tCPTransport := p2p.NewTCPTransport(opts)

	// create a file server
	fileServerOpts := FileServerOpts{
		StorageRootFileName: StorageRootFileName,
		PathTransformFunc:   SHA1PathTransformFunc,
		Transport:           tCPTransport,
		BootstrapNodes:      BootstrapNodes,
	}

	fileServer := NewFileServer(fileServerOpts)
	tCPTransport.OnPeer = fileServer.OnPeer

	return fileServer
}

func main() {
	// fileServer := createFileServer(":3000", "", []string{})
	// //start the file server service
	// if err := fileServer.Start(); err != nil {
	// 	panic(err)
	// }

	// test
	go func() {
		fileServer1 := createFileServer(":3000", "", []string{})
		//start the file server service
		if err := fileServer1.Start(); err != nil {
			panic(err)
		}
	}()

	go func() {
		fileServer2 := createFileServer(":4000", "", []string{":3000"})
		// if err := fileServer2.Start(); err != nil {
		// 	panic(err)
		// }

		go fileServer2.Start()
		time.Sleep(time.Second)
		reader := bytes.NewReader([]byte("hello world"))
		fileServer2.StoreData("hello.txt", reader)
	}()

	select {}
}
