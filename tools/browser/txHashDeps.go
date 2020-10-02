package main

import (
	"flag"
	"fmt"

	//"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/params"
)

var (
	txHashStr = flag.String("txHash", "", "Hash of the transaction to find the dependencies tree for")
	datadir   = flag.String("datadir", "", "Data Directory")
)

func main() {

	flag.Parse()
	// Get DB handle-> Get Genesis -> create eth
	// From datadir get DB Handle
	// txHashStr := "0x0600757d20153994fc338bc5fee958ded27dabfe198b7bfd36a9ba949f769a44"
	txHash := common.HexToHash(*txHashStr)

	// datadir := "/data/hermitsage/test-import-generates-state"
	config := &node.Config{
		Name:    "geth",
		Version: params.Version,
		DataDir: *datadir,
		P2P: p2p.Config{
			ListenAddr:  "0.0.0.0:0",
			NoDiscovery: true,
			MaxPeers:    25,
		},
		NoUSB:             true,
		UseLightweightKDF: true,
	}

	stack, err := node.New(config)
	if err != nil {
		fmt.Println("Failed to create a node")
	}

	db, err := stack.OpenDatabase("chaindata", 0, 0, "")
	if err != nil {
		fmt.Println("Failed to open Database", err)
	}
	tx, blkHash, blkNum, loc := rawdb.ReadTransaction(db, txHash)

	fmt.Println("Transaction: ")
	fmt.Println("Recipient: ", (*tx).To().String())
	fmt.Println("Block Hash: ", blkHash.String())
	fmt.Println("Block Number: ", blkNum)
	fmt.Println("Transaction Index: ", loc)

	block := rawdb.ReadBlock(db, blkHash, blkNum)
	stateDb := state.NewDatabase(db)
	if block == nil {
		return
	}
	fmt.Println("Block Root: ", block.Root().String())

	BlkState, err := state.New(blkHash, stateDb, nil)

	fmt.Println("StateDB: ", stateDb)
	fmt.Println("Block State: ", BlkState)
	fmt.Println("Error: ", err)

	blkHash = common.HexToHash("0x4e454b49dc8a2e2a229e0ce911e9fd4d2aa647de4cf6e0df40cf71bff7283330")
	blkNum = 8000000
	block = rawdb.ReadBlock(db, blkHash, blkNum)
	BlkState, err = state.New(block.Root(), stateDb, nil)

	fmt.Println(stateDb, BlkState, err)
}

// 	if err := stack.Register(func(ctx *node.ServiceContext) (node.Service, error) {
// 		return eth.New(ctx, &eth.Config{
// 			Genesis:         nil,
// 			NetworkId:       nil,
// 			SyncMode:        downloader.FullSync,
// 			DatabaseCache:   256,
// 			DatabaseHandles: 256,
// 			TxPool:          core.DefaultTxPoolConfig,
// 			PO:             eth.DefaultConfig.GPO,
// 			Ethash:          eth.DefaultConfig.Ethash,
// 			Miner: Config{
// 				GasFloor: 0,
// 				GasCeil:  genesis.GasLimit * 11 / 10,
// 				GasPrice: big.NewInt(1),
// 				Recommit: time.Second,
// 			},
// 		})
// 	}); err != nil {
// 		return nil, err
// 	}
// 	// Start the node and return if successful
// 	stack.Start()
//   fmt.Println("Successfully Started Stack")
//   // fullNode, err := eth.New(ctx, cfg)
//   // bc, err := core.NewBlockChain(db, cacheConfig, params.AllEthashProtocolChanges , eth.engine, vm.Config{}, nil)
// }
