package chain

import (
	"github.com/adithyabhatkajake/libchatter/crypto"
	ethcore "github.com/ethereum/go-ethereum/core"
	ethdb "github.com/ethereum/go-ethereum/ethdb"
)

// SetupGenesis creates a default genesis on the db
func SetupGenesis(db ethdb.Database) {
	ethcore.SetupGenesisBlock(db, nil)
}

var (
	GenesisHash = crypto.Hash{}
)
