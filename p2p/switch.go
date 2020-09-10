package p2p

import (
	"math/rand"
)

// Switch handles peer connections and exposes an API to receive incoming messages
// on `Reactors`.  Each `Reactor` is responsible for handling incoming messages of one
// or more `Channels`.  So while sending outgoing messages is typically performed on the peer,
// incoming messages are received on the reactor.
type Switch struct {
	// service.BaseService

	// config       *config.P2PConfig
	// reactors     map[string]Reactor
	// chDescs      []*conn.ChannelDescriptor
	// reactorsByCh map[byte]Reactor
	// peers        *PeerSet
	// dialing      *cmap.CMap
	// reconnecting *cmap.CMap
	// nodeInfo     NodeInfo // our node info
	// nodeKey      *NodeKey // our node privkey
	// addrBook     AddrBook
	// peers addresses with whom we'll maintain constant connection
	// persistentPeersAddrs []*NetAddress
	// unconditionalPeerIDs map[ID]struct{}

	// transport Transport

	// filterTimeout time.Duration
	// peerFilters   []PeerFilterFunc

	rng *rand.Rand // seed for randomizing dial times and orders

	// metrics *Metrics
}
