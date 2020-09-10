package consensus

import (
	"bufio"

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
	go n.addClient(rw)
	// Event Handler
	for {
		// Receive a message from a client and process them
		len, err := rw.Read(buf)
		if err != nil {
			log.Error("Error receiving a message from the client-", err)
			go n.removeClient(rw)
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
		go n.addCmdsAndProposeIfSufficientCommands(cmd)
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
