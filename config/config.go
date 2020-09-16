package config

import (
	"fmt"
	"time"

	"github.com/adithyabhatkajake/libchatter/crypto"
	peerstore "github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

// Implement all the interfaces, i.e.,
// 1. net
// 2. crypto
// 3. config

// GetID returns the Id of this instance
func (shs *NodeConfig) GetID() uint64 {
	return shs.GetId()
}

// GetP2PAddrFromID gets the P2P address of the node rid
func (shs *NodeConfig) GetP2PAddrFromID(rid uint64) string {
	address := shs.GetNodeAddressMap()[rid]
	addr := fmt.Sprintf("/ip4/%s/tcp/%s", address.IP, address.Port)
	return addr
}

// GetMyKey returns the private key of this instance
func (shs *NodeConfig) GetMyKey() crypto.PrivKey {
	return shs.pvtKey
}

// GetPubKeyFromID returns the Public key of node whose ID is nid
func (shs *NodeConfig) GetPubKeyFromID(nid uint64) crypto.PubKey {
	return shs.nodeKeyMap[nid]
}

// GetPeerFromID returns libp2p peerInfo from the config
func (shs *NodeConfig) GetPeerFromID(nid uint64) peerstore.AddrInfo {
	pID, err := peerstore.IDFromPublicKey(shs.GetPubKeyFromID(nid))
	if err != nil {
		panic(err)
	}
	addr, err := ma.NewMultiaddr(shs.GetP2PAddrFromID(nid))
	if err != nil {
		panic(err)
	}
	pInfo := peerstore.AddrInfo{
		ID:    pID,
		Addrs: []ma.Multiaddr{addr},
	}
	return pInfo
}

// GetNumNodes returns the protocol size
func (shs *NodeConfig) GetNumNodes() uint64 {
	return shs.GetInfo().GetNodeSize()
}

// GetClientListenAddr returns the address where to talk to/from clients
func (shs *NodeConfig) GetClientListenAddr() string {
	id := shs.GetID()
	address := shs.GetNodeAddressMap()[id]
	addr := fmt.Sprintf("/ip4/%s/tcp/%s", address.IP, shs.ClientPort)
	return addr
}

// GetBlockSize returns the number of commands that can be inserted in one block
func (shs *NodeConfig) GetBlockSize() uint64 {
	return shs.GetInfo().GetBlockSize()
}

// GetDelta returns the synchronous wait time
func (shs *NodeConfig) GetDelta() time.Duration {
	timeInSeconds := shs.SyncHSConfig.GetDelta()
	return time.Duration(int(timeInSeconds*1000)) * time.Millisecond
}

// GetCommitWaitTime returns how long to wait before committing a block
func (shs *NodeConfig) GetCommitWaitTime() time.Duration {
	return shs.GetDelta() * 2
}

// GetNPBlameWaitTime returns how long to wait before sending the NP Blame
func (shs *NodeConfig) GetNPBlameWaitTime() time.Duration {
	return shs.GetDelta() * 3
}

// GetNumberOfFaultyNodes returns f
// We do this because f can be less than floor(n-1)/2
func (shs *NodeConfig) GetNumberOfFaultyNodes() uint64 {
	return shs.GetInfo().GetFaults()
}
