package consensus

import (
	"bufio"
	"context"
	"math/big"
	"sync"
	"time"

	"github.com/adithyabhatkajake/libchatter/log"
	"github.com/adithyabhatkajake/libsynchs/chain"
	"github.com/adithyabhatkajake/libsynchs/util"

	"github.com/libp2p/go-libp2p"

	"github.com/adithyabhatkajake/libchatter/net"
	config "github.com/adithyabhatkajake/libsynchs/config"

	msg "github.com/adithyabhatkajake/libsynchs/msg"

	"github.com/libp2p/go-libp2p-core/network"
	peerstore "github.com/libp2p/go-libp2p-core/peer"
)

const (
	// ProtocolID is the ID for E2C Protocol
	ProtocolID = "synchs/proto/0.0.1"
	// ProtocolMsgBuffer defines how many protocol messages can be buffered
	ProtocolMsgBuffer = 100
)

// Init implements the Protocol interface
func (shs *SyncHS) Init(c *config.NodeConfig) {
	shs.NodeConfig = c
	shs.leader = DefaultLeaderID
	shs.view = 1 // View Number starts from 1 (convert view to round)
	shs.pendingCommands = make([][]byte, 1000)
	shs.timer = util.Timer{}
	// Setup maps
	shs.streamMap = make(map[uint64]*bufio.ReadWriter) //!!
	shs.cliMap = make(map[*bufio.ReadWriter]bool)      //!!
	// shs.pendingCommands = make(map[crypto.Hash]*chain.Command)
	// shs.timerMaps = make(map[uint64]*util.Timer)
	// shs.blameMap = make(map[uint64]map[uint64]*msg.Blame)
	// shs.certMap = make(map[uint64]*msg.BlockCertificate) // if we should add votemap and proposal map and how to
	shs.reputationMap = make(map[uint64]map[uint64]*big.Float)
	shs.voteMap = make(map[uint64]map[uint64]uint64)
	shs.proposalMap = make(map[uint64]map[uint64]uint64)
	shs.maliproposalMap = make(map[uint64]map[uint64]map[uint64]uint64)
	shs.equiproposalMap = make(map[uint64]map[uint64]map[uint64]uint64)
	shs.withproposalMap = make(map[uint64]map[uint64]map[uint64]uint64)
	shs.voteMaliMap = make(map[uint64]map[uint64]map[uint64]uint64)
	// shs.certBlockMap = make(map[*msg.BlockCertificate]chain.ExtBlock)

	shs.proposalByviewMap = make(map[uint64]*msg.Proposal)

	// Setup channels
	shs.msgChannel = make(chan *msg.SyncHSMsg, ProtocolMsgBuffer)
	shs.cmdChannel = make(chan []byte, ProtocolMsgBuffer)
	shs.voteChannel = make(chan *msg.Vote, ProtocolMsgBuffer)
	// shs.blockCandidateChannel = make(chan *chain.Candidateblock, ProtocolMsgBuffer)
	shs.SyncChannel = make(chan bool, 1)

	shs.certMap = make(map[uint64]*msg.BlockCertificate)
	// Setup certificate for the first block
	shs.certMap[0] = &msg.GenesisCert

	shs.callFuncNotFinish = true
	shs.gcallFuncFinish = true
	shs.maliciousVoteInject = false
	shs.equivocatingProposalInject = false
	shs.withholdingProposalInject = false
	shs.maliciousProposalInject = false
	shs.initialReplicaSore = new(big.Float).SetFloat64(1e-6)

}

// Setup sets up the network components
func (shs *SyncHS) Setup(n *net.Network) error {
	shs.host = n.H
	host, err := libp2p.New(
		context.Background(),
		libp2p.ListenAddrStrings(shs.GetClientListenAddr()),
		libp2p.Identity(shs.GetMyKey()),
	)
	if err != nil {
		panic(err)
	}
	shs.pMap = n.PeerMap
	shs.cliHost = host
	shs.ctx = n.Ctx

	// Obtain a new chain
	shs.bc = chain.NewChain()
	// TODO: create a new chain only if no chain is present in the data directory

	// How to react to Protocol Messages
	shs.host.SetStreamHandler(ProtocolID, shs.ProtoMsgHandler)

	// How to react to Client Messages
	shs.cliHost.SetStreamHandler(ClientProtocolID, shs.ClientMsgHandler)

	// Connect to all the other nodes talking E2C protocol
	wg := &sync.WaitGroup{} // For faster setup
	for idx, p := range shs.pMap {
		wg.Add(1)
		go func(idx uint64, p *peerstore.AddrInfo) {
			log.Trace("Attempting to open a stream with", p, "using protocol", ProtocolID)
			retries := 300
			for i := retries; i > 0; i-- {
				s, err := shs.host.NewStream(shs.ctx, p.ID, ProtocolID)
				if err != nil {
					log.Error("Error connecting to peers:", err)
					log.Info("Retry attempt ", retries-i+1, " to connect to node ", idx, " in a second")
					<-time.After(10 * time.Millisecond)
					continue
				}
				shs.netMutex.Lock()
				shs.streamMap[idx] = bufio.NewReadWriter(
					bufio.NewReader(s), bufio.NewWriter(s))
				shs.netMutex.Unlock()
				log.Info("Connected to Node ", idx)
				break
			}
			wg.Done()
		}(idx, p)
	}
	wg.Wait()
	log.Info("Setup Finished. Ready to do SMR:)", "The begin time is", time.Now())

	return nil
}

// Start implements the Protocol Interface
func (shs *SyncHS) Start() {
	// First, start vote handler concurrently
	go shs.voteHandler()
	// Start E2C Protocol - Start message handler
	shs.protocol()
}

// ProtoMsgHandler reacts to all protocol messages in the network !!
func (shs *SyncHS) ProtoMsgHandler(s network.Stream) {
	// A global buffer to collect messages
	buf := make([]byte, msg.MaxMsgSize)
	// Event Handler
	reader := bufio.NewReader(s)
	for {
		// Receive a message from anyone and process them
		len, err := reader.Read(buf)
		if err != nil {
			return
		}
		// Use a copy of the message and send it to off for processing
		msgBuf := make([]byte, len)
		copy(msgBuf, buf[0:len])
		// React to the message in parallel and continue
		go shs.react(msgBuf)
	}
}
