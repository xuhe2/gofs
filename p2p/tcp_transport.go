package p2p

import (
	"fmt"
	"net"
	"sync"
)

// the peer in the tcp transport connection
type TCPPeer struct {
	conn net.Conn
	// if we dial and accept, we are the outbound peer
	// if we accept and dial, we are the inbound peer
	outBound bool
}

func NewTCPPeer(conn net.Conn, outBound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outBound: outBound,
	}
}

type TCPTransport struct {
	listenAddress string
	listener      net.Listener
	shakeHands    HandshakeFunc
	decoder       Decoder

	mu    sync.RWMutex
	peers map[net.Addr]Peer
}

func NewTCPTransport(listenAddress string) *TCPTransport {
	return &TCPTransport{
		shakeHands:    NOPHandshake,
		listenAddress: listenAddress,
	}
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.listenAddress)
	if err != nil {
		return err
	}

	go t.startAcceptLoop() // start the accept loop in a separate goroutine

	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
		}
		// handle the connection in a separate goroutine
		go t.handleConnect(conn)
	}
}

func (t *TCPTransport) handleConnect(conn net.Conn) {
	peer := NewTCPPeer(conn, false) // we are the inbound peer

	// perform the handshake
	if err := t.shakeHands(peer); err != nil {
		fmt.Printf("Handshake failed: %v\n", err)
		return
	}

	message := new(interface{})
	for {
		if err := t.decoder.Decode(conn, message); err != nil {
			fmt.Printf("Error decoding message: %v\n", err)
			continue
		}
		// handle the message
		// ...
	}

	fmt.Printf("Peer connected from %v\n", peer.conn.RemoteAddr())
}
