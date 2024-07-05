package p2p

// peer is an interface for a remote node in the network.
type Peer interface {
	SendBytes([]byte) error
	GetRemoteAddr() string
	Close() error
}

// can handle the request
type Transport interface {
	ListenAndAccept() error //start listen and accept the network info
	Consume() <-chan RPC    // consume the info from network, use RPC.Payload to get the info
	Close() error           //close the transport
	Dial(string) error      //dial to a peer
	GetListenAddr() string  //get the PC's listen addr
}
