package consensus

import (
	"bufio"
	"sync"
	"time"

	"github.com/adithyabhatkajake/libchatter/log"
	"github.com/adithyabhatkajake/libsynchs/msg"
	pb "github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p-core/network"
)

// Implement how to talk to clients
const (
	ClientProtocolID = "synchs/client/0.0.1"
)

func (n *SyncHS) addClient(rw *bufio.ReadWriter) {
	// Add new client to cliMap
	n.cliMutex.Lock()
	n.cliMap[rw] = true
	n.cliMutex.Unlock()
}

func (n *SyncHS) removeClient(rw *bufio.ReadWriter) {
	// Remove rw from cliMap after disconnection
	n.cliMutex.Lock()
	delete(n.cliMap, rw)
	n.cliMutex.Unlock()
}

// ClientMsgHandler defines how to talk to client messages
func (n *SyncHS) ClientMsgHandler(s network.Stream) {
	// A buffer to collect messages
	buf := make([]byte, msg.MaxMsgSize)
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	// Add client for later contact
	n.addClient(rw)
	// Set timer for all nodes
	n.setConsensusTimer()
	log.Debug("finish the setting of timer")
	// Event Handler
	for {
		// Receive a message from a client and process them
		len, err := rw.Read(buf)
		if err != nil {
			log.Error("Error receiving a message from the client-", err)
			n.removeClient(rw)
			return
		}
		// Send a copy for reacting
		inMsg := &msg.SyncHSMsg{}
		err = pb.Unmarshal(buf[0:len], inMsg)
		if err != nil {
			log.Error("Error unmarshalling cmd from client")
			log.Error(err)
			continue
		}
		var cmd []byte
		if cmd = inMsg.GetTx(); cmd == nil {
			log.Error("Invalid command received from client")
			continue
		}
		// Add command

		go n.addCmdsAndStartTimerIfSufficientCommands(cmd)

	}
}

// ClientBroadcast sends a protocol message to all the clients known to this instance
func (n *SyncHS) ClientBroadcast(m *msg.SyncHSMsg) {
	data, err := pb.Marshal(m)
	if err != nil {
		log.Error("Failed to send message", m, "to client")
		return
	}
	n.cliMutex.Lock()
	defer n.cliMutex.Unlock()
	for cliBuf := range n.cliMap {
		log.Trace("Sending to", cliBuf)
		cliBuf.Write(data)
		cliBuf.Flush()
	}
	log.Trace("Finish client broadcast for", m)
}

func (n *SyncHS) setConsensusTimer() {
	n.timer0.SetCallAndCancel(n.callback)
	n.timer0.SetTime(25 * time.Second)
	n.timer1.SetCallAndCancel(n.callback)
	n.timer1.SetTime(25 * time.Second)
	n.timer2.SetTime(25 * time.Second)
	n.timer2.SetCallAndCancel(n.callback)
}

func (n *SyncHS) callback() {

	log.Debug(n.GetID(), "callbackFuncation have been prepared!", time.Now())
	// _, exists := n.equiproposalMap[n.GetID()][n.view][n.leader]
	// if n.withholdingProposalInject {

	// 	log.Info("withholding block detected")
	// 	//Handle withholding behaviour
	// 	n.handleWithholdingProposal()

	// 	//calculate myself reputation
	// 	// n.ReputationCalculateinCurrentRound(n.GetID())
	// 	// if n.leader == n.GetID() {
	// 	// 	n.propose()
	// 	// }
	// 	// We have committed this empty block
	// 	go func() {
	// 		log.Info("Committing an withholdemptyblock-", n.view)
	// 		log.Info("The block commit time is", time.Now())

	// 		// Let the client know that we committed this block
	// 		emptyBlockforwh := &chain.ProtoBlock{
	// 			Header: &chain.ProtoHeader{
	// 				Height: n.view,
	// 			},
	// 			BlockHash: chain.EmptyHash.GetBytes(),
	// 		}
	// 		synchsmsg := &msg.SyncHSMsg{}
	// 		ack := &msg.SyncHSMsg_Ack{}
	// 		ack.Ack = &msg.CommitAck{
	// 			Block: emptyBlockforwh,
	// 		}
	// 		synchsmsg.Msg = ack
	// 		// Tell all the clients, that I have committed this block
	// 		n.ClientBroadcast(synchsmsg)

	// 	}()
	// 	n.view++
	// 	n.changeLeader()
	// 	return
	// }
	// if n.equivocatingProposalInject {
	// 	log.Info("Equivocation block detected")
	// 	// if n.leader == n.GetID() {
	// 	// 	n.propose()
	// 	// }
	// 	log.Info("Committing equivocationblock-", n.view)
	// 	log.Info("The block commit time is", time.Now())

	// 	// We have committed this block
	// 	// Let the client know that we committed this block
	// 	go func() {
	// 		emptyBlockforeq := &chain.ProtoBlock{
	// 			Header: &chain.ProtoHeader{
	// 				Height: n.view,
	// 			},
	// 			BlockHash: chain.EmptyHash.GetBytes(),
	// 		}
	// 		synchsmsg := &msg.SyncHSMsg{}
	// 		ack := &msg.SyncHSMsg_Ack{}
	// 		ack.Ack = &msg.CommitAck{
	// 			Block: emptyBlockforeq,
	// 		}
	// 		synchsmsg.Msg = ack
	// 		// Tell all the clients, that I have committed this block
	// 		n.ClientBroadcast(synchsmsg)
	// 	}()
	// 	n.view++
	// 	n.changeLeader()
	// 	return
	// }
	// We have committed this block
	// Let the client know that we committed this block
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		n.ReputationCalculateinCurrentRound(2)
		n.ReputationCalculateinCurrentRound(1)
		n.ReputationCalculateinCurrentRound(0)

	}()
	wg.Wait()
	// go n.ReputationCalculateinCurrentRound(2)
	// go n.ReputationCalculateinCurrentRound(1)
	// go n.ReputationCalculateinCurrentRound(0)
	synchsmsg := &msg.SyncHSMsg{}
	ack := &msg.SyncHSMsg_Ack{}
	_, exist := n.getCertForBlockIndex(n.bc.Head)

	if !exist {
		log.Debug("fail to generate certificate")
		return
	}

	log.Info("Committing an correct block-", n.view)
	log.Info("The block commit time of ", n.GetID(), "is", time.Now())
	ack.Ack = &msg.CommitAck{
		Block: n.proposalByviewMap[n.view].Block,
	}
	synchsmsg.Msg = ack
	// Tell all the clients, that I have committed this block
	n.ClientBroadcast(synchsmsg)
	// }

	log.Debug(n.view)
	if n.view < n.bc.Head {
		n.view++
		n.changeLeader()
		log.Debug(n.leader)
		log.Debug(n.view)
	}
	n.SyncChannel <- true
	log.Debug(len(n.SyncChannel))

}

//TODO ADD LOG for this

// if !n.callFuncFinish && n.callFuncPrepare {
// 	n.callFuncFinish = true
// 	n.callFuncPrepare = false
// }
// log.Debug("funcCallback has been finish!")

// if n.callFuncNotFinish {
// 	n.callFuncNotFinish = false
// }
