package consensus

import (
	"github.com/adithyabhatkajake/libchatter/log"
	"github.com/adithyabhatkajake/libsynchs/chain"
	msg "github.com/adithyabhatkajake/libsynchs/msg"

	// "github.com/adithyabhatkajake/libsynchs/msg"
	// msg "github.com/adithyabhatkajake/libsynchs/msg"
	pb "github.com/golang/protobuf/proto"
)

//Now, I want to change blame.go to misbehavour handler.
//(except withholding and equivocation,there are malicious block and votes)

func (shs *SyncHS) sendEqProEvidence(prop1 *msg.Proposal, propo2 *msg.Proposal) {
	log.Warn("sending an Equivocation proposal evidence to all nodes")
	eqEvidence := &msg.EquivocationEvidence{}
	eqEvidence.Evidence = &msg.Evidence{}
	eqEvidence.Evidence.EvidenceData = &msg.EvidenceData{}
	eqEvidence.Evidence.EvidenceData.MisbehaviourTarget = shs.leader
	eqEvidence.Evidence.EvidenceData.View = shs.view
	eqEvidence.Evidence.EvOrigin = shs.GetID()
	eqEvidence.E1 = prop1
	eqEvidence.E2 = propo2
	data, err := pb.Marshal(eqEvidence) // signature should include overall content
	if err != nil {
		log.Errorln("Error marshalling eqEvidence", err)
		return
	}
	eqEvidence.Evidence.OrSignature, err = shs.GetMyKey().Sign(data)
	if err != nil {
		log.Errorln("Error Signing the eqEvidence", err)
	}
	eqprEvMsg := &msg.SyncHSMsg{}
	eqprEvMsg.Msg = &msg.SyncHSMsg_Eqevidence{Eqevidence: eqEvidence}
	shs.Broadcast(eqprEvMsg)
	go shs.handleMisbehaviourEvidence(eqprEvMsg)

}

func (shs *SyncHS) sendMaliProEvidence(prop *msg.Proposal) {
	log.Warn("sending an Malicous proposal evidence to all nodes")
	maliproEvidence := &msg.MalicousProposalEvidence{}
	maliproEvidence.Evidence = &msg.Evidence{}
	maliproEvidence.Evidence.EvidenceData = &msg.EvidenceData{}
	maliproEvidence.Evidence.EvidenceData.MisbehaviourTarget = prop.GetMiner()
	maliproEvidence.Evidence.EvidenceData.View = shs.view
	maliproEvidence.Evidence.EvOrigin = shs.GetID()
	maliproEvidence.E = prop
	data, err := pb.Marshal(maliproEvidence)
	if err != nil {
		log.Errorln("Error marshalling maliproEvidence", err)
		return
	}
	maliproEvidence.Evidence.OrSignature, err = shs.GetMyKey().Sign(data)
	if err != nil {
		log.Errorln("Error Signing the maliproEvidence", err)
	}
	maliEvprMsg := &msg.SyncHSMsg{}
	maliEvprMsg.Msg = &msg.SyncHSMsg_Mpevidence{Mpevidence: maliproEvidence}
	shs.Broadcast(maliEvprMsg)
	go shs.handleMisbehaviourEvidence(maliEvprMsg)

}

// msg.vote or proto vote?
func (shs *SyncHS) sendMalivoteEvidence(v *msg.Vote) {
	log.Warn("sending an Malicious vote evidence to all nodes")
	malivoteEvidence := &msg.MalicousVoteEvidence{}
	malivoteEvidence.Evidence = &msg.Evidence{}
	malivoteEvidence.Evidence.EvidenceData = &msg.EvidenceData{}
	malivoteEvidence.Evidence.EvidenceData.MisbehaviourTarget = v.GetVoter()
	malivoteEvidence.Evidence.EvOrigin = shs.GetID()
	malivoteEvidence.E = v.ToProto()
	data, err := pb.Marshal(malivoteEvidence)
	if err != nil {
		log.Errorln("Error marshalling malivoteEvidence", err)
		return
	}
	malivoteEvidence.Evidence.OrSignature, err = shs.GetMyKey().Sign(data)
	if err != nil {
		log.Errorln("Error Signing the malivoteEvidence", err)
	}
	malivoteEvMsg := &msg.SyncHSMsg{}
	malivoteEvMsg.Msg = &msg.SyncHSMsg_Mvevidence{Mvevidence: malivoteEvidence}
	shs.Broadcast(malivoteEvMsg)
	go shs.handleMisbehaviourEvidence(malivoteEvMsg)

}

func (shs *SyncHS) handleMisbehaviourEvidence(ms *msg.SyncHSMsg) {

	switch x := ms.Msg.(type) {
	case *msg.SyncHSMsg_Eqevidence:
		log.Warn("Received a Equicocation proposal evidence!")
		log.Trace("Received a Equivocation proposal evidence against",
			ms.GetEqevidence().Evidence.EvidenceData.MisbehaviourTarget, "from",
			ms.GetEqevidence().Evidence.EvOrigin)
		//check if the evidence is correct
		isValid := shs.isEqpEvidenceValid(ms.GetEqevidence())
		if !isValid {
			log.Debugln("Received an invalid Equivocation proposal evidence message",
				ms.GetEqevidence().String())
			return
		}
		//change corresponding map
		shs.addEquiProposaltoMap(ms.GetEqevidence().E1)

		//we should delete all vote and proposal ocur in this view
		shs.propMapLock.RLock()

		for _, vp := range shs.proposalMap[shs.GetID()][shs.view] {
			vp--

		}
		shs.propMapLock.RUnlock()
		shs.voteMapLock.RLock()
		for _, vv := range shs.voteMap[shs.GetID()][shs.view] {
			vv--
		}
		shs.voteMapLock.RUnlock()
		//how to commit a empty block with certificate
		//and we should change all map need this empty block
		//round = view = height = head

		emptyBlockforEq := &chain.ProtoBlock{}
		emptyBlockforEq.BlockHash = chain.EmptyHash[:]
		emptyBlockforEq.Header.Height = shs.view
		precert, _ := shs.getCertForBlockIndex(shs.view - 1)
		bhash, _ := precert.GetBlockInfo()

		//why we don't use E1/E2 in proposal,because the certificate in proposal maybe invalid
		emptyBlockforEq.Header.ParentHash = bhash.GetBytes()
		exemptyBlockforEq := &chain.ExtBlock{}
		exemptyBlockforEq.FromProto(emptyBlockforEq)
		shs.addNewBlock(exemptyBlockforEq)
		// shs.view++
		// shs.changeLeader() timer end do this

	case *msg.SyncHSMsg_Mpevidence:
		log.Warn("Received a Malicious proposal evidence!")
		log.Trace("Received a Malicious proposal evidence against",
			ms.GetMpevidence().Evidence.EvidenceData.MisbehaviourTarget, "from",
			ms.GetMpevidence().Evidence.EvOrigin)
		isValid := shs.isMalipEvidenceValid(ms.GetMpevidence())
		if !isValid {
			log.Debugln("Received an invalid Malicious proposal evidence message",
				ms.GetMpevidence().String())
		}
		shs.addMaliProposaltoMap(ms.GetProp())
		//continue best-case !

	case *msg.SyncHSMsg_Mvevidence:
		log.Warn("Received a Malicious vote evidence!")
		log.Trace("Received a Malicious vote evidence against",
			ms.GetMpevidence().Evidence.EvidenceData.MisbehaviourTarget, "from",
			ms.GetMpevidence().Evidence.EvOrigin)
		isValid := shs.isMalivEvidenceValid(ms.GetMvevidence())
		if !isValid {
			log.Debugln("Received an invalid Malicious vote evidence message",
				ms.GetMvevidence().String())
		}
		//continue best-case ! //TODO add malicious vote to map!
		shs.addMaliVotetoMap(ms.GetVote())

	case nil:
		log.Warn("Unspecified msg type ", x)
	default:
		log.Warn("Unknown msg type or Unmeet msg type", x)

	}

}

// how to handle WithholdingProposal
func (shs *SyncHS) handleWithholdingProposal() {
	shs.addWitholdProposaltoMap()
	shs.propMapLock.RLock()

	for _, vp := range shs.proposalMap[shs.GetID()][shs.view] {
		vp--

	}
	shs.propMapLock.RUnlock()
	shs.voteMapLock.RLock()
	for _, vv := range shs.voteMap[shs.GetID()][shs.view] {
		vv--
	}
	shs.voteMapLock.RUnlock()
	emptyBlockforwh := &chain.ProtoBlock{}
	emptyBlockforwh.BlockHash = chain.EmptyHash[:]
	emptyBlockforwh.Header.Height = shs.view
	precert, _ := shs.getCertForBlockIndex(shs.view - 1)
	bhash, _ := precert.GetBlockInfo()
	emptyBlockforwh.Header.ParentHash = bhash.GetBytes()
	exemptyBlockforwh := &chain.ExtBlock{}
	exemptyBlockforwh.FromProto(emptyBlockforwh)
	shs.addNewBlock(exemptyBlockforwh)
	shs.view++
	shs.changeLeader()

}

func (shs *SyncHS) isEqpEvidenceValid(eq *msg.EquivocationEvidence) bool {
	log.Traceln("Function isEqpEvidenceValid with input", eq.String())
	// Check if the evidence is for the current leader
	if eq.Evidence.EvidenceData.MisbehaviourTarget != shs.leader {
		log.Debug("Invalid eqpMisbehaviour Target. Found", eq.Evidence.EvidenceData.MisbehaviourTarget,
			",Expected:", shs.leader)
		return false
	}
	// Check if the view is correct!
	if eq.Evidence.EvidenceData.View != shs.view {
		log.Debug("Invalid View. Found", eq.Evidence.EvidenceData.View,
			",Expected:", shs.view)
		return false
	}
	//check the signature of sender
	data, err := pb.Marshal(eq)
	if err != nil {
		log.Debug("Error Marshalling eqEvidence message")
		return false
	}
	isSigValid, err := shs.GetPubKeyFromID(
		eq.Evidence.EvOrigin).Verify(data, eq.Evidence.OrSignature)
	if !isSigValid || err != nil {
		log.Debug("Invalid signature for eqEvidence message")
		return false
	}
	//check the content of the equivocation proposal come from leader
	data1, err := pb.Marshal(eq.E1.Block.Header)
	if err != nil {
		log.Debug("Invalid Marshalling Block.Header1")
		return false
	}
	data2, err := pb.Marshal(eq.E2.Block.Header)
	if err != nil {
		log.Debug("Invalid Marshalling Block.Header2")
		return false
	}
	// ck := eq.E1.Miner == eq.E2.Miner && shs.leader == eq.E2.Miner;
	isSigValidP1, err := shs.GetPubKeyFromID(shs.leader).Verify(data1, eq.E1.MiningProof)
	if err != nil {
		log.Debug("Invalid signature for Block.Header1")
		return false
	}
	isSigValidP2, err := shs.GetPubKeyFromID(shs.leader).Verify(data2, eq.E2.MiningProof)
	if err != nil {
		log.Debug("Invalid signature for Block.Header2")
		return false
	}
	if !isSigValidP1 || !isSigValidP2 {
		log.Debug("Invalid signature on Proposal")
		return false
	}
	return true

}

func (shs *SyncHS) isMalipEvidenceValid(ml *msg.MalicousProposalEvidence) bool {
	log.Traceln("Function isMalipEvidenceValid with input", ml.String())
	//check if the miner is not leader
	if ml.Evidence.EvidenceData.MisbehaviourTarget == shs.leader {
		log.Debug("Invalid malipMisbehaviour Target. Found", ml.Evidence.EvidenceData.MisbehaviourTarget,
			",Expected: other non-leader node")
		return false
	}
	// Check if the view is correct!
	if ml.Evidence.EvidenceData.View != shs.view {
		log.Debug("Invalid View. Found", ml.Evidence.EvidenceData.View,
			",Expected:", shs.view)
		return false
	}
	//check the signature of sender
	data, err := pb.Marshal(ml)
	if err != nil {
		log.Debug("Error Marshalling maliqEvidence message")
		return false
	}
	isSigValid, err := shs.GetPubKeyFromID(
		ml.Evidence.EvOrigin).Verify(data, ml.Evidence.OrSignature)
	if !isSigValid || err != nil {
		log.Debug("Invalid signature for maliqEvidence message")
		return false
	}
	//check the content of the Malicous proposal come from miner
	data1, err := pb.Marshal(ml.E.Block.Header)
	if err != nil {
		log.Debug("Invalid Marshalling Block.Header")
		return false
	}
	isSigValidP, err := shs.GetPubKeyFromID(ml.E.Miner).Verify(data1, ml.E.MiningProof)
	if err != nil || !isSigValidP {
		log.Debug("Invalid signature for Block.Header")
		return false
	}
	return true

}

func (shs *SyncHS) isMalivEvidenceValid(mlv *msg.MalicousVoteEvidence) bool {
	log.Traceln("Function isMalivEvidenceValid with input", mlv.String())
	//check if  voter's object is leader
	if mlv.E.Data.Owner == shs.leader {
		log.Debug("Invalid malivMisbehaviour Evidence ,Expected: other non-leader block")
	}
	// Check if the view is correct!
	if mlv.Evidence.EvidenceData.View != shs.view {
		log.Debug("Invalid View. Found", mlv.Evidence.EvidenceData.View,
			",Expected:", shs.view)
		return false
	}
	//check the signature of sender
	data, err := pb.Marshal(mlv)
	if err != nil {
		log.Debug("Error Marshalling eqEvidence message")
		return false
	}
	isSigValid, err := shs.GetPubKeyFromID(
		mlv.Evidence.EvOrigin).Verify(data, mlv.Evidence.OrSignature)
	if !isSigValid || err != nil {
		log.Debug("Invalid signature for eqEvidence message")
		return false
	}
	//check if the Malicous vote come from miner
	data1, err := pb.Marshal(mlv.E.Data)
	if err != nil {
		log.Debug("Invalid Marshalling ProtoVoteData")
		return false
	}
	isSigValidv, err := shs.GetPubKeyFromID(mlv.E.Body.Voter).Verify(data1, mlv.E.Body.Signature)
	if err != nil || !isSigValidv {
		log.Debug("Invalid vote for ProtoVoteData ")
		return false
	}

	return true

}
