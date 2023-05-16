package consensus

import (
	"time"

	"github.com/adithyabhatkajake/libchatter/crypto"
	"github.com/adithyabhatkajake/libchatter/log"
	"github.com/adithyabhatkajake/libchatter/util"
	"github.com/adithyabhatkajake/libsynchs/chain"
	msg "github.com/adithyabhatkajake/libsynchs/msg"
	pb "google.golang.org/protobuf/proto"
)

// !!!!!! lock and unlock can be use for the security of thread
// In reputation-based SMR all things begin with Timer!
// ！！version1 use timer
func (n *SyncHS) startConsensusTimer() {

	n.timer.Start()
	log.Debug(n.GetID(), " start a 4Delta timer ", time.Now())
	go func() {
		// if n.GetID()%2 != 0 {
		// 	// n.cmdMutex.Lock()
		// 	// defer n.cmdMutex.Unlock()
		// 	// n.pendingCommands = n.pendingCommands[:uint64(len(n.pendingCommands))-n.GetBlockSize()]
		// 	time.Sleep(time.Second * 120)
		// 	return
		// } else {
		if n.leader == n.GetID() {
			n.Propose()
			//falut number 1/4/8/16/32
		}

		// else {
		// 	//non leader node update its command pool
		// 	n.cmdMutex.Lock()
		// 	defer n.cmdMutex.Unlock()
		// 	n.pendingCommands = n.pendingCommands[:uint64(len(n.pendingCommands))-n.GetBlockSize()]
		// }

		// }

		// }
		//
	}()

}

func (n *SyncHS) Propose() {

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
	// newHeight := n.bc.Head + 1
	log.Info("node", n.GetID(), "is proposing block")
	prop := n.NewCandidateProposal(cmds, cert, newHeight, nil)
	block := &chain.ExtBlock{}
	block.FromProto(prop.Block)
	// Add this block to the chain, for future proposals
	n.bc.BlocksByHeight[newHeight] = block
	n.bc.BlocksByHash[block.GetBlockHash()] = block
	//Add this Propsal to the Proposal-view map
	n.proposalByviewMap[n.view] = prop
	log.Trace("Finished prepare Proposing")
	// Ship proposal to processing
	relayMsg := &msg.SyncHSMsg{}
	relayMsg.Msg = &msg.SyncHSMsg_Prop{Prop: prop}
	//prop.String()
	log.Debug("Proposing block: ", n.GetBlockSize(), "cmd")
	go func() {
		//Change itself proposal map
		n.addProposaltoMap()
		// Leader sends new block to all the other nodes
		n.Broadcast(relayMsg)
	}()

}

// TODO{Deal with the proposal(add the forward step)}
func (n *SyncHS) forward(prop *msg.Proposal) {
	// log.Debug("Node", n.GetID(), "Receive ", prop.GetMiner(), "'s proposal, preparing forward")
	ht := prop.Block.GetHeader().GetHeight()
	log.Debug("Handling leader proposal ", ht)
	ep := &msg.ExtProposal{}
	ep.FromProto(prop)
	if crypto.ToHash(ep.Block.BlockHash) != ep.GetBlockHash() {
		log.Warn("Invalid block. Computed Hash and the Obtained hash does not match")
		return
	}
	data, _ := pb.Marshal(prop.GetBlock().GetHeader())
	correct, err := n.GetPubKeyFromID(n.leader).Verify(data, ep.GetMiningProof())
	if !correct || err != nil {
		log.Error("Forward Incorrect signature for proposal ", ht)
		return
	}
	// Check block certificate for non-genesis blocks
	if !n.IsCertValid(&ep.BlockCertificate) {
		log.Error("Invalid certificate received for block", ht)
		return
	}

	////change propsoal forwardSender and forwardsig
	prop.ForwardSender = n.GetID()
	data1, _ := pb.Marshal(prop.GetBlock().GetHeader())
	sig, err := n.GetMyKey().Sign(data1)
	if err != nil {
		log.Error("Error in signing a block during Forward preparing")
		panic(err)
	}
	prop.ForwardSig = sig
	fRelayMsg := &msg.SyncHSMsg{}
	fRelayMsg.Msg = &msg.SyncHSMsg_Prop{Prop: prop}
	go func() {
		//forward this prospoal
		n.Broadcast(fRelayMsg)
		//hanlde myself forward prospoal
		n.proposeChannel <- prop

	}()

}

// TODO{}
func (n *SyncHS) forwardProposalHandler() {
	fpropMap := make(map[uint64]map[uint64]*msg.Proposal)
	for {
		//check if all forward proposal have been detected
		if len(fpropMap[n.view]) >= len(n.pMap) {
			continue
		}
		fprop, ok := <-n.proposeChannel
		if !ok {
			log.Error("Proposal channel error")
			continue
		}
		//check if equivocation have been &&detected
		if n.equivocatingProposalInject {
			continue
		}
		// log.Debug("NODE", n.GetID(), "Receive forwardSender", fprop.ForwardSender, "'s prospoal")
		ht := fprop.Block.GetHeader().GetHeight()
		log.Trace("Handling forwardSender proposal ", ht)
		ep := &msg.ExtProposal{}
		ep.FromProto(fprop)
		if crypto.ToHash(ep.Block.BlockHash) != ep.GetBlockHash() {
			log.Warn("Invalid block. Computed Hash and the Obtained hash does not match")
			continue
		}
		data, _ := pb.Marshal(fprop.GetBlock().GetHeader())
		correct, err := n.GetPubKeyFromID(n.leader).Verify(data, ep.GetMiningProof())
		if !correct || err != nil {
			log.Error("Forwardhandler Incorrect leader signature for proposal ", ht)
			continue
		}
		// Check block certificate for non-genesis blocks
		if !n.IsCertValid(&ep.BlockCertificate) {
			log.Error("Invalid certificate received for block", ht)
			continue
		}
		// Check forward sender signature
		correctSenderSig, errSig := n.GetPubKeyFromID(ep.ForwardSender).Verify(data, ep.GetForwardSig())
		if !correctSenderSig || errSig != nil {
			log.Error("Incorrect ForwardSender signature for proposal ", ht)
			continue
		}
		//check equivocation prospoal
		_, exists := n.proposalByviewMap[n.view]
		if !exists {
			n.proposalByviewMap[n.view] = fprop

		}
		ep2 := &msg.ExtProposal{}
		ep2.FromProto(n.proposalByviewMap[n.view])
		n.equivocatingProposalInject = ep2.GetBlockHash() != ep.GetBlockHash()
		//Faulty leader don't send his misbehavious
		if n.equivocatingProposalInject {
			// log.Warn("Node", n.GetID(), " detect  Equivocation .", ep2.GetBlockHash(),
			// 	ep.GetBlockHash())
			if n.GetID() != n.leader {
				go n.sendEqProEvidence(n.proposalByviewMap[n.view], fprop)
				continue
			}
			//leader only need to wait for handle equicocation
			continue

		}
		_, exists = fpropMap[n.view]
		if !exists {
			fpropMap[n.view] = make(map[uint64]*msg.Proposal)
		}
		fpropMap[n.view][fprop.ForwardSender] = fprop
		if len(fpropMap[n.view]) < len(n.pMap) {
			log.Debug("NO enough forward prospoal have received")
			continue
		}
		// log.Debug("enough forward prospoal !!")
		//!!!!!!!!!!!
		//set node2 votes for nonexists block
		if n.GetID() != n.leader {
			n.bc.Head++
			n.addProposaltoMap()
			n.addNewBlock(&ep.ExtBlock)
			n.addProposaltoViewMap(fprop)
			n.ensureBlockIsDelivered(&ep.ExtBlock)
			go func() {
				//malicious vote injection!
				if n.GetID()%2 != 0 && n.maliciousVoteInject {
					n.voteForNonLeaderBlk()
					n.maliciousVoteInject = false
				} else {
					// Vote for the forward proposal
					n.voteForBlock(ep)
				}
			}()

		} else {
			// 	//leader only need to vote
			n.voteForBlock(ep)
		}

	}

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
	//genesis block is always ture
	if parentIdx == 0 {
		log.Debug("All parents are delivered")
		return

	}
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
	log.Debug("All parents are delivered")
}

//TODO ，need chain.block  to start timer?

func (n *SyncHS) addNewBlock(blk *chain.ExtBlock) {
	// Otherwise, add the current block to map
	// n.bc.Mu.Lock()
	n.bc.BlocksByHeight[blk.GetHeight()] = blk
	n.bc.BlocksByHash[blk.GetBlockHash()] = blk
	// n.bc.Mu.Unlock()
}

// Note that, there may be many many nodes to do this in same roound, so this case is same with votecase
func (n *SyncHS) addMaliProposaltoMap(prop *msg.Proposal) {
	// n.malipropLock.Lock()
	if _, exists := n.maliproposalMap[n.view]; exists {
		n.maliproposalMap[n.view][prop.Miner] = 1
	} else {
		n.maliproposalMap[n.view] = make(map[uint64]uint64)
		n.maliproposalMap[n.view][prop.Miner] = 1
	}
	// n.malipropLock.Unlock()
	// log.Debug("malipropsoalMAP IN VIEW", n.view, "is", n.maliproposalMap[n.view])
}
func (n *SyncHS) addEquiProposaltoMap() {
	// n.equipropLock.Lock()
	value, exists := n.equiproposalMap[n.view][n.leader]
	if exists && value == 1 {
		log.Debug("equivocation propsoal of this leader in this view has been recorded")
		return
	}
	equiSenderMap := make(map[uint64]uint64)
	equiSenderMap[n.leader] = 1
	n.equiproposalMap[n.view] = equiSenderMap
	// n.equipropLock.Unlock()
}

func (n *SyncHS) addWitholdProposaltoMap() {
	// n.withpropoLock.Lock()
	value, exists := n.withproposalMap[n.view][n.leader]
	if exists && value == 1 {
		log.Debug("withholding propsoal of this leader in this view has been recorded")
		return
	}
	withSenderMap := make(map[uint64]uint64)
	withSenderMap[n.leader] = 1
	n.withproposalMap[n.view] = withSenderMap
	// n.withpropoLock.Unlock()
}
func (n *SyncHS) addProposaltoMap() {
	// n.propMapLock.Lock()
	value, exists := n.proposalMap[n.view][n.leader]
	if exists && value == 1 {
		log.Debug(n.GetID(), " has been recorded the propsoal of this leader in this round")
		return
	}
	senderMap := make(map[uint64]uint64)
	senderMap[n.leader] = 1
	n.proposalMap[n.view] = senderMap
	// n.propMapLock.Unlock()
}

func (n *SyncHS) addProposaltoViewMap(prop *msg.Proposal) {
	// n.proposalByviewLock.Lock()
	n.proposalByviewMap[n.view] = prop
	// n.proposalByviewLock.Unlock()
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
	log.Trace("PrevHash:",
		util.HashToString(crypto.ToHash(pheader.GetParentHash())))
	log.Trace("Computed Proposal ", newHeight,
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
		ForwardSender:   n.leader, //no forward happen, regard leader as defalut value as well as malicious prospoal
		ForwardSig:      nil,
		View:            view,
		Block:           blk,
		MiningProof:     sig,       // Signature from the leader in the current view
		ProposeEvidence: pevidence, // Certificate for parent block
	}
	return prop
}

func (n *SyncHS) createAnEmptyBlock(cmds [][]byte, cert *msg.BlockCertificate, newHeight uint64, extra []byte) *chain.ExtBlock {
	bhash, _ := cert.GetBlockInfo()
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
	blk := &chain.ProtoBlock{
		Header:    pheader,
		Body:      pbody,
		BlockHash: chain.EmptyHash.GetBytes(),
	}
	exblk := &chain.ExtBlock{}
	exblk.FromProto(blk)
	return exblk
}
