package consensus

import (
	"math/big"

	"github.com/adithyabhatkajake/libchatter/crypto"
	"github.com/adithyabhatkajake/libchatter/log"
	"github.com/adithyabhatkajake/libsynchs/chain"
	msg "github.com/adithyabhatkajake/libsynchs/msg"
	pb "github.com/golang/protobuf/proto"
)

// Vote channel //TODO malicious vote and equivocation proposal detected!
// TODO{convert the setting of certificate to reputation setting}
func (n *SyncHS) voteHandler() {
	// Map leader to a map from sender to its vote
	voteMap := make(map[uint64]map[uint64]*msg.Vote)
	// pendingVotes := make(map[crypto.Hash][]*msg.Vote)
	isCertified := make(map[uint64]bool)
	myID := n.GetID()
	// Faults := n.GetNumberOfFaultyNodes()
	for {
		v, ok := <-n.voteChannel
		if !ok {
			log.Error("Vote channel error")
			continue
		}
		// if v.Owner != n.leader && n.maliciousVoteInject {
		// 	log.Warn(v.GetVoter(), "'s Malicious vote have been detected.")
		// 	go func() {
		// 		n.addMaliVotetoMap(v)
		// 		n.sendMalivoteEvidence(v)

		// 	}()
		// 	continue
		// }
		bhash := crypto.ToHash(v.GetBlockHash())
		blk := n.getBlock(bhash)

		// if blk == nil {
		// 	// log.Warn(v.GetVoter(), "'s Malicious vote have been detected.")
		// 	// pendingVotes[bhash] = append(pendingVotes[bhash], v)
		// 	// TODO, what to do in this case? malicious vote!
		// 	// go func() {
		// 	// 	n.addMaliVotetoMap(v)
		// 	// 	n.sendMalivoteEvidence(v)

		// 	// }()

		// 	continue
		// }

		//Check if this the first vote for this block height!!

		height := n.view
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
		n.addVotetoMap(v.ToProto())
		//only change votemap for reputation
		isCert, exists := isCertified[height]
		if exists && isCert {
			// This vote for this block is already certified, ignore
			continue
		}
		voteMap[height][v.GetVoter()] = v
		// if uint64(len(voteMap[height])) <= Faults {
		// 	log.Debug("Not enough votes to build a certificate")
		// 	continue
		// }

		//reutation version!!TODO
		var repSumInCert *big.Float = new(big.Float).SetFloat64(0)
		for i := range voteMap[height] {
			repSumInCert = repSumInCert.Add(repSumInCert, n.reputationMap[n.view][i])
		}
		log.Debug("OUTPUT SCORE", repSumInCert)
		if repSumInCert.Cmp(n.GetCertBenchMark(n.view)) == -1 || repSumInCert.Cmp(n.GetCertBenchMark(n.view)) == 0 {
			log.Debug("Not enough reputation to build a certificate")
			continue
		}
		log.Debug("Building a certificate")

		// Our certificate for height v.Data.Block.Data.Index is ready now
		bcert := NewCert(voteMap[height], bhash, view)
		isCertified[height] = true
		// need an anthoer map for non-leader node get exblock
		go func() {
			//add this vote to votemap
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
	log.Info("NODE", n.GetID(), "is voting for", exprop.Miner, "'s block")
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
	n.addVotetoMap(pv)
	go func() {
		//the voter change his voteMap by himself
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
	// n.voterMap[v.GetBody().Voter] = 1
	if _, exists := n.voteMap[n.view]; exists {
		n.voteMap[n.view][v.GetBody().Voter] = 1
	} else {
		n.voteMap[n.view] = make(map[uint64]uint64)
		n.voteMap[n.view][v.GetBody().Voter] = 1
	}

	// log.Debug("vote", n.voteMap)
	n.voteMapLock.Unlock()
}

func (n *SyncHS) GetCertBenchMark(viewNum uint64) *big.Float {
	var totalReputationScore *big.Float = new(big.Float).SetFloat64(0)
	var certBenchMark *big.Float = new(big.Float).SetFloat64(0.5)
	for _, v := range n.reputationMap[viewNum] {
		totalReputationScore = totalReputationScore.Add(totalReputationScore, v)
	}
	certBenchMarkScore := new(big.Float).Mul(totalReputationScore, certBenchMark)
	return certBenchMarkScore
}
