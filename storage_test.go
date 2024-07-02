package main

import (
	"bytes"
	"io"
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

	key := "dataDir"
	data := bytes.NewReader([]byte("hello world"))
	if err := storage.writeStream(key, data); err != nil {
		t.Error(err)
	}

	r, err := storage.Read(key)
	if err != nil {
		t.Error(err)
	}

	b, _ := io.ReadAll(r)
	if string(b) != "hello world" {
		t.Error("Expected 'hello world', got", string(b))
	}
}
