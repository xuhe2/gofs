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
	BootstrapNodes      []string // the peer need to connects
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

type Message struct {
	From    string
	Payload any //use decoder can not use `any`
}

type MessagePayload struct {
	Key  string
	Data []byte
}

// broadcast the data to the network
func (s *FileServer) broadcastData(msg *Message) error {
	buf := new(bytes.Buffer)
	// encode the payload to the buffer
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		fmt.Println("encode payload failed:", err)
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
	// copy the reader
	buf := new(bytes.Buffer)
	tee := io.TeeReader(r, buf)

	// store the data in the storage
	if err := s.storage.Write(key, tee); err != nil {
		return err
	}
	// and broadcast the data to the network
	// we broadcast the data, so the address is the listenaddress of the server
	msg := &Message{
		From: s.Transport.GetListenAddr(),
		Payload: MessagePayload{
			Key:  key,
			Data: buf.Bytes(),
		},
	}
	return s.broadcastData(msg)
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
		case rpc := <-s.Transport.Consume():
			msg := &Message{}
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(msg); err != nil {
				log.Fatalf("failed to decode the payload: %v\n", err)
			}
			// handle the msg
			if err := s.handleMessage(msg); err != nil {
				log.Fatalf("failed to handle the message: %v\n", err)
			}
		case <-s.quitSignalChannel:
			// if main program send the quit signal
			//stop the main task loop
			return
		}
	}
}

// handle the payload
// receive a pointer can perform better
func (s *FileServer) handleMessage(message *Message) error {
	switch v := message.Payload.(type) {
	case *MessagePayload:
		// if the payload is a MessagePayload
		fmt.Printf("receive a message from %+v\n", v)
	}
	return nil
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

func init() {
	gob.Register(&Message{})
	gob.Register(&MessagePayload{})
}
