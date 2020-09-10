package consensus

import (
	"github.com/adithyabhatkajake/libchatter/crypto"
	"github.com/adithyabhatkajake/libsynchs/chain"
)

func (n *SyncHS) getBlock(bhash crypto.Hash) *chain.ExtBlock {
	n.bc.Mu.RLock()
	defer n.bc.Mu.RUnlock()

	return n.bc.BlocksByHash[bhash]
}
