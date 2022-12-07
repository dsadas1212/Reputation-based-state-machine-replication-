package consensus

import (
	"github.com/adithyabhatkajake/libchatter/log"
	msg "github.com/adithyabhatkajake/libsynchs/msg"
	pb "github.com/golang/protobuf/proto"
)

func (n *SyncHS) react(m []byte) {
	log.Trace("Received a message of size", len(m))
	inMessage := &msg.SyncHSMsg{}
	err := pb.Unmarshal(m, inMessage)
	if err != nil {
		log.Error("Received an invalid protocol message from client", err)
		return
	}
	n.msgChannel <- inMessage
}

func (n *SyncHS) protocol() {
	// Process protocol messages
	for {
		msgIn, ok := <-n.msgChannel
		if !ok {
			log.Error("Msg channel error")
			return
		}
		log.Trace("Received msg", msgIn.String())
		switch x := msgIn.Msg.(type) {
		case *msg.SyncHSMsg_Prop:
			prop := msgIn.GetProp()
			if prop.ForwardSender == n.leader {
				if prop.GetMiner() == n.leader {
					log.Debug("Received a proposal from ", prop.GetMiner())
					// Send proposal to forward step
					go n.forward(prop)
				} else {
					log.Debug("Received a Malicious proposal from ", prop.GetMiner(), "in round ", n.view)
					go func() {
						n.maliPropseChannel <- prop
					}()
				}
			} else {
				n.proposeChannel <- prop
			}
			//(start*)
		// case *msg.SyncHSMsg_Eqevidence:
		// 	eqEvidence := msgIn.GetEqevidence()
		// 	go func() {
		// 		n.eqEvidenceChannel <- eqEvidence
		// 	}()

		// case *msg.SyncHSMsg_Mpevidence:
		// 	maliProEvidence := msgIn.GetMpevidence()
		// 	go func() {
		// 		n.maliProEvidenceChannel <- maliProEvidence
		// 	}()

		// case *msg.SyncHSMsg_Mvevidence:
		// 	maliVoteEvidence := msgIn.GetMvevidence()
		// 	go func() {
		// 		n.maliVoteEvidenceChannel <- maliVoteEvidence
		// 	}()
		case *msg.SyncHSMsg_Vote:
			pvote := msgIn.GetVote()
			vote := &msg.Vote{}
			vote.FromProto(pvote)
			go func() {
				n.voteChannel <- vote
			}()
		case nil:
			log.Warn("Unspecified msg type", x)
		default:
			log.Warn("Unknown msg type", x)
		}
	}
}
