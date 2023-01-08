package consensus

import (
	"bufio"
	"context"
	"encoding/binary"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/adithyabhatkajake/libchatter/io"
	"github.com/adithyabhatkajake/libchatter/log"
	"github.com/adithyabhatkajake/libchatter/net"
	config "github.com/adithyabhatkajake/libsynchs/config"
	msg "github.com/adithyabhatkajake/libsynchs/msg"
	"github.com/libp2p/go-libp2p"
	p2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	pb "google.golang.org/protobuf/proto"
)

var (
	idMap       = make(map[string]uint64)
	rwMap       = make(map[uint64]*bufio.Writer)
	streamMutex = &sync.Mutex{}
)

const (
	TxInterval = 2*time.Millisecond + 0*time.Microsecond
	payload    = uint64(0)
)

// under 2000tx blksize, there is an error which lead reputation-based state machine replication
// can't normal processing
// we need to find what fault cause it
func TestProtoMsgHandler(t *testing.T) {
	//First we mock client send message to nodes, Note that, there are four nodes and a client
	//step 1 : we start a client and one node
	//client
	go startaNodeForConsensus(0)
	go startaNodeForConsensus(1)
	go startaNodeForConsensus(2)
	go startaNodeForConsensus(3)
	time.Sleep(2 * time.Second)
	log.Info("I am the client")
	ctx := context.Background()
	// Get client config
	confData := &config.ClientConfig{}
	io.ReadFromFile(confData, "/root/github.com/Reputation-based-state-machine-replication-/testData/4-node-test2000/client.txt")
	// Start networking stack
	node, err := p2p.New(ctx,
		libp2p.Identity(confData.GetMyKey()),
	)
	if err != nil {
		panic(err)
	}
	// Print self information
	log.Info("Client at", node.Addrs())
	pMap := make(map[uint64]peer.AddrInfo)
	streamMap := make(map[uint64]network.Stream) //libp2p!!!
	connectedNodes := uint64(0)
	wg := &sync.WaitGroup{}
	updateLock := &sync.Mutex{}
	log.Info("node number is ", confData.GetNumNodes())
	for i := uint64(0); i < confData.GetNumNodes(); i++ {
		wg.Add(1)
		go func(i uint64, peer peer.AddrInfo) {
			defer wg.Done()
			// Connect to node i
			log.Info("Attempting connection to node ", peer)
			err := node.Connect(ctx, peer)
			if err != nil {
				log.Error("Connection Error ", err)
				return
			}
			for {
				stream, err := node.NewStream(ctx, peer.ID,
					ClientProtocolID)
				if err != nil {
					log.Trace("Stream opening Error-", err)
					<-time.After(100 * time.Millisecond)
					continue
				}
				updateLock.Lock()
				defer updateLock.Unlock()
				streamMap[i] = stream
				pMap[i] = peer
				idMap[stream.ID()] = i
				connectedNodes++
				rwMap[i] = bufio.NewWriter(stream)
				break
			}
			log.Info("Successfully connected to node ", i)
		}(i, confData.GetPeerFromID(i))
	}
	wg.Wait()

	// step 2 : client send 2000tx blksize block to nodes
	sendCommandToServerByConfigData(confData)

	//node7

	// Then for any node peer1, it processes these txs by ProtoMsgHandler
	// step 1 : check if readbuffer have error

	// step 2 : check if reactor have error
}
func startaNodeForConsensus(num uint64) {
	log.Info("I am the replica" + strconv.FormatUint(num, 10))
	Config := &config.NodeConfig{}
	io.ReadFromFile(Config, "/root/github.com/Reputation-based-state-machine-replication-/testData/4-node-test2000/nodes-"+strconv.FormatUint(num, 10)+".txt")
	netw := net.Setup(Config, Config, Config)
	netw.Connect()
	log.Info("Finished connection to all the nodes")
	// Configure E2C protocol
	shs := SyncHS{}
	shs.Init(Config)
	shs.Setup(netw)
	shs.Start()
	// netw.ShutDown()

}

func sendCommandToServerMock(cmd *msg.SyncHSMsg) {
	log.Trace("Processing Command")
	// cmdHash := crypto.DoHash(cmd.GetTx())
	data, err := pb.Marshal(cmd)
	if err != nil {
		log.Error("Marshaling error", err)
		return
	}
	streamMutex.Lock()
	for idx, rw := range rwMap {
		rw.Write(data)
		rw.Flush()
		log.Trace("Sending command to node", idx)
	}
	streamMutex.Unlock()
}

func sendCommandToServerByConfigData(configdata *config.ClientConfig) {
	idx := uint64(0)
	//CONTROL RATE
	// Then, run a goroutine that sends the first Blocksize requests to the nodes
	blksize := configdata.GetBlockSize()
	for ; idx < blksize; idx++ {
		//+ 750*time.Microsecond
		<-time.After(TxInterval) //for 400
		// Build a command
		cmd := make([]byte, 8+payload)
		binary.LittleEndian.PutUint64(cmd, idx)

		// Build a protocol message
		cmdMsg := &msg.SyncHSMsg{}
		cmdMsg.Msg = &msg.SyncHSMsg_Tx{
			Tx: cmd,
		}
		go sendCommandToServerMock(cmdMsg)
	}
}
