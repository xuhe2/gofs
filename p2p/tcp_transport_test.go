package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T) {
	listenAddress := ":4000"

	opts := &TCPTransportOpts{
		ListenAddress: listenAddress,
	}

	tr := NewTCPTransport(*opts)

	assert.Equal(t, tr.ListenAddress, listenAddress)

	// server
	assert.Nil(t, tr.ListenAndAccept())
}
