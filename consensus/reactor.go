package consensus

import (
	"github.com/adithyabhatkajake/libchatter/log"
	msg "github.com/adithyabhatkajake/libsynchs/msg"
	pb "google.golang.org/protobuf/proto"
)

func (n *SyncHS) react(m []byte) error {
	log.Trace("Received a message of size", len(m))
	inMessage := &msg.SyncHSMsg{}
	// n.netMutex.Lock()
	err := pb.Unmarshal(m, inMessage)
	// n.netMutex.Unlock()
	if err != nil {
		log.Error("Received an invalid protocol message", err)
		return err
	}
	n.msgChannel <- inMessage
	log.Trace("there are no error")
	return nil
}

func (n *SyncHS) protocol() {
	// 处理协议中产生的信息
	for {
		msgIn, ok := <-n.msgChannel
		if !ok {
			log.Error("Msg channel error")
			return
		}
		log.Trace("Received msg", msgIn.String())
		switch x := msgIn.Msg.(type) {
		//区块提议
		case *msg.SyncHSMsg_Prop:
			prop := msgIn.GetProp()
			if prop.ForwardSender == n.leader {
				if prop.GetMiner() == n.leader {
					log.Debug("Received a proposal from ", prop.GetMiner())
					// 将区块提议转发
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
		// (start*)
		//歧义化提议证据
		// case *msg.SyncHSMsg_Eqevidence:
		// 	eqEvidence := msgIn.GetEqevidence()
		// 	go func() {
		// 		n.eqEvidenceChannel <- eqEvidence
		// 	}()
		// 	//恶意区块提议证据
		// case *msg.SyncHSMsg_Mpevidence:
		// 	maliProEvidence := msgIn.GetMpevidence()
		// 	go func() {
		// 		n.maliProEvidenceChannel <- maliProEvidence
		// 	}()
		// 	//恶意投票证据
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
