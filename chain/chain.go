package chain

import (
	"github.com/adithyabhatkajake/libchatter/crypto"
)

// NewChain returns an empty chain
func NewChain() *BlockChain {
	c := &BlockChain{}
	// genesis := GetGenesis()

	c.BlocksByHash = make(map[crypto.Hash]*ExtBlock)
	c.BlocksByHeight = make(map[uint64]*ExtBlock)

	// Set genesis block as the first block
	// c.HeightBlockMap[genesis.Data.Index] = genesis
	// c.Chain[genesis.GetHash()] = genesis
	c.Head = 0

	return c
}
