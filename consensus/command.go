package consensus

func (n *SyncHS) addCmdsAndStartTimerIfSufficientCommands(cmd []byte) {
	n.cmdMutex.Lock()
	defer n.cmdMutex.Unlock()
	n.pendingCommands = append(n.pendingCommands, cmd)
	// n.pendingCommands.PushBack(cmd)
	//&& n.GetID() == n.leader
	if uint64(len(n.pendingCommands)) >= n.GetBlockSize() {
		// log.Info("node", n.GetID(), "'s pendingCommands len is", len(n.pendingCommands))
		if n.gcallFuncFinish {
			// n.startConsensusTimerWithWithhold()
			n.startConsensusTimer()
			n.gcallFuncFinish = false
		}
		//16
		if len(n.SyncChannel) == 1 {
			<-n.SyncChannel
			// time.Sleep(time.Second * 5)
			n.startConsensusTimer()
		}
	}
	//
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
