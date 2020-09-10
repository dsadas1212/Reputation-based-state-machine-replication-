package consensus

import (
	"github.com/adithyabhatkajake/libchatter/crypto"
	"github.com/adithyabhatkajake/libchatter/log"
	"github.com/adithyabhatkajake/libsynchs/chain"
	msg "github.com/adithyabhatkajake/libsynchs/msg"
	pb "github.com/golang/protobuf/proto"
)

// Vote channel
func (n *SyncHS) voteHandler() {
	// Map leader to a map from sender to its vote
	voteMap := make(map[uint64]map[uint64]*msg.Vote)
	pendingVotes := make(map[crypto.Hash][]*msg.Vote)
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
			log.Warn("Got a vote for a block, which we do not have yet.")
			pendingVotes[bhash] = append(pendingVotes[bhash], v)
			// TODO, what to do in this case?
			continue
		} else {
			// Process pending votes (if any)
		}
		// Check if this the first vote for this block height
		height := blk.GetHeight()
		view := v.GetView()
		_, exists := voteMap[height]
		if !exists {
			voteMap[height] = make(map[uint64]*msg.Vote)
			isCertified[height] = false
		}
		isCert, exists := isCertified[height]
		if exists && isCert {
			// This vote for this block is already certified, ignore
			continue
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
			if n.leader == myID {
				n.propose()
			}
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

func (n *SyncHS) voteForBlock(blk *chain.ExtBlock) {
	pvd := &msg.ProtoVoteData{
		BlockHash: blk.GetBlockHash().GetBytes(),
		View:      n.view,
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
	go func() {
		// Send vote to all the nodes
		n.Broadcast(voteMsg)
		// Handle my own vote
		n.voteChannel <- v
	}()
}
