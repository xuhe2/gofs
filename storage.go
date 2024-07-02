package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

type PathTransformFunc func(string) PathKey

func DefaultPathTransformFunc(key string) PathKey {
	return PathKey{
		Path:     key,
		FileName: "",
	}
}

func SHA1PathTransformFunc(key string) PathKey {
	// get the hash of the key
	hash := sha1.Sum([]byte(key))
	hashString := hex.EncodeToString(hash[:])

	// split the hash into 10-character blocks
	blockSize := 10
	paths := make([]string, len(hashString)/blockSize)
	for i := 0; i < len(hashString); i += blockSize {
		paths[i/blockSize] = hashString[i : i+blockSize]
	}
	// join the blocks with slashes
	return PathKey{
		Path:     strings.Join(paths, "/"),
		FileName: hashString,
	}
}

type PathKey struct {
	Path     string
	FileName string
}

func (p PathKey) getFullPath() string {
	return fmt.Sprintf("%s/%s", p.Path, p.FileName)
}

type StorageOpts struct {
	PathTransformFunc PathTransformFunc //from the key to the path to store the file
}

type Storage struct {
	StorageOpts
}

func NewStorage(opts StorageOpts) *Storage {
	return &Storage{
		StorageOpts: opts,
	}
}

func (s *Storage) writeStream(key string, r io.Reader) error {
	pathKey := s.PathTransformFunc(key)

	// if the dir is note exists
	if err := os.MkdirAll(pathKey.Path, os.ModePerm); err != nil {
		return err
	}

	// get the file name by using the md5
	fullPath := pathKey.getFullPath()

	f, err := os.Create(fullPath)
	if err != nil {
		return err
	}

	n, err := io.Copy(f, r)
	if err != nil {
		return err
	}

	fmt.Printf("Wrote %d bytes to %s\n", n, fullPath)

	return nil
}

func (s *Storage) readStream(key string) (io.ReadCloser, error) {
	pathKey := s.PathTransformFunc(key)
	return os.Open(pathKey.getFullPath())
}

func (s *Storage) Read(key string) (io.Reader, error) {
	// open the file
	f, err := s.readStream(key)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// read the content
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(f)
	if err != nil {
		return nil, err
	}
	return buf, err
}
