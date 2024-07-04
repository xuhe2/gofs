package p2p

import (
	"fmt"
	"io"
	"log"
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

// get remote address
func (p *TCPPeer) GetRemoteAddr() string {
	return p.conn.RemoteAddr().String()
}

func NewTCPPeer(conn net.Conn, outBound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outBound: outBound,
	}
}

// implement the interface, it can
func (p *TCPPeer) Close() error {
	return p.conn.Close()
}

type TCPTransportOpts struct {
	ListenAddress string
	ShakeHands    HandshakeFunc
	Decoder       Decoder
	OnPeer        func(Peer) error //it can be nil
}

type TCPTransport struct {
	TCPTransportOpts
	listener   net.Listener
	rpcChannel chan RPC

	mu    sync.RWMutex
	peers map[net.Addr]Peer
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcChannel:       make(chan RPC), //it is a message queue
	}
}

// close the tcp transport
func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.ListenAddress)
	if err != nil {
		return err
	}

	log.Printf("Listening on %s\n", t.ListenAddress)

	go t.startAcceptLoop() // start the accept loop in a separate goroutine

	return nil
}

// consume implement the interface, it is a read-only channel
// read the info from the peer in the network
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcChannel
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		// if user action close the fileServer
		// and use the Close() function to close the listener
		if err == net.ErrClosed {
			fmt.Printf("TCPTransport: listener closed\n")
			return
		}

		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
		}
		log.Printf("TCPTransport: accepted connection from %v\n", conn.RemoteAddr())
		// handle the connection in a separate goroutine
		go t.handleConnect(conn, false)
	}
}

// it has the for loop, so it is a blocking function
// use this function with `go` keyword to run it in a separate goroutine
func (t *TCPTransport) handleConnect(conn net.Conn, outBound bool) {
	var err error

	// close the connection if the function is closed
	defer func() {
		conn.Close()
	}()

	peer := NewTCPPeer(conn, outBound) // we are the inbound peer

	// perform the handshake
	if err = t.ShakeHands(peer); err != nil {
		fmt.Printf("Handshake failed: %v\n", err)
		return
	}

	// if `OnPeer` function not exists, do nothing
	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			fmt.Printf("OnPeer failed: %v\n", err)
			return
		}
	}

	message := RPC{}
	for {
		if err = t.Decoder.Decode(peer.conn, &message); err != nil {
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
		t.rpcChannel <- message
	}
}

// tcp `Dial` function to connect to the peer
func (t *TCPTransport) Dial(address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}
	// handle the connection
	go t.handleConnect(conn, true)

	return nil
}
