package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/xuhe2/go-fs/p2p"
)

type FileServerOpts struct {
	StorageRootFileName string
	PathTransformFunc   PathTransformFunc
	Transport           p2p.Transport
	BootstrapNodes      []string // the peer need to connect
}

type FileServer struct {
	FileServerOpts

	peers    map[string]p2p.Peer
	peerLock sync.Mutex

	storage           *Storage
	quitSignalChannel chan struct{} //receive the quit signal from main program to stop main task loop
}

func NewFileServer(fileServerOpts FileServerOpts) *FileServer {
	storageOpts := StorageOpts{
		RootDirName:       fileServerOpts.StorageRootFileName,
		PathTransformFunc: fileServerOpts.PathTransformFunc,
	}
	return &FileServer{
		FileServerOpts:    fileServerOpts,
		storage:           NewStorage(storageOpts),
		quitSignalChannel: make(chan struct{}),
		peers:             make(map[string]p2p.Peer),
	}
}

// start the network connection
func (s *FileServer) Start() error {
	// start the network connection and accept the connection
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}

	// connect to bootstrap nodes
	s.connectBootstrapNodesNetwork()

	// start the main task loop
	s.runMainTaskLoop()

	return nil
}

type Payload struct {
	Key  string
	Data []byte
}

// broadcast the data to the network
func (s *FileServer) broadcastData(p *Payload) error {
	buf := new(bytes.Buffer)
	// encode the payload to the buffer
	if err := gob.NewEncoder(buf).Encode(p); err != nil {
		return err
	}

	// send the payload to all the peers
	for _, peer := range s.peers {
		// send the payload to the peer
		if err := peer.SendBytes(buf.Bytes()); err != nil {
			log.Printf("send data to peer %s failed: %s", peer.GetRemoteAddr(), err)
		}
	}
	return nil
}

// store the data in the storage
// and broadcast the data to the network
func (s *FileServer) StoreData(key string, r io.Reader) error {
	// store the data in the storage
	if err := s.storage.Write(key, r); err != nil {
		return err
	}
	// and broadcast the data to the network
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r); err != nil {
		return err
	}

	p := &Payload{
		Key:  key,
		Data: buf.Bytes(),
	}
	return s.broadcastData(p)
}

func (s *FileServer) OnPeer(peer p2p.Peer) error {
	// add a peerlock to avoid the race condition
	// when the peer is added, we need to lock the peer map
	s.peerLock.Lock()
	defer s.peerLock.Unlock()

	// check the peers
	peerRemoteAddress := peer.GetRemoteAddr()
	if _, ok := s.peers[peerRemoteAddress]; ok {
		// if the peer is already in the map, we return the error
		return fmt.Errorf("peer %s is already in the map", peerRemoteAddress)
	}

	// if the peer is not in the map, we add it to the map
	s.peers[peerRemoteAddress] = peer
	return nil
}

// run the loop to handle the incoming connection
// process the info from network
func (s *FileServer) runMainTaskLoop() {
	// when the main task is finished by user action
	// we close the fileServer and its Transport's listener
	defer func() {
		fmt.Printf("FileServer quit\n")
		s.Transport.Close()
	}()

	for {
		select {
		case msg := <-s.Transport.Consume():
			fmt.Println("received msg:", msg)
		case <-s.quitSignalChannel:
			// if main program send the quit signal
			//stop the main task loop
			return
		}
	}
}

// close the fileServer's main task
func (s *FileServer) Stop() {
	close(s.quitSignalChannel)
}

// connect to bootstrap nodes
// if the connection is successful, the fileServer will start to listen and accept the connection
func (s *FileServer) connectBootstrapNodesNetwork() error {
	if len(s.BootstrapNodes) == 0 {
		return nil
	}

	for _, bootstrapNodeAddress := range s.BootstrapNodes {
		// use `go routine` can avoid the blocking of the main task loop
		// but we can not return err in the `go routine`
		go func(address string) {
			if err := s.Transport.Dial(address); err != nil {
				log.Printf("failed to connect to bootstrap node %s: %v\n", address, err)
			}
		}(bootstrapNodeAddress)
	}
	return nil
}
