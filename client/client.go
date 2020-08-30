package main

/*
 * A client does the following:
 * Read the config to get public key and IP maps
 * Let B be the number of commands.
 * Send B commands to the nodes and wait for f+1 acknowledgements for every acknowledgement
 */

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/adithyabhatkajake/libchatter/log"
	synchs "github.com/adithyabhatkajake/libsynchs/config"

	"github.com/adithyabhatkajake/libchatter/crypto"
	e2cio "github.com/adithyabhatkajake/libchatter/io"
	"github.com/adithyabhatkajake/libsynchs/chain"
	"github.com/adithyabhatkajake/libsynchs/consensus"
	msg "github.com/adithyabhatkajake/libsynchs/msg"

	pb "github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p"
	p2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/network"
	peerstore "github.com/libp2p/go-libp2p-core/peer"
)

var (
	// BufferCommands defines how many commands to wait for
	// acknowledgement in a batch
	BufferCommands = uint64(10)
	// PendingCommands tells how many commands we are waiting
	// for acknowledgements from replicas
	PendingCommands  = uint64(0)
	cmdMutex         = &sync.Mutex{}
	streamMutex      = &sync.Mutex{}
	voteMutex        = &sync.Mutex{}
	condLock         = &sync.RWMutex{}
	voteChannel      chan *msg.CommitAck
	idMap            = make(map[string]uint64)
	votes            = make(map[crypto.Hash]uint64)
	timeMap          = make(map[crypto.Hash]time.Time)
	commitTimeMetric = make(map[crypto.Hash]time.Duration)
	f                uint64
	metricCount      = uint64(5)
)

func sendCommandToServer(cmd *msg.SyncHSMsg, rwMap map[uint64]*bufio.Writer) {
	log.Trace("Processing Command")
	cmdHash := cmd.GetCmd().GetHash()
	data, err := pb.Marshal(cmd)
	if err != nil {
		log.Error("Marshaling error", err)
		return
	}
	condLock.Lock()
	timeMap[cmdHash] = time.Now()
	condLock.Unlock()
	// Ship command off to the nodes
	for idx, rw := range rwMap {
		streamMutex.Lock()
		rw.Write(data)
		rw.Flush()
		streamMutex.Unlock()
		log.Trace("Sending command to node", idx)
	}
}

func ackMsgHandler(s network.Stream, serverID uint64) {
	reader := bufio.NewReader(s)
	// Prepare a buffer to receive an acknowledgement
	msgBuf := make([]byte, msg.MaxMsgSize)
	log.Trace("Started acknowledgement message handler")
	for {
		len, err := reader.Read(msgBuf)
		if err != nil {
			log.Error("bufio read error", err)
			return
		}
		log.Trace("Received a message from the server.", serverID)
		msg := &msg.SyncHSMsg{}
		err = pb.Unmarshal(msgBuf[0:len], msg)
		if err != nil {
			log.Error("Unmarshalling error", serverID, err)
			continue
		}
		voteChannel <- msg.GetAck()
	}
}

func handleVotes(cmdChannel chan *msg.SyncHSMsg, rwMap map[uint64]*bufio.Writer) {
	voteMap := make(map[crypto.Hash]uint64)
	commitMap := make(map[crypto.Hash]bool)
	for {
		// Get Acknowledgements from nodes after consensus
		v, ok := <-voteChannel
		log.Trace("Received an acknowledgement")
		if !ok {
			log.Error("vote channel closed")
			return
		}
		if v == nil {
			continue
		}
		// Deal with the vote
		// We received a vote. Now add this to the conformation map
		cmdHash := crypto.ToHash(v.CmdHash)
		_, exists := voteMap[cmdHash]
		if !exists {
			voteMap[cmdHash] = 1       // 1 means we have seen one vote so far.
			commitMap[cmdHash] = false // To say that we have not yet committed this value
		} else {
			voteMap[cmdHash]++ // Add another vote
		}
		// To ensure this is executed only once, check old committed state
		old := commitMap[cmdHash]
		if voteMap[cmdHash] > f {
			condLock.Lock()
			commitTimeMetric[cmdHash] = time.Since(timeMap[cmdHash])
			condLock.Unlock()
			commitMap[cmdHash] = true
		}
		new := commitMap[cmdHash]
		// If we commit the block for the first time, then ship off a new command to the server
		if old != new {
			cmd := <-cmdChannel
			// log.Info("Sending command ", cmd, " to the servers")
			go sendCommandToServer(cmd, rwMap)
		}
	}
}

func printMetrics() {
	printDuration, err := time.ParseDuration("60s")
	if err != nil {
		panic(err)
	}
	var count int64 = 0
	for i := uint64(0); i < metricCount; i++ {
		<-time.After(printDuration)
		condLock.RLock()
		num := 0
		for _, dur := range commitTimeMetric {
			num++
			count += dur.Milliseconds()
		}
		fmt.Println("Metric")
		fmt.Printf("%d cmds in %d milliseconds\n", num, count)
		condLock.RUnlock()
	}
	os.Exit(0)
}

func main() {

	confFile := flag.String("conf", "", "Path to client config file")
	batch := flag.Uint64("batch", BufferCommands, "Number of commands to wait for")
	count := flag.Uint64("metric", metricCount, "Number of metrics to collect before exiting")
	// Setup Logger
	log.SetLevel(log.InfoLevel)

	log.Info("I am the client")
	ctx := context.Background()

	flag.Parse()

	// Set values based on command line arguments
	BufferCommands = *batch
	metricCount = *count

	// Get client config
	confData := &synchs.ClientConfig{}
	e2cio.ReadFromFile(confData, *confFile)

	f = confData.Config.Info.Faults
	// Start networking stack
	node, err := p2p.New(ctx,
		libp2p.Identity(confData.GetMyKey()),
	)
	if err != nil {
		panic(err)
	}

	// Print self information
	log.Info("Client at", node.Addrs())

	// Handle all messages received using ackMsgHandler
	// node.SetStreamHandler(e2cconsensus.ClientProtocolID, ackMsgHandler)
	// Setting stream handler is useless :/

	pMap := make(map[uint64]peerstore.AddrInfo)
	streamMap := make(map[uint64]network.Stream)
	rwMap := make(map[uint64]*bufio.Writer)
	connectedNodes := uint64(0)

	for i := uint64(0); i < confData.GetNumNodes(); i++ {
		// Prepare peerInfo
		pMap[i] = confData.GetPeerFromID(i)
		// Connect to node i
		log.Trace("Attempting connection to node ", pMap[i])
		err = node.Connect(ctx, pMap[i])
		if err != nil {
			log.Error("Connection Error ", err)
			continue
		}
		streamMap[i], err = node.NewStream(ctx, pMap[i].ID,
			consensus.ClientProtocolID)
		if err != nil {
			log.Error("Stream opening Error", err)
			continue
		}
		idMap[streamMap[i].ID()] = i
		connectedNodes++
		rwMap[i] = bufio.NewWriter(streamMap[i])
		go ackMsgHandler(streamMap[i], i)
	}

	// Ensure we are connected to sufficient nodes
	if connectedNodes <= f {
		log.Warn("Insufficient connections to replicas")
		return
	}

	cmdChannel := make(chan *msg.SyncHSMsg, BufferCommands)
	voteChannel = make(chan *msg.CommitAck, BufferCommands)

	// First, spawn a thread that handles acknowledgement received for the
	// various requests
	go handleVotes(cmdChannel, rwMap)

	idx := uint64(0)

	// Then, run a goroutine that sends the first BufferCommands requests to the nodes
	for ; idx < BufferCommands; idx++ {
		// Send via stream a command
		cmdStr := fmt.Sprintf("Do my bidding #%d my servant!", idx)

		// Build a command
		cmd := &chain.Command{}
		// Set command
		cmd.Cmd = []byte(cmdStr)
		// Sign the command
		cmd.Clientsig, err = confData.GetMyKey().Sign(cmd.Cmd)
		if err != nil {
			panic(err)
		}

		// Build a protocol message
		cmdMsg := &msg.SyncHSMsg{}
		cmdMsg.Msg = &msg.SyncHSMsg_Cmd{
			Cmd: cmd,
		}

		// log.Info("Sending command ", idx, " to the servers")
		go sendCommandToServer(cmdMsg, rwMap)
	}

	go printMetrics()

	// Make sure we always fill the channel with commands
	for {
		// Send via stream a command
		cmdStr := fmt.Sprintf("Do my bidding #%d my servant!", idx)

		// Build a command
		cmd := &chain.Command{}
		// Set command
		cmd.Cmd = []byte(cmdStr)
		// Sign the command
		cmd.Clientsig, err = confData.GetMyKey().Sign(cmd.Cmd)
		if err != nil {
			panic(err)
		}

		// Build a protocol message
		cmdMsg := &msg.SyncHSMsg{}
		cmdMsg.Msg = &msg.SyncHSMsg_Cmd{
			Cmd: cmd,
		}

		// Dispatch E2C message for processing
		// This will block until some of the commands are committed
		cmdChannel <- cmdMsg
		// Increment command number
		idx++
	}
}
