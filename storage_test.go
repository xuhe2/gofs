package main

import (
	"bytes"
	"testing"
)

func TestSHA1PathTransformFunc(t *testing.T) {
	// Test the SHA1PathTransformFunc function
	pathKey := SHA1PathTransformFunc("dataDir/file.txt")
	if pathKey.Path != "3aa32f5658/d036c968e1/2f4c33c478/f5fc214669" {
		t.Error("SHA1PathTransformFunc returned an unexpected path")
	}
}

func TestStorage(t *testing.T) {
	opts := StorageOpts{
		PathTransformFunc: SHA1PathTransformFunc,
	}
	storage := NewStorage(opts)

	data := bytes.NewReader([]byte("hello world"))
	if err := storage.writeStream("dataDir", data); err != nil {
		t.Error(err)
	}
}
