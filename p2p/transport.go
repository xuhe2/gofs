package p2p

// peer is an interface for a remote node in the network.
type Peer interface {
	Close() error
}

// can handle the request
type Transport interface {
	ListenAndAccept() error
	Consume() <-chan RPC
}
