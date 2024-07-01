package p2p

import (
	"fmt"
	"io"
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

type TCPTransportOpts struct {
	ListenAddress string
	ShakeHands    HandshakeFunc
	Decoder       Decoder
}

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener

	mu    sync.RWMutex
	peers map[net.Addr]Peer
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
	}
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.ListenAddress)
	if err != nil {
		return err
	}

	go t.startAcceptLoop() // start the accept loop in a separate goroutine

	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		fmt.Printf("Accepted connection from %v\n", conn.RemoteAddr())
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
	if err := t.ShakeHands(peer); err != nil {
		fmt.Printf("Handshake failed: %v\n", err)
		peer.conn.Close()
		return
	}

	message := &Message{}
	for {
		if err := t.Decoder.Decode(peer.conn, message); err != nil {
			if err == io.EOF {
				fmt.Printf("Connection closed by peer: %v\n", peer.conn.RemoteAddr())
				return
			}
			fmt.Printf("Error decoding message: %v\n", err)
			continue
		}
		// set the network address in order to send back info
		message.From = peer.conn.RemoteAddr()

		// handle the message
		fmt.Printf("Received message: %v\n", string(message.Payload))
	}
}
