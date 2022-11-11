package consensus

import (
	"time"
)

func (n *SyncHS) addCmdsAndStartTimerIfSufficientCommands(cmd []byte) {
	n.cmdMutex.Lock()
	defer n.cmdMutex.Unlock()
	n.pendingCommands = append(n.pendingCommands, cmd)
	// .pendingCommands.PushBack(cmd)
	//&& n.GetID() == n.leader
	if uint64(len(n.pendingCommands)) >= n.GetBlockSize() { //Sufficient Commands start our timer!
		// change the condition of this: 1. if this block is gensis block start timer directly
		if n.gcallFuncFinish {
			go n.startConsensusTimer()
			time.Sleep(time.Second * 3)
			n.gcallFuncFinish = false
		}
		//16
		if len(n.SyncChannel) == 1 {
			<-n.SyncChannel
			n.startConsensusTimer()
		}
		// go n.startConsensusTimer()
		// if uint64(len(n.pendingCommands)) >= n.GetBlockSize() {
		// 	go n.gnerateCandidateBlock()
		// }

		// }
	}
}

// !!TODO all pengdingCommands should be change not only  leader?
func (n *SyncHS) getCmdsIfSufficient() ([][]byte, bool) {
	blkSize := n.GetBlockSize()
	n.cmdMutex.Lock()
	defer n.cmdMutex.Unlock()
	numCmds := uint64(len(n.pendingCommands))
	if numCmds < blkSize {
		return nil, false
	}
	cmds := make([][]byte, blkSize)
	// Copy slice blkSize commands from pending Commands
	copy(cmds, n.pendingCommands[numCmds-blkSize:])
	// Update old slice
	n.pendingCommands = n.pendingCommands[:numCmds-blkSize]
	return cmds, true
}
