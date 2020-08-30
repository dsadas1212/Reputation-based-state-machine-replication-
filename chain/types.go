package chain

import (
	"sync"

	"github.com/adithyabhatkajake/libchatter/crypto"
)

// BlockChain is what we call a blockchain
type BlockChain struct {
	Chain map[crypto.Hash]*Block
	// A lock that we use to safely update the chain
	ChainLock sync.RWMutex
	// A height block map
	HeightBlockMap map[uint64]*Block
	// Unconfirmed Blocks
	UnconfirmedBlocks map[crypto.Hash]*Block
	// Chain head
	Head uint64
}
