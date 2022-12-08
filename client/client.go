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
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/adithyabhatkajake/libchatter/log"
	"github.com/adithyabhatkajake/libsynchs/config"

	"github.com/adithyabhatkajake/libchatter/crypto"
	"github.com/adithyabhatkajake/libchatter/io"
	"github.com/adithyabhatkajake/libsynchs/consensus"
	"github.com/adithyabhatkajake/libsynchs/msg"

	"github.com/libp2p/go-libp2p"
	p2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	pb "google.golang.org/protobuf/proto"
)

const (
	TxInterval = 2*time.Millisecond + 0*time.Microsecond
)

var (
	// BufferCommands defines how many commands to wait for
	// acknowledgement in a batch
	BufferCommands = uint64(10)
	// PendingCommands tells how many commands we are waiting
	// for acknowledgements from replicas
	PendingCommands = uint64(0)
	//cmdMutex         = &sync.Mutex{}
	streamMutex = &sync.Mutex{}
	//voteMutex        = &sync.Mutex{}
	condLock    = &sync.RWMutex{}
	voteChannel chan *msg.CommitAck
	idMap       = make(map[string]uint64)
	//votes            = make(map[crypto.Hash]uint64)
	timeMap          = make(map[crypto.Hash]time.Time)
	commitTimeMetric = make(map[crypto.Hash]time.Duration)
	f                uint64
	metricCount      = uint64(5)
	rwMap            = make(map[uint64]*bufio.Writer)
)

func sendCommandToServer(cmd *msg.SyncHSMsg) {
	log.Trace("Processing Command")
	cmdHash := crypto.DoHash(cmd.GetTx())
	data, err := pb.Marshal(cmd)
	if err != nil {
		log.Error("Marshaling error", err)
		return
	}
	condLock.Lock()
	timeMap[cmdHash] = time.Now()
	condLock.Unlock()
	// Ship command off to the nodes
	streamMutex.Lock()
	for idx, rw := range rwMap {
		rw.Write(data)
		rw.Flush()
		log.Trace("Sending command to node", idx)
	}
	streamMutex.Unlock()
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
		protoMsg := make([]byte, len)
		copy(protoMsg, msgBuf[0:len])
		err = pb.Unmarshal(protoMsg, msg)
		if err != nil {
			log.Error("Unmarshalling error", serverID, err)
			continue
		}
		voteChannel <- msg.GetAck()
	}
}

func handleVotes(cmdChannel chan *msg.SyncHSMsg) {
	voteMap := make(map[crypto.Hash]uint64)
	commitMap := make(map[crypto.Hash]bool)
	for {
		// Get Acknowledgements from nodes after consensus
		ack, ok := <-voteChannel
		timeReceived := time.Now()
		log.Trace("Received an acknowledgement")
		if !ok {
			log.Error("vote channel closed")
			return
		}
		if ack == nil {
			continue
		}
		bhash := crypto.ToHash(ack.GetBlock().GetBlockHash())
		_, exists := voteMap[bhash]
		if !exists {
			voteMap[bhash] = 1       // 1 means we have seen one vote so far.
			commitMap[bhash] = false // To say that we have not yet committed this value
		} else {
			voteMap[bhash]++ // Add another vote
		}
		// To ensure this is executed only once, check old committed state
		old := commitMap[bhash]
		if voteMap[bhash] <= f {
			// Not enough votes for this block
			// So this is not yet committed
			// Deal with it later
			return
		}
		commitMap[bhash] = true
		new := commitMap[bhash]
		//need change the timereceive!!
		log.Trace("Committed block. Processing block",
			ack.GetBlock().GetHeader().GetHeight())
		sendNewCommands := old != new
		txs := ack.GetBlock().GetBody().GetTxs()
		for _, tx := range txs {
			cmdHash := crypto.DoHash(tx)
			condLock.Lock()
			commitTimeMetric[cmdHash] = timeReceived.Sub(timeMap[cmdHash])
			// Time from sending to getting back in some block
			condLock.Unlock()
			// If we commit the block for the first time, then ship off a new command to the server
			if sendNewCommands { // will be triggered once when commitMap value changes
				// 750*time.Microsecond
				<-time.After(TxInterval)
				cmd := <-cmdChannel
				// log.Info("Sending command ", cmd, " to the servers")
				go sendCommandToServer(cmd)
			}
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
		fmt.Printf("Throughput: %f\n", float64(num)/60.0)
		fmt.Printf("Latency: %f\n", float64(count)/float64(num))
		condLock.RUnlock()
	}
	os.Exit(0)
}

func main() {

	confFile := flag.String("conf", "", "Path to client config file")
	batch := flag.Uint64("batch", BufferCommands, "Number of commands to wait for")
	payload := flag.Uint64("payload", 0, "Number of bytes to get as response")
	count := flag.Uint64("metric", metricCount, "Number of metrics to collect before exiting")
	var logLevelPtr = flag.Uint64("loglevel", uint64(log.DebugLevel),
		"Loglevels are one of \n0 - PanicLevel\n1 - FatalLevel\n2 - ErrorLevel\n3 - WarnLevel\n4 - InfoLevel\n5 - DebugLevel\n6 - TraceLevel")
	// Setup Logger
	switch uint32(*logLevelPtr) {
	case 0:
		log.SetLevel(log.PanicLevel)
	case 1:
		log.SetLevel(log.FatalLevel)
	case 2:
		log.SetLevel(log.ErrorLevel)
	case 3:
		log.SetLevel(log.WarnLevel)
	case 4:
		log.SetLevel(log.InfoLevel)
	case 5:
		log.SetLevel(log.DebugLevel)
	case 6:
		log.SetLevel(log.TraceLevel)
	}

	log.Info("I am the client")
	ctx := context.Background()

	flag.Parse()

	// Set values based on command line arguments
	BufferCommands = *batch
	metricCount = *count

	// Get client config
	confData := &config.ClientConfig{}
	io.ReadFromFile(confData, *confFile)

	f = confData.GetFaults()
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

	pMap := make(map[uint64]peer.AddrInfo)
	streamMap := make(map[uint64]network.Stream) //libp2p!!!
	connectedNodes := uint64(0)
	wg := &sync.WaitGroup{}
	updateLock := &sync.Mutex{}

	for i := uint64(0); i < confData.GetNumNodes(); i++ {
		wg.Add(1)
		go func(i uint64, peer peer.AddrInfo) {
			defer wg.Done()
			// Connect to node i
			log.Trace("Attempting connection to node ", peer)
			err := node.Connect(ctx, peer)
			if err != nil {
				log.Error("Connection Error ", err)
				return
			}
			for {
				stream, err := node.NewStream(ctx, peer.ID,
					consensus.ClientProtocolID)
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
				go ackMsgHandler(stream, i)
				break
			}
			log.Debug("Successfully connected to node ", i)
		}(i, confData.GetPeerFromID(i))
	}
	wg.Wait()

	// Ensure we are connected to sufficient nodes
	if connectedNodes <= f {
		log.Warn("Insufficient connections to replicas")
		return
	}

	blksize := confData.GetBlockSize()

	cmdChannel := make(chan *msg.SyncHSMsg, blksize)
	voteChannel = make(chan *msg.CommitAck, blksize)

	// First, spawn a thread that handles acknowledgement received for the
	// various requests
	go handleVotes(cmdChannel)

	idx := uint64(0)
	//CONTROL RATE
	// Then, run a goroutine that sends the first Blocksize requests to the nodes
	for ; idx < blksize; idx++ {
		//+ 750*time.Microsecond
		<-time.After(TxInterval) //for 400
		// Build a command
		cmd := make([]byte, 8+*payload)
		binary.LittleEndian.PutUint64(cmd, idx)

		// Build a protocol message
		cmdMsg := &msg.SyncHSMsg{}
		cmdMsg.Msg = &msg.SyncHSMsg_Tx{
			Tx: cmd,
		}

		// log.Info("Sending command ", idx, " to the servers")
		go sendCommandToServer(cmdMsg)
	}

	go printMetrics()

	// Make sure we always fill the channel with commands
	for {
		// Build a command
		cmd := make([]byte, 8+*payload)

		// Make every command unique so that the hashes are unique
		binary.LittleEndian.PutUint64(cmd, idx)

		// Build a protocol message
		cmdMsg := &msg.SyncHSMsg{}
		cmdMsg.Msg = &msg.SyncHSMsg_Tx{
			Tx: cmd,
		}

		// Dispatch E2C message for processing
		// This will block until some of the commands are committed
		cmdChannel <- cmdMsg
		// Increment command number
		idx++
	}
}
