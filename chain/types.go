package chain

import (
	"sync"

	"github.com/adithyabhatkajake/libchatter/crypto"
)

// BlockChain is what we call a blockchain
type BlockChain struct {
	BlocksByHash map[crypto.Hash]*ExtBlock
	// A lock that we use to safely update the chain
	Mu sync.RWMutex
	// A heeight block map
	BlocksByHeight map[uint64]*ExtBlock
	// Chain head
	Head uint64
}

// Block is an Ethereum Block
type Block interface {
	GetHeader() Header
	GetBody() Body
	GetSize() uint64
	// GetBlockHash returns the hash of the block (i.e header)
	GetBlockHash() crypto.Hash
	IsValid() bool
}

// Header is an Ethereum Header
type Header interface {
	// GetParentHash returns the hash of the parent block of this block
	GetParentHash() crypto.Hash
	// GetTxHash returns the hash of all the transactions in the block
	GetTxHash() crypto.Hash
	// GetHeight returns the height of this block
	GetHeight() uint64
	// GetExtradata returns extra data from the block
	GetExtradata() []byte
}

// Body is an Ethereum Body
type Body interface {
	GetTransactions() [][]byte // We are agnostic of what the transaction is
}
