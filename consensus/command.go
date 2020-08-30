package consensus

import (
	"github.com/adithyabhatkajake/libsynchs/chain"
)

func (n *SyncHS) addCmdsAndProposeIfSufficientCommands(cmd *chain.Command) {
	n.cmdMutex.Lock()
	n.pendingCommands = append(n.pendingCommands, cmd)
	// n.pendingCommands.PushBack(cmd)
	if uint64(len(n.pendingCommands)) >= n.config.GetBlockSize() &&
		// Sufficient Commands
		n.config.GetID() == n.leader { // And I am the leader
		go n.propose()
	}
	n.cmdMutex.Unlock()
}

func (n *SyncHS) getCmdsIfSufficient() ([]*chain.Command, bool) {
	blkSize := n.config.GetBlockSize()
	n.cmdMutex.Lock()
	defer n.cmdMutex.Unlock()
	numCmds := uint64(len(n.pendingCommands))
	if numCmds < blkSize {
		return nil, false
	}
	cmds := make([]*chain.Command, blkSize)
	// Copy slice blkSize commands from pending Commands
	copy(cmds, n.pendingCommands[numCmds-blkSize:])
	// Update old slice
	n.pendingCommands = n.pendingCommands[:numCmds-blkSize]
	return cmds, true
}
