package p2p

// peer is an interface for a remote node in the network.
type Peer interface {
}

// can handle the request
type Transport interface {
	ListenAndAccept() error
}
