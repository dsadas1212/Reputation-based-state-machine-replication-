package consensus

import (
	"bufio"
	"math/big"
	"time"

	"github.com/adithyabhatkajake/libchatter/log"
	"github.com/adithyabhatkajake/libsynchs/msg"
	"github.com/libp2p/go-libp2p-core/network"
	pb "google.golang.org/protobuf/proto"
)

// Implement how to talk to clients
const (
	ClientProtocolID = "synchs/client/0.0.1"
	Delta            = 4*time.Second + 500*time.Millisecond
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
	//inital reputationMap for all nodes
	n.initialReputationMap()
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
			log.Error("invalid command received from client")
			continue
		}
		// Add command
		// log.Debug("now round is", n.view)
		go n.addCmdsAndStartTimerIfSufficientCommands(cmd)

	}
}

// ClientBroadcast sends a protocol message to all the clients known to this instance
func (n *SyncHS) ClientBroadcast(m *msg.SyncHSMsg) {
	data, err := pb.Marshal(m)
	if err != nil {
		log.Error("Failed to send message", m, "to client")
		panic(err)
		// return
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
	n.timer.SetCallAndCancel(n.callback)
	//+ 150*time.Millisecond
	n.timer.SetTime(Delta)
}

func (n *SyncHS) callback() {

	log.Debug(n.GetID(), "callbackFuncation have been prepared!", time.Now())
	if n.withholdingProposalInject {
		log.Info("In round", n.view, "withholding block have been detected")
		//Handle withholding behaviour
		// n.handleWithholdingProposal()
		n.addNewViewReputaiontoMap()
		synchsmsg := &msg.SyncHSMsg{}
		ack := &msg.SyncHSMsg_Ack{}
		log.Info("Committing an emptyblock in withholding case-", n.view)
		log.Info("The block commit time is", time.Now())
		log.Debug(n.GetID(), "node Blockchain height and view number is", n.bc.Head, "AND", n.view)
		// Let the client know that we committed this block
		ack.Ack = &msg.CommitAck{
			Block: n.bc.BlocksByHeight[n.bc.Head].ToProto(),
		}
		synchsmsg.Msg = ack
		// Tell all the clients, that I have committed this block
		n.ClientBroadcast(synchsmsg)
		n.view++
		n.changeLeader()
		n.SyncChannel <- true
		log.Debug(len(n.SyncChannel), "the next leader is", n.leader)
		return
	}
	if n.equivocatingProposalInject {
		//this equivocation behaviour have been handle
		n.equivocatingProposalInject = false
		log.Info("In round", n.view, "equivocating block have been detected")
		n.addNewViewReputaiontoMap()
		synchsmsg := &msg.SyncHSMsg{}
		ack := &msg.SyncHSMsg_Ack{}
		log.Info("Committing an emptyblock in equivocation case-", n.view)
		log.Info("The block commit time is", time.Now())
		log.Debug(n.GetID(), "node Blockchain height and view number is", n.bc.Head, "AND", n.view)
		// Let the client know that we committed this block
		ack.Ack = &msg.CommitAck{
			Block: n.bc.BlocksByHeight[n.bc.Head].ToProto(),
		}
		synchsmsg.Msg = ack
		// Tell all the clients, that I have committed this block
		n.ClientBroadcast(synchsmsg)
		n.view++
		n.changeLeader()
		n.SyncChannel <- true
		log.Debug(len(n.SyncChannel))
		return

	}
	n.addNewViewReputaiontoMap()
	synchsmsg := &msg.SyncHSMsg{}
	ack := &msg.SyncHSMsg_Ack{}
	//
	log.Debug(n.GetID(), "node Blockchain height and view number is", n.bc.Head, "AND", n.view)
	_, exist := n.getCertForBlockIndex(n.view)
	if !exist {
		log.Debug("fail to generate certificate in round", n.view)
		n.SyncChannel <- true
		return
	}

	log.Info("Committing an correct block-", n.view)
	log.Info("The block commit time of ", n.GetID(), "is", time.Now())
	ack.Ack = &msg.CommitAck{
		Block: n.bc.BlocksByHeight[n.bc.Head].ToProto(),
	}
	synchsmsg.Msg = ack

	// Tell all the clients, that I have committed this block
	n.ClientBroadcast(synchsmsg)
	n.view++
	n.changeLeader()
	n.SyncChannel <- true
	log.Debug(len(n.SyncChannel))

}

func (n *SyncHS) initialReputationMap() {
	n.repMapLock.Lock()
	defer n.repMapLock.Unlock()
	log.Debug(n.pMap)
	for i := uint64(0); i <= uint64(len(n.pMap)); i++ {
		// n.reputationMapwithoutRound[i] = n.initialReplicaSore
		if _, exists := n.reputationMap[n.view]; exists {
			n.reputationMap[n.view][i] = n.initialReplicaSore

		} else {
			n.reputationMap[n.view] = make(map[uint64]*big.Float)
			n.reputationMap[n.view][i] = n.initialReplicaSore
		}
	}

	log.Debug("finish repmap setting with Node", n.GetID(), "'s repmap is", n.reputationMap[n.view])

}
