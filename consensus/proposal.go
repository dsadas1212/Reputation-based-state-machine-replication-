package consensus

import (
	"time"

	"github.com/adithyabhatkajake/libchatter/crypto"
	"github.com/adithyabhatkajake/libchatter/log"
	"github.com/adithyabhatkajake/libchatter/util"
	"github.com/adithyabhatkajake/libsynchs/chain"
	msg "github.com/adithyabhatkajake/libsynchs/msg"
	pb "github.com/golang/protobuf/proto"
)

// Monitor pending commands, if there is any change and the current node is the leader, then start proposing blocks
func (n *SyncHS) propose() {
	log.Info("node", n.GetID(), "is proposing block")
	log.Debug("Starting a propose step")
	// Do we have a certificate?
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
	n.bc.Head++
	newHeight := n.bc.Head
	prop := n.NewCandidateProposal(cmds, cert, newHeight, nil)
	block := &chain.ExtBlock{}
	block.FromProto(prop.Block)
	// Add this block to the chain, for future proposals
	n.bc.BlocksByHeight[newHeight] = block
	n.bc.BlocksByHash[block.GetBlockHash()] = block
	//Add this Propsal to the Proposal-view map
	n.proposalByviewMap[n.view] = prop
	log.Trace("Finished Proposing")
	// Ship proposal to processing
	relayMsg := &msg.SyncHSMsg{}
	relayMsg.Msg = &msg.SyncHSMsg_Prop{Prop: prop}
	ep := &msg.ExtProposal{}
	ep.FromProto(prop)
	log.Debug("Proposing block:", prop.String())
	//Change itself proposal map
	n.addProposaltoMap()
	go func() {
		// Leader sends new block to all the other nodes
		n.Broadcast(relayMsg)
		// Leader should also vote
		n.voteForBlock(ep)
		// Start 3\delta timer
		n.startBlockTimer(block)
	}()
}

// Deal with the proposal
func (n *SyncHS) proposeHandler(prop *msg.Proposal) {

	ht := prop.Block.GetHeader().GetHeight()
	log.Trace("Handling proposal ", ht)
	ep := &msg.ExtProposal{}
	ep.FromProto(prop)
	if crypto.ToHash(ep.Block.BlockHash) != ep.GetBlockHash() {
		log.Warn("Invalid block. Computed Hash and the Obtained hash does not match")
		return
	}
	data, _ := pb.Marshal(prop.GetBlock().GetHeader())
	correct, err := n.GetPubKeyFromID(n.leader).Verify(data, ep.GetMiningProof())
	if !correct || err != nil {
		log.Error("Incorrect signature for proposal ", ht)
		return
	}
	// Check block certificate for non-genesis blocks
	if !n.IsCertValid(&ep.BlockCertificate) {
		log.Error("Invalid certificate received for block", ht)
		return
	}
	//malicous proposal
	if prop.Miner != n.leader {
		log.Info("There is a malicious propsoal behaviour")
		n.addMaliProposaltoMap(prop)
		n.sendMaliProEvidence(prop)
		//TODO send evidence!
		//We should let current head be the block have certificate although
	} //miabehaviour ocur(empty block)
	var exists bool
	var propOther *msg.Proposal
	{
		// First check for equivocation
		// n.bc.Mu.RLock()
		// blk, exists = n.bc.BlocksByHeight[ht]
		// n.bc.Mu.RUnlock()
		n.propMapLock.RLock()
		propOther, exists = n.proposalByviewMap[n.view]
		n.propMapLock.RUnlock()
	}
	// if exists && ep.GetBlockHash() != blk.GetBlockHash() {
	// 	// Equivocation
	// 	log.Warn("Equivocation detected.", ep.GetBlockHash(), blk.GetBlockHash())
	// 	// TODO trigger view change
	// 	n.addEquiProposaltoMap(prop)
	// 	//TODO send evidence
	// }

	if exists && ep.Proposal.Block.ComputeHash() != propOther.Block.ComputeHash() {
		log.Warn("Equivocation detected.", ep.Proposal.Block.ComputeHash(),
			propOther.Block.ComputeHash())
		n.addEquiProposaltoMap()
		n.sendEqProEvidence(prop, propOther)
		return
	}
	if exists {
		// Duplicate block received,
		// we have already committed this block, IGNORE
		return
	}
	n.addProposaltoMap()
	n.addNewBlock(&ep.ExtBlock)
	n.addProposaltoViewMap(prop)
	n.ensureBlockIsDelivered(&ep.ExtBlock)

	// Vote for the proposal
	go func() {
		n.voteForBlock(ep)
		// Start 3\delta timer
		n.startBlockTimer(&ep.ExtBlock)
	}()

}

// NewBlockBody creates a new block body from the commands received.
func NewBlockBody(cmds [][]byte) *chain.ExtBody {
	bd := &chain.ExtBody{}
	bd.Txs = cmds
	return bd
}

func (n *SyncHS) ensureBlockIsDelivered(blk *chain.ExtBlock) {
	var exists bool
	var parentblk *chain.ExtBlock
	// Ensure that all the parents are delivered first.
	parentIdx := blk.GetHeight() - 1
	// Wait for parents to be delivered first
	for tries := 30; tries > 0; tries-- {
		<-time.After(time.Millisecond)
		n.bc.Mu.RLock()
		parentblk, exists = n.bc.BlocksByHeight[parentIdx]
		n.bc.Mu.RUnlock()
		if exists && parentblk.GetBlockHash() != blk.GetParentHash() {
			// This block is delivered.
			log.Warn("Block  ", blk.GetHeight(), " extending wrong parent.\n",
				"Wanted Parent Block:", util.HashToString(parentblk.GetBlockHash()),
				"Found Parent Block:", util.HashToString(blk.GetParentHash()))
			return
		}
		if exists {
			// The parent of the proposed block is the same as the block we have at the parent's position, CONTINUE
			break
		}
	}
	if !exists {
		// The parents are not delivered, so we cant process this block
		// Return
		log.Warn("Parents not delivered, aborting this proposal")
		return
	}
	// All parents are delivered, lets break out!!
	log.Trace("All parents are delivered")
}

//TODO ，need chain.block  to start timer?

func (n *SyncHS) startBlockTimer(blk *chain.ExtBlock) {
	// Start 3delta timer
	timer := util.NewTimer(func() {
		//check withholding proposal /height = round =view
		_, exists := n.proposalMap[n.GetID()][n.view][n.leader]
		_, exists1 := n.equiproposalMap[n.GetID()][n.view][n.leader]
		if !exists && !exists1 {
			log.Info("withholding block detected")
			//Handle withholding behaviour

			n.handleWithholdingProposal()

			//calculate myself reputation
			n.ReputationCalculateinCurrentRound(n.GetID())
			n.view++
			n.changeLeader()
			if n.leader == n.GetID() {
				n.propose()
			}
			// We have committed this empty block
			log.Info("Committing an withholdemptyblock-", n.view)
			log.Info("The block commit time is", time.Now())

			// Let the client know that we committed this block
			emptyBlockforwh := &chain.ProtoBlock{
				Header: &chain.ProtoHeader{
					Height: n.view,
				},
				BlockHash: chain.EmptyHash.GetBytes(),
			}
			synchsmsg := &msg.SyncHSMsg{}
			ack := &msg.SyncHSMsg_Ack{}
			ack.Ack = &msg.CommitAck{
				Block: emptyBlockforwh,
			}
			synchsmsg.Msg = ack
			// Tell all the clients, that I have committed this block
			n.ClientBroadcast(synchsmsg)
			return
		}
		if exists1 {
			log.Info("Equivocation block detected")
			n.ReputationCalculateinCurrentRound(n.GetID())
			n.view++
			n.changeLeader()
			if n.leader == n.GetID() {
				n.propose()
			}
			log.Info("Committing equivocationblock-", n.view)
			log.Info("The block commit time is", time.Now())
			// We have committed this block
			// Let the client know that we committed this block
			emptyBlockforeq := &chain.ProtoBlock{
				Header: &chain.ProtoHeader{
					Height: n.view,
				},
				BlockHash: chain.EmptyHash.GetBytes(),
			}
			synchsmsg := &msg.SyncHSMsg{}
			ack := &msg.SyncHSMsg_Ack{}
			ack.Ack = &msg.CommitAck{
				Block: emptyBlockforeq,
			}
			synchsmsg.Msg = ack
			// Tell all the clients, that I have committed this block
			n.ClientBroadcast(synchsmsg)
			return
		}
		n.ReputationCalculateinCurrentRound(0)
		n.ReputationCalculateinCurrentRound(1)
		n.ReputationCalculateinCurrentRound(2)
		n.view++
		//TODO ADD LOG for this
		n.changeLeader()
		if n.leader == n.GetID() {
			n.propose()
		}

		log.Info("Committing an correct block-", blk.GetHeight())
		log.Info("The block commit time is", time.Now())
		// We have committed this block
		// Let the client know that we committed this block
		synchsmsg := &msg.SyncHSMsg{}
		ack := &msg.SyncHSMsg_Ack{}
		ack.Ack = &msg.CommitAck{
			Block: blk.ToProto(),
		}
		synchsmsg.Msg = ack
		// Tell all the clients, that I have committed this block
		n.ClientBroadcast(synchsmsg)
		// }
	})
	log.Info("Started timer for block-", blk.GetHeight())
	timer.SetTime(n.GetCommitWaitTime())
	n.addNewTimer(blk.GetHeight(), timer)
	timer.Start()
}

func (n *SyncHS) addNewBlock(blk *chain.ExtBlock) {
	// Otherwise, add the current block to map
	n.bc.Mu.Lock()
	n.bc.BlocksByHeight[blk.GetHeight()] = blk
	n.bc.BlocksByHash[blk.GetBlockHash()] = blk
	n.bc.Mu.Unlock()
}

func (n *SyncHS) addMaliProposaltoMap(prop *msg.Proposal) {
	n.malipropLock.Lock()
	// n.maliproposalMap[n.GetID()] = make(map[uint64]map[uint64]uint64)
	// n.maliproposalMap[n.GetID()][n.view] = make(map[uint64]uint64)
	// mp := n.maliproposalMap[n.GetID()][n.view][prop.Miner]
	// if mp == 0 {
	// 	mp++
	// } else {
	// 	log.Info("Malicious proposal of the leader has been recorded")
	// }
	value, exists := n.maliproposalMap[n.GetID()][n.view][prop.GetMiner()]
	if exists && value == 1 {
		log.Debug("Malicious proposal of this miner in this view has been recorded")
	}
	maliSenderMap := make(map[uint64]uint64)
	maliSenderMap[prop.GetMiner()] = 1
	maliSMapcurrentView := make(map[uint64]map[uint64]uint64)
	maliSMapcurrentView[n.view] = maliSenderMap
	n.maliproposalMap[n.GetID()] = maliSMapcurrentView
	n.malipropLock.Unlock()
}
func (n *SyncHS) addEquiProposaltoMap() {
	n.equipropLock.Lock()
	value, exists := n.equiproposalMap[n.GetID()][n.view][n.leader]
	if exists && value == 1 {
		log.Debug("equivocation propsoal of this leader in this view has been recorded")
	}
	equiSenderMap := make(map[uint64]uint64)
	equiSenderMap[n.leader] = 1
	equiSMapcurrentView := make(map[uint64]map[uint64]uint64)
	equiSMapcurrentView[n.view] = equiSenderMap
	n.equiproposalMap[n.GetID()] = equiSMapcurrentView
	// n.equiproposalMap[n.GetID()] = make(map[uint64]map[uint64]uint64)
	// n.equiproposalMap[n.GetID()][n.view] = make(map[uint64]uint64)
	// e := n.equiproposalMap[n.GetID()][n.view][n.leader]
	// if e == 0 {
	// 	e++
	// } else {
	// 	log.Info("equivocation of the leader has been recorded")
	// }
	n.equipropLock.Unlock()
}

func (n *SyncHS) addWitholdProposaltoMap() {
	n.withpropoLock.Lock()
	// n.withproposalMap[n.GetID()] = make(map[uint64]map[uint64]uint64)
	// n.withproposalMap[n.GetID()][n.view] = make(map[uint64]uint64)
	// w := n.withproposalMap[n.GetID()][n.view][n.leader]
	// if w == 0 {
	// 	w++
	// } else {
	// 	log.Info("this withholding proposal has been recorded")
	// }
	value, exists := n.withproposalMap[n.GetID()][n.view][n.leader]
	if exists && value == 1 {
		log.Debug("withholding propsoal of this leader in this view has been recorded")
	}
	withSenderMap := make(map[uint64]uint64)
	withSenderMap[n.leader] = 1
	withSMapcurrentView := make(map[uint64]map[uint64]uint64)
	withSMapcurrentView[n.view] = withSenderMap
	n.withproposalMap[n.GetID()] = withSMapcurrentView
	n.withpropoLock.Unlock()
}
func (n *SyncHS) addProposaltoMap() {
	n.propMapLock.Lock()
	// n.proposalMap[n.GetID()] = make(map[uint64]map[uint64]uint64)
	// n.proposalMap[n.GetID()][n.view] = make(map[uint64]uint64)
	// p := n.proposalMap[n.GetID()][n.view][n.leader]
	// if p == 0 {
	// 	p++
	// } else {
	// 	log.Info("thpis proposal has been recorded")
	// }
	value, exists := n.proposalMap[n.GetID()][n.view][n.leader]
	if exists && value == 1 {
		log.Debug("propsoal of this leader in this view has been recorded")
	}
	senderMap := make(map[uint64]uint64)
	senderMap[n.leader] = 1
	sMapcurrentView := make(map[uint64]map[uint64]uint64)
	sMapcurrentView[n.view] = senderMap
	n.proposalMap[n.GetID()] = sMapcurrentView
	n.propMapLock.Unlock()
}

func (n *SyncHS) addProposaltoViewMap(prop *msg.Proposal) {
	n.proposalByviewLock.Lock()
	n.proposalByviewMap[n.view] = prop
	n.proposalByviewLock.Unlock()
}

func (n *SyncHS) addNewTimer(pos uint64, timer *util.Timer) {
	n.timerLock.Lock()
	n.timerMaps[pos] = timer
	n.timerLock.Unlock()
}

// NewCandidateProposal returns a proposal message built using commands
func (n *SyncHS) NewCandidateProposal(cmds [][]byte,
	cert *msg.BlockCertificate, newHeight uint64, extra []byte) *msg.Proposal {
	bhash, view := cert.GetBlockInfo()
	// Start setting block fields
	pbody := &chain.ProtoBody{
		Txs:       cmds,
		Responses: cmds, // For now, the response is the same as the cmd
	}
	pheader := &chain.ProtoHeader{
		Extra:      extra,
		Height:     newHeight,
		ParentHash: bhash.GetBytes(),
		TxHash:     nil, // Compute merkle tree out of transactions in the block body
	}
	// Set Hash
	log.Debug("PrevHash:",
		util.HashToString(crypto.ToHash(pheader.GetParentHash())))
	log.Debug("Computed Proposal ", newHeight,
		" with hash ", util.HashToString(bhash))
	// Sign
	data, _ := pb.Marshal(pheader)
	newBlockHash := crypto.DoHash(data)
	sig, err := n.GetMyKey().Sign(data)
	if err != nil {
		log.Error("Error in signing a block during proposal")
		panic(err)
	}
	blk := &chain.ProtoBlock{
		Header:    pheader,
		Body:      pbody,
		BlockHash: newBlockHash.GetBytes(),
	}
	// Build Propose Evidence
	pevidence, _ := pb.Marshal(cert.ToProto())
	prop := &msg.Proposal{
		Miner:           n.GetId(),
		View:            view,
		Block:           blk,
		MiningProof:     sig,       // Signature from the leader in the current view
		ProposeEvidence: pevidence, // Certificate for parent block
	}
	return prop
}
