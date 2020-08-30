package consensus

import (
	"github.com/adithyabhatkajake/libchatter/log"
	"github.com/adithyabhatkajake/libchatter/util"
	msg "github.com/adithyabhatkajake/libsynchs/msg"
	pb "github.com/golang/protobuf/proto"
)

func (shs *SyncHS) startBlameTimer() {
	log.Info("Starting Blame Timer")
	shs.blTimerLock.Lock()
	defer shs.blTimerLock.Unlock()
	if shs.blTimer != nil {
		shs.blTimer.Reset()
		log.Debug("Blame timer reset")
		return
	}
	shs.blTimer = util.NewTimer(func() {
		shs.sendNPBlame()
	})
	shs.blTimer.SetTime(shs.config.GetNPBlameWaitTime())
	shs.blTimer.Start()
	log.Debug("Finished starting a blame timer")
}

func (shs *SyncHS) stopBlameTimer() {
	log.Info("Stopping Blame Timer")
	shs.blTimerLock.Lock()
	defer shs.blTimerLock.Unlock()
	if shs.blTimer != nil {
		shs.blTimer.Cancel()
		// Set blTimer to nil because the blame timer is not reusable.
		shs.blTimer = nil
		// NOTE: I hope the golang gc frees the memory later
	}
}

func (shs *SyncHS) resetBlameTimer() {
	log.Info("Resetting Blame Timer")
	shs.blTimerLock.Lock()
	shs.blTimer.Reset() // Reset the timer
	shs.blTimerLock.Unlock()
}

/* sendNPBlame sends NoProgress blame.
This is usually triggered when the blame timer for the leader in a view, times out. */
func (shs *SyncHS) sendNPBlame() {
	log.Warn("Sending an NP-Blame message")
	blame := &msg.NoProgressBlame{}
	blame.Blame = &msg.Blame{}
	blame.Blame.BlData = &msg.BlameData{}
	blame.Blame.BlData.BlameTarget = shs.leader
	blame.Blame.BlData.View = shs.view
	blame.Blame.BlOrigin = shs.config.GetID()
	data, err := pb.Marshal(blame.Blame.BlData)
	if err != nil {
		log.Errorln("Error marshalling blame message", err)
		return
	}
	blame.Blame.Signature, err = shs.config.GetMyKey().Sign(data)
	if err != nil {
		log.Errorln("Error Signing the blame message", err)
	}
	blMsg := &msg.SyncHSMsg{}
	blMsg.Msg = &msg.SyncHSMsg_Npblame{Npblame: blame}
	shs.Broadcast(blMsg)
	go shs.handleNoProgressBlame(blame) // Process self blame
}

func (shs *SyncHS) handleNoProgressBlame(bl *msg.NoProgressBlame) {
	log.Trace("Received a No-Progress blame against ",
		bl.Blame.BlData.BlameTarget, " from ", bl.Blame.BlOrigin)
	// Check if the blame is correct
	isValid := shs.isNPBlameValid(bl)
	if !isValid {
		log.Debugln("Received an invalid blame message", bl.String())
		return
	}
	view := bl.Blame.BlData.View
	// Add it to the blame map
	shs.blLock.Lock()
	_, exists := shs.blameMap[view]
	if !exists {
		shs.blameMap[view] = make(map[uint64]*msg.Blame)
	}
	_, exists = shs.blameMap[view][bl.Blame.BlOrigin]
	if !exists {
		shs.blameMap[view][bl.Blame.BlOrigin] = bl.Blame
	}
	// Check if the blame map has sufficient blames
	// If there are more than f blames, then initiate quit view
	numBlames := uint64(len(shs.blameMap[view]))
	if numBlames > shs.config.GetNumberOfFaultyNodes() {
		// Quit the View
		go shs.QuitView()
	}
	shs.blLock.Unlock()
}

func (shs *SyncHS) isNPBlameValid(bl *msg.NoProgressBlame) bool {
	log.Traceln("Function isNPBlameValid with input", bl.String())
	// Check if the blame is for the current leader
	if bl.Blame.BlData.BlameTarget != shs.leader {
		log.Debug("Invalid Blame Target. Found", bl.Blame.BlData.BlameTarget,
			",Expected:", shs.leader)
		return false
	}
	// Check if the view is correct!
	if bl.Blame.BlData.View != shs.view {
		log.Debug("Invalid Blame View. Found", bl.Blame.BlData.View,
			",Expected:", shs.view)
		return false
	}
	// Get bl data
	data, err := pb.Marshal(bl.Blame.BlData)
	if err != nil {
		log.Debug("Error Marshalling blame message")
		return false
	}
	// Check if the signature is correct
	isSigValid, err := shs.config.GetPubKeyFromID(
		bl.Blame.BlOrigin).Verify(data, bl.Blame.Signature)
	if !isSigValid || err != nil {
		log.Debug("Invalid signature for blame message")
		return false
	}
	return true
}
