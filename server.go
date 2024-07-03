package main

import "github.com/xuhe2/go-fs/p2p"

type FileServerOpts struct {
	StorageRootFileName string
	PathTransformFunc   PathTransformFunc
	Transport           p2p.Transport
}

type FileServer struct {
	FileServerOpts

	storage *Storage
}

func NewFileServer(fileServerOpts FileServerOpts) *FileServer {
	storageOpts := StorageOpts{
		RootDirName:       fileServerOpts.StorageRootFileName,
		PathTransformFunc: fileServerOpts.PathTransformFunc,
	}
	return &FileServer{
		FileServerOpts: fileServerOpts,
		storage:        NewStorage(storageOpts),
	}
}

// start the network connection
func (s *FileServerOpts) Start() error {
	// start the network connection and accept the connection
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}
	return nil
}
