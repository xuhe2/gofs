package main

import (
	"fmt"

	"github.com/xuhe2/go-fs/p2p"
)

type FileServerOpts struct {
	StorageRootFileName string
	PathTransformFunc   PathTransformFunc
	Transport           p2p.Transport
}

type FileServer struct {
	FileServerOpts

	storage           *Storage
	quitSignalChannel chan struct{}
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
	}
}

// start the network connection
func (s *FileServer) Start() error {
	// start the network connection and accept the connection
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}

	s.runMainTaskLoop()

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
