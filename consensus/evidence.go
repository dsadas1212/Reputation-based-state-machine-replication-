package consensus

import (
	"github.com/adithyabhatkajake/libchatter/log"
	msg "github.com/adithyabhatkajake/libsynchs/msg"

	// "github.com/adithyabhatkajake/libsynchs/msg"
	// msg "github.com/adithyabhatkajake/libsynchs/msg"
	pb "github.com/golang/protobuf/proto"
)

//TODO chenge it to handler version!!!

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
	data, err := pb.Marshal(eqEvidence.Evidence.EvidenceData) // signature should include overall content
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

	go func() {
		shs.Broadcast(eqprEvMsg)
		//handle myself evidence
		shs.eqEvidenceChannel <- eqEvidence
	}()
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
	data, err := pb.Marshal(maliproEvidence.Evidence.EvidenceData)
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
	go func() {
		shs.Broadcast(maliEvprMsg)
		//handle myself evidence
		shs.maliProEvidenceChannel <- maliproEvidence
	}()
}

// msg.vote or proto vote?
func (shs *SyncHS) sendMalivoteEvidence(v *msg.Vote) {
	log.Warn("sending an Malicious vote evidence to all nodes")
	malivoteEvidence := &msg.MalicousVoteEvidence{}
	malivoteEvidence.Evidence = &msg.Evidence{}
	malivoteEvidence.Evidence.EvidenceData = &msg.EvidenceData{}
	malivoteEvidence.Evidence.EvidenceData.MisbehaviourTarget = v.ProtoVoteBody.GetVoter()
	malivoteEvidence.Evidence.EvOrigin = shs.GetID()
	malivoteEvidence.E = v.ToProto()
	data, err := pb.Marshal(malivoteEvidence.Evidence.EvidenceData)
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
	go func() {
		shs.Broadcast(malivoteEvMsg)
		//handle myself evidence
		shs.maliVoteEvidenceChannel <- malivoteEvidence
	}()

}

func (shs *SyncHS) EquivocationEvidenceHandler() {
	for {
		eqEvidence, ok := <-shs.eqEvidenceChannel
		if !ok {
			log.Error("Equivocation Evidence channel error")
			continue
		}
		log.Warn("Received a Equicocation proposal evidence!")
		log.Debug("Received a Equivocation proposal evidence against",
			eqEvidence.GetEvidence().GetEvidenceData().GetMisbehaviourTarget(), "from",
			eqEvidence.GetEvidence().GetEvOrigin())
		value, exists1 := shs.equiproposalMap[shs.view][shs.leader]
		if exists1 && value == 1 {
			log.Debug("the equivocation evidence of leader in round", shs.view, "have been recorded!")
			continue
		}
		isValid := shs.isEqpEvidenceValid(eqEvidence)
		if !isValid {
			log.Debugln("Received an invalid Equivocation proposal evidence message")
			// ms.GetEqevidence().String()
			continue
		}
		head := shs.bc.Head
		cert, exists := shs.getCertForBlockIndex(head)
		if !exists {
			log.Debug("The head does not have a certificate, abort handle equivocation evidecne")
			continue
		}

		shs.bc.Head++
		newHeight := shs.bc.Head
		//CREATE non-cmds block and proposal
		blksize := shs.GetBlockSize()
		emptyCmds := make([][]byte, blksize)
		exemptyblock := shs.createAnEmptyBlock(emptyCmds, cert, newHeight, []byte{'E'})
		// Add this block to the chain
		shs.bc.BlocksByHeight[newHeight] = exemptyblock
		shs.bc.BlocksByHash[exemptyblock.GetBlockHash()] = exemptyblock
		//gnerate empty certificate for this block directly
		emptyCertificate := &msg.BlockCertificate{}
		emptyCertificate.SetBlockInfo(exemptyblock.GetBlockHash(), shs.view)
		shs.addCert(emptyCertificate, shs.view)
		//for continue to do evil, misbehaviour leader keep consistency
		shs.addEquiProposaltoMap()

	}
}

func (shs *SyncHS) MaliciousProposalEvidenceHandler() {
	for {
		maliProEvidence, ok := <-shs.maliProEvidenceChannel
		if !ok {
			log.Error("Malicous Propsal Evidence channel error")
			continue
		}
		log.Warn("Received a Malicious proposal evidence!")
		log.Debug("Received a Malicious proposal evidence against",
			maliProEvidence.Evidence.EvidenceData.MisbehaviourTarget, "from",
			maliProEvidence.Evidence.EvOrigin)
		isValid := shs.isMalipEvidenceValid(maliProEvidence)
		if !isValid {
			log.Debugln("Received an invalid Malicious proposal evidence message")
			continue
		}
		maliSenderMap, exists := shs.maliproposalMap[shs.view]
		if exists {
			for i := range maliSenderMap {
				if i == maliProEvidence.Evidence.EvidenceData.MisbehaviourTarget {
					log.Debugln("the malicious prospoal evidence of node",
						maliProEvidence.Evidence.EvidenceData.MisbehaviourTarget,
						"in round", shs.view, "have been recorded!")
					shs.maliProspoalExists = true
					break
				}
			}
			if shs.maliProspoalExists {
				shs.maliProspoalExists = false
				continue
			} else {
				log.Debug("there is anthoner maliciouspropsoal have been recorded")
				shs.addMaliProposaltoMap(maliProEvidence.E)
				shs.maliProspoalExists = false
				continue
			}
		}
		shs.addMaliProposaltoMap(maliProEvidence.E)
	}
}

func (shs *SyncHS) MaliciousVoteEvidenceHandler() {
	for {
		maliVoteEvidence, ok := <-shs.maliVoteEvidenceChannel
		if !ok {
			log.Error("Malicous vote Evidence channel error")
			continue
		}
		log.Warn("Received a Malicious vote evidence!")
		log.Debug("Received a Malicious vote evidence against",
			maliVoteEvidence.Evidence.EvidenceData.MisbehaviourTarget, "from",
			maliVoteEvidence.Evidence.EvOrigin)
		isValid := shs.isMalivEvidenceValid(maliVoteEvidence)
		if !isValid {
			log.Debugln("Received an invalid Malicious vote evidence message")
			continue
		}
		shs.voteMaliLock.RLock()
		maliVoterMap, exists := shs.voteMaliMap[shs.view]
		if exists {
			for i := range maliVoterMap {
				if i == maliVoteEvidence.Evidence.EvidenceData.MisbehaviourTarget {
					log.Debugln("the malicious vote evidence of node",
						maliVoteEvidence.Evidence.EvidenceData.MisbehaviourTarget,
						"in round", shs.view, "have been recorded!")
					shs.maliVoteExists = true
					break
				}
			}
			if shs.maliVoteExists {
				shs.maliVoteExists = false
				continue
			} else {
				shs.addMaliVotetoMap(maliVoteEvidence.E)
				shs.maliVoteExists = false
				continue
			}
		}
		shs.voteMaliLock.RUnlock()
		shs.addMaliVotetoMap(maliVoteEvidence.E)
	}
}

// how to handle WithholdingProposal,the evidence of it need not boradcast.
func (shs *SyncHS) handleWithholdingProposal() {
	shs.bc.Mu.Lock()
	defer shs.bc.Mu.Unlock()
	head := shs.bc.Head
	cert, exists := shs.getCertForBlockIndex(head)
	if !exists {
		log.Debug("The head does not have a certificate, abort handle withholding evidence")
		return
	}
	shs.bc.Head++
	newHeight := shs.bc.Head
	//CREATE non-cmds block and proposal
	blksize := shs.GetBlockSize()
	emptyCmds := make([][]byte, blksize)
	// withholdProp := shs.NewCandidateProposal(emptyCmds, cert, newHeight, []byte{'E'})
	exemptyblock := shs.createAnEmptyBlock(emptyCmds, cert, newHeight, []byte{'E'})
	// Add this block to the chain
	shs.bc.BlocksByHeight[newHeight] = exemptyblock
	shs.bc.BlocksByHash[exemptyblock.GetBlockHash()] = exemptyblock
	//gnerate empty certificate for this block directly
	emptyCertificate := &msg.BlockCertificate{}
	emptyCertificate.SetBlockInfo(exemptyblock.GetBlockHash(), shs.view)
	shs.addCert(emptyCertificate, shs.view)
	//For consistency change itself withhold map
	shs.addWitholdProposaltoMap()
}

func (shs *SyncHS) isEqpEvidenceValid(eq *msg.EquivocationEvidence) bool {
	log.Traceln("Function isEqpEvidenceValid with input", eq.String())
	// Check if the evidence against for the current leader
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
	data, err := pb.Marshal(eq.Evidence.EvidenceData)
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
	data1, err := pb.Marshal(eq.E1.GetBlock().GetHeader())
	if err != nil {
		log.Debug("Invalid Marshalling Block.Header1")
		return false
	}
	data2, err := pb.Marshal(eq.E2.GetBlock().GetHeader())
	if err != nil {
		log.Debug("Invalid Marshalling Block.Header2")
		return false
	}
	// ck := eq.E1.Miner == eq.E2.Miner && shs.leader == eq.E2.Miner;
	isSigValidP1, err := shs.GetPubKeyFromID(shs.leader).Verify(data1, eq.E1.GetMiningProof())
	if !isSigValidP1 || err != nil {
		log.Debug("Invalid signature for Block.Header1,err is", err)
		return false
	}
	isSigValidP2, err := shs.GetPubKeyFromID(shs.leader).Verify(data2, eq.E2.GetMiningProof())

	if !isSigValidP2 || err != nil {
		log.Debug("Invalid signature for Block.Header2")
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
	//Check if the view is correct!
	if ml.Evidence.EvidenceData.View != shs.view {
		log.Debug("malipropsoal Invalid View. Found", ml.Evidence.EvidenceData.View,
			",Expected:", shs.view)
		return false
	}
	//check the signature of sender
	data, err := pb.Marshal(ml.Evidence.EvidenceData)
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
	data, err := pb.Marshal(mlv.Evidence.EvidenceData)
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
