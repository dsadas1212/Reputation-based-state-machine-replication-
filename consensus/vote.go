package consensus

import (
	"github.com/adithyabhatkajake/libchatter/crypto"
	"github.com/adithyabhatkajake/libchatter/log"
	"github.com/adithyabhatkajake/libsynchs/chain"
	msg "github.com/adithyabhatkajake/libsynchs/msg"
	pb "github.com/golang/protobuf/proto"
)

// Vote channel //TODO malicious vote and equivocation proposal detected!
func (n *SyncHS) voteHandler() {
	// Map leader to a map from sender to its vote
	voteMap := make(map[uint64]map[uint64]*msg.Vote)
	// pendingVotes := make(map[crypto.Hash][]*msg.Vote)
	isCertified := make(map[uint64]bool)
	myID := n.GetID()
	Faults := n.GetNumberOfFaultyNodes()
	for {
		v, ok := <-n.voteChannel
		if !ok {
			log.Error("Vote channel error")
			continue
		}
		bhash := crypto.ToHash(v.GetBlockHash())
		blk := n.getBlock(bhash)

		if blk == nil {
			log.Warn("Malicious vote have been detected.")
			// pendingVotes[bhash] = append(pendingVotes[bhash], v)
			// TODO, what to do in this case? malicious vote!
			go func() {
				n.sendMalivoteEvidence(v)
				n.addMaliVotetoMap(v)
			}()

			continue
		}
		// Check if this the first vote for this block height!!
		height := blk.GetHeight()
		view := v.GetView()
		_, exists := voteMap[height]
		if !exists {
			voteMap[height] = make(map[uint64]*msg.Vote)
			isCertified[height] = false
		}

		_, exists = voteMap[height][v.GetVoter()]
		if exists {
			log.Debug("Duplicate vote received")
			continue
		}

		// My vote is always valid
		if v.GetVoter() != myID {
			isValid := n.isVoteValid(v, blk)
			if !isValid {
				log.Error("Invalid vote message")
				continue
			}
		}
		//add this vote to votemap
		n.addVotetoMap(v.ToProto())
		//only change votemap for reputation
		isCert, exists := isCertified[height]
		if exists && isCert {
			// This vote for this block is already certified, ignore
			continue
		}
		voteMap[height][v.GetVoter()] = v
		if uint64(len(voteMap[height])) <= Faults {
			log.Debug("Not enough votes to build a certificate")
			continue
		}
		log.Debug("Building a certificate")
		// Our certificate for height v.Data.Block.Data.Index is ready now
		bcert := NewCert(voteMap[height], bhash, view)
		isCertified[height] = true
		go func() {
			n.addCert(bcert, height)
			// if n.leader == myID {
			// 	n.propose()
			// }
		}()
	}
}

func (n *SyncHS) isVoteValid(v *msg.Vote, blk *chain.ExtBlock) bool {
	data, err := pb.Marshal(v.ProtoVoteData)
	if err != nil {
		log.Error("Error marshalling vote data")
		log.Error(err)
		return false
	}
	sigOk, err := n.GetPubKeyFromID(v.GetVoter()).Verify(data, v.GetSignature())
	if err != nil {
		log.Error("Vote Signature check error")
		log.Error(err)
		return false
	}
	if !sigOk {
		log.Error("Invalid vote Signature")
		return sigOk
	}
	return sigOk
}

// change it and we can found the miner of the block in vote
func (n *SyncHS) voteForBlock(exprop *msg.ExtProposal) {
	pvd := &msg.ProtoVoteData{
		BlockHash: exprop.ExtBlock.GetBlockHash().GetBytes(),
		View:      n.view,
		Owner:     exprop.Miner,
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
	//the voter change his voteMap by himself
	n.addVotetoMap(pv)
	go func() {
		// Send vote to all the nodes
		n.Broadcast(voteMsg)
		// Handle my own vote
		n.voteChannel <- v
	}()
}
func (n *SyncHS) addMaliVotetoMap(v *msg.Vote) {
	n.voteMaliLock.Lock()
	value, exists := n.voteMaliMap[n.GetID()][n.view][v.ProtoVoteBody.GetVoter()]
	if exists && value == 1 {
		log.Debug("Malicious vote of this voter in this view has been recorded")
	}
	malivoterMap := make(map[uint64]uint64)
	malivoterMap[v.ProtoVoteBody.GetVoter()] = 1
	malivMapcurrentView := make(map[uint64]map[uint64]uint64)
	malivMapcurrentView[n.view] = malivoterMap
	n.voteMaliMap[n.GetID()] = malivMapcurrentView

	// n.voteMaliMap[n.GetID()] = make(map[uint64]map[uint64]uint64)
	// n.voteMaliMap[n.GetID()][n.view] = make(map[uint64]uint64)
	// n.voteMaliMap[n.GetID()][n.view] = map[uint64]uint64{
	// 	v.Body.GetVoter(): 1,
	// }
	// if vomali == 0 {
	// 	vomali++
	// } else {
	// 	log.Debug("Malicious vote of this voter in this view has been recorded")
	// }
	n.voteMaliLock.Unlock()
}

func (n *SyncHS) addVotetoMap(v *msg.ProtoVote) {
	n.voteMapLock.Lock()
	value, exists := n.voteMap[n.GetID()][n.view][v.Body.GetVoter()]
	if exists && value == 1 {
		log.Debug(" vote of this voter in this view has been recorded")
	}
	voterMap := make(map[uint64]uint64)
	voterMap[v.Body.GetVoter()] = 1
	vMapcurrentView := make(map[uint64]map[uint64]uint64)
	vMapcurrentView[n.view] = voterMap
	n.voteMaliMap[n.GetID()] = vMapcurrentView
	n.voteMapLock.Unlock()
}
