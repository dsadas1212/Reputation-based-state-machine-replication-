package chain

import (
	"github.com/adithyabhatkajake/libchatter/crypto"
)

var (
	genesisBlock = &Block{
		Data: &BlockData{
			Index:    0,
			PrevHash: make([]byte, crypto.HashLen),
			Cmds:     nil,
		},
		Decision: true,
		Proposer: 1,
	}
)

// GetGenesis returns the genesis block for the chain
func GetGenesis() *Block {
	return genesisBlock
}

// Ensure that the blockhash is computed
func init() {
	genesisBlock.BlockHash = genesisBlock.GetHash().GetBytes()
}
