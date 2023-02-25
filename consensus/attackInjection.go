package consensus

import (
	"time"

	"github.com/adithyabhatkajake/libchatter/log"
	msg "github.com/adithyabhatkajake/libsynchs/msg"
	pb "google.golang.org/protobuf/proto"
)

// attack injection!!
// Leader propose two diferent proposal in this round
// note that equivocationpropsoe and withholdingpropose lead nodes this round
// commit empty block which means if equicocationpropose or withholding propose
// exists propsose() should be convert
// But the other two cases(malicious) are just the opposite

// swithch the case case1: best-case/case2: withholding block/case3 : equivocating block/case4(5) : maliciousblock(vote)
// withholding case = normal case
func (n *SyncHS) startConsensusTimerWithWithhold() {
	n.timer.Start()
	log.Debug(n.GetID(), " start a 4Delta timer ", time.Now(), "IN ROOUND", n.view)
	go func() {
		if n.GetID() == n.leader {
			// if n.GetID()%3 == 0 && n.GetID() != 0
			if n.GetID()%2 != 0 {
				n.Withholdingpropose()
				n.withholdingProposalInject = true
				time.After(time.Second * 2)
				n.handleWithholdingProposal()

			} else {
				n.withholdingProposalInject = false
				n.Propose()
			}

		} else {
			//non leader node update its command pool
			// n.cmdMutex.Lock()
			// n.pendingCommands = n.pendingCommands[:uint64(len(n.pendingCommands))-n.GetBlockSize()]
			// n.cmdMutex.Unlock()
			// if n.leader%3 == 0 && n.leader != 0
			if n.leader%2 != 0 {
				n.withholdingProposalInject = true
				time.After(time.Second * 3)
				n.handleWithholdingProposal()
			} else {
				n.withholdingProposalInject = false
			}

		}
	}()
}

// equivocating case
func (n *SyncHS) startConsensusTimerWithEquivocation() {
	n.timer.Start()
	log.Debug(n.GetID(), " start a 4Delta timer ", time.Now(), "IN ROOUND", n.view)
	go func() {
		if n.leader == n.GetID() {
			// n.GetID()%3 == 0 && n.GetID() != 0
			if n.GetID()%2 != 0 {
				n.Equivocationpropose()
			} else {
				n.Propose()
			}

		}
	}()

}

// malicious prospoal case
func (n *SyncHS) startConsensusTimerWithMaliciousPropsoal() {

	n.timer.Start()
	log.Debug(n.GetID(), " start a 4Delta timer ", time.Now(), "IN ROOUND", n.view)
	go func() {
		if n.GetID() == n.leader {
			n.Propose()

		} else {
			// //non leader node update its command pool
			// n.pendingCommands = n.pendingCommands[:uint64(len(n.pendingCommands))-n.GetBlockSize()]
			if n.GetID()%2 != 0 {
				n.Maliciousproposalpropose()
			}
		}

	}()

}

// malicious vote case
func (n *SyncHS) startConsensusTimerWithMaliciousVote() {

	n.timer.Start()
	log.Debug(n.GetID(), " start a 4Delta timer ", time.Now(), "IN ROOUND", n.view)
	go func() {
		if n.GetID() == n.leader {
			n.Propose()
		} else {
			// //non leader node update its command pool
			// n.pendingCommands = n.pendingCommands[:uint64(len(n.pendingCommands))-n.GetBlockSize()]
			if n.GetID()%2 != 0 {
				n.maliciousVoteInject = true
			}
		}

	}()
}

// malicious vote case need to change the step in forwardhandler
func (n *SyncHS) Equivocationpropose() {
	log.Info("leader", n.GetID(), "equivocating block in view(round)", n.view)
	n.bc.Mu.Lock()
	defer n.bc.Mu.Unlock()
	head := n.bc.Head
	cert, exists := n.getCertForBlockIndex(head)
	if !exists {
		log.Debug("The head does not have a certificate")
		log.Debug("Cancelling the proposal")
		return
	}
	cmds, isSuff := n.getCmdsIfSufficient()
	if !isSuff {
		log.Debug("Insufficient commands, aborting the proposal")
		return
	}
	//for simply, same cmds load in two equivocation block
	extra1 := []byte{'1'}
	extra2 := []byte{'2'}
	prop1 := n.NewCandidateProposal(cmds, cert, n.bc.Head+1, extra1)
	prop2 := n.NewCandidateProposal(cmds, cert, n.bc.Head+1, extra2)
	log.Trace("Finished prepare Proposing")
	relayMsg1 := &msg.SyncHSMsg{}
	relayMsg2 := &msg.SyncHSMsg{}
	relayMsg1.Msg = &msg.SyncHSMsg_Prop{Prop: prop1}
	relayMsg2.Msg = &msg.SyncHSMsg_Prop{Prop: prop2}
	go n.EquivocatingBroadcast(relayMsg1, relayMsg2)

}

// leader withholding his proposal
func (n *SyncHS) Withholdingpropose() {
	log.Info("leader", n.GetID(), " witholding block in view(round)", n.view)

}

// non-leader node propose propsoal
func (n *SyncHS) Maliciousproposalpropose() {
	log.Info("node", n.GetID(), "propose an invaild propsoal in view(round)", n.view)
	n.bc.Mu.Lock()
	defer n.bc.Mu.Unlock()
	head := n.bc.Head
	cert, exists := n.getCertForBlockIndex(head)
	if !exists {
		log.Debug("The head does not have a certificate")
		log.Debug("Cancelling the proposal")
		return
	}
	cmds, isSuff := n.getCmdsIfSufficient()
	if !isSuff {
		log.Debug("Insufficient commands, aborting the proposal")
		return
	}
	maliProp := n.NewCandidateProposal(cmds, cert, n.bc.Head+1, nil)
	maliRelayMsg := &msg.SyncHSMsg{}
	maliRelayMsg.Msg = &msg.SyncHSMsg_Prop{Prop: maliProp}
	//we need other channel to handle maliciousprospoal
	go func() {
		//malicious broadcast
		n.Broadcast(maliRelayMsg)
		//hanlde myself malicious proposal
		n.maliPropseChannel <- maliProp

	}()

}

func (n *SyncHS) MaliciousPropsoalHandler() {
	for {
		malipro, ok := <-n.maliPropseChannel
		if !ok {
			log.Error("MaliciousProposal channel error")
			continue
		}
		log.Debug("Node", n.GetID(), "Receive ", malipro.GetMiner(), "'s MaliciousProposal in round", n.view)
		//this misbehaviour only ocur in propose step,so we only detect the sign of miner
		data, _ := pb.Marshal(malipro.GetBlock().GetHeader())
		correct, err := n.GetPubKeyFromID(malipro.GetMiner()).Verify(data, malipro.GetMiningProof())
		if !correct || err != nil {
			log.Error("Incorrect signature for maliciousproposal ", n.view)
			continue
		}
		//malicous proposal
		n.maliciousProposalInject = malipro.Miner != n.leader
		if !n.maliciousProposalInject {
			log.Debug("There is an invalid mailiciouspropsoal")
			continue
		}
		log.Info("There is a malicious propsoal")
		if n.GetID() != 0 {
			//misbehaviourtarget don't send its evidence!
			n.sendMaliProEvidence(malipro)
			continue
		} else {
			//misbehaviourtarget only need to handle its misbehave!
			continue
		}

	}
}

// node vote for an non-leader block
func (n *SyncHS) voteForNonLeaderBlk() {
	// create a invalid block hash
	log.Info("NODE", n.GetID(), "is voting for a nonexistent block")
	pvd := &msg.ProtoVoteData{
		BlockHash: []byte{'I', 'n', 'v', 'a', 'l', 'i', 'd'},
		View:      n.view,
		Owner:     n.GetID(),
	}
	data, err := pb.Marshal(pvd)
	if err != nil {
		log.Error("Error marshing vote data during voting")
		log.Error(err)
		return
	}
	sig, err := n.GetMyKey().Sign(data)
	if err != nil {
		log.Error("Error signing vote")
		log.Error(err)
		return
	}
	pvb := &msg.ProtoVoteBody{
		Voter:     n.GetID(),
		Signature: sig,
	}
	pv := &msg.ProtoVote{
		Body: pvb,
		Data: pvd,
	}
	voteMsg := &msg.SyncHSMsg{}
	voteMsg.Msg = &msg.SyncHSMsg_Vote{Vote: pv}
	v := &msg.Vote{}
	v.FromProto(pv)
	n.addVotetoMap(pv)
	go func() {
		//the voter change his voteMap by himself
		// Send vote to all the nodes
		n.Broadcast(voteMsg)
		// Handle my own vote
		n.voteChannel <- v

	}()

}

//
