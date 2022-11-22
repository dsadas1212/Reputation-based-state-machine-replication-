package consensus

import (
	"github.com/adithyabhatkajake/libchatter/log"
	msg "github.com/adithyabhatkajake/libsynchs/msg"
	pb "github.com/golang/protobuf/proto"
)

// attack injection!!
// Leader propose two diferent proposal in this round
// note that equivocationpropsoe and withholdingpropose lead nodes this round
// commit empty block which means if equicocationpropose or withholding propose
// exists propsose() should be convert
// But the other two cases(malicious) are just the opposite
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
	go n.Broadcast(maliRelayMsg)

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
