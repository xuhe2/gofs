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

const (
	DefaultRootDirName = "dataRootDir"
)

type PathTransformFunc func(string) PathKey

func DefaultPathTransformFunc(key string) PathKey {
	return PathKey{
		Path:     key,
		FileName: key,
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

func (p PathKey) getPathFirstDirName() string {
	return strings.Split(p.getFullPath(), "/")[0]
}

type StorageOpts struct {
	RootDirName       string            //the name of root dir that contains all dir/file
	PathTransformFunc PathTransformFunc //from the key to the path to store the file
}

type Storage struct {
	StorageOpts
}

func NewStorage(opts StorageOpts) *Storage {
	// if the path transform func is not set, use the default one
	if opts.PathTransformFunc == nil {
		opts.PathTransformFunc = DefaultPathTransformFunc
	}
	// if the root dir name is empty
	if len(opts.RootDirName) == 0 {
		opts.RootDirName = DefaultRootDirName
	}
	return &Storage{
		StorageOpts: opts,
	}
}

func (s *Storage) writeStream(key string, r io.Reader) error {
	pathKey := s.PathTransformFunc(key)

	// get the path name with the root dir name
	pathWithRootDirName := s.RootDirName + "/" + pathKey.Path
	// if the dir is note exists
	if err := os.MkdirAll(pathWithRootDirName, os.ModePerm); err != nil {
		return err
	}

	// get the file name by using the md5
	fullPathWithRoorDirName := s.RootDirName + "/" + pathKey.getFullPath()

	f, err := os.Create(fullPathWithRoorDirName)
	if err != nil {
		return err
	}

	n, err := io.Copy(f, r)
	if err != nil {
		return err
	}

	fmt.Printf("Wrote %d bytes to %s\n", n, fullPathWithRoorDirName)

	return nil
}

func (s *Storage) readStream(key string) (io.ReadCloser, error) {
	pathKey := s.PathTransformFunc(key)
	fullPathWithRoorDirName := s.RootDirName + "/" + pathKey.getFullPath()
	return os.Open(fullPathWithRoorDirName)
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

// delete the specific file
func (s *Storage) Delete(key string) error {
	pathKey := s.PathTransformFunc(key)
	// get the dirst dir name with root dir name
	pathFirstDirNameWithRootDirName := s.RootDirName + "/" + pathKey.getPathFirstDirName()
	return os.RemoveAll(pathFirstDirNameWithRootDirName)
}

// check if the file exists
func (s *Storage) Has(key string) bool {
	pathKey := s.PathTransformFunc(key)
	// get full path name with root dir name
	fullPathWithRoorDirName := s.RootDirName + "/" + pathKey.getFullPath()
	if _, err := os.Stat(fullPathWithRoorDirName); err != nil {
		return false
	}
	return true
}

// clear the all data
func (s *Storage) Clear() error {
	return os.RemoveAll(s.RootDirName)
}
