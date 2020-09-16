package main

import (
	"fmt"

	"github.com/adithyabhatkajake/libchatter/io"

	"github.com/adithyabhatkajake/libchatter/crypto/secp256k1"
	"github.com/adithyabhatkajake/libsynchs/config"
	synchsconfig "github.com/adithyabhatkajake/libsynchs/config"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

/* This program performs the followins things: */
/* 1. Generate n public-private key pairs */
/* 2. Print the <replicaID, address, port, public key> in nodes.txt */
/* 3. Print the <Private Key> in nodes-<nodeID>.txt for every node */
func main() {
	ParseOptions()

	// TODO: Pick PKI Algorithm from command line

	// Fetching the context
	alg := secp256k1.Secp256k1Context

	// NodeConfig
	nodeMap := make(map[uint64]*synchsconfig.NodeDataConfig)
	// Address Map for Protocol Nodes
	addrMap := make(map[uint64]*config.Address)
	// Address Map for Clients
	cliMap := make(map[uint64]*config.Address)
	// Public Key Map
	pubKeyMap := make(map[uint64][]byte)

	var err error

	for i := uint64(0); i < nReplicas; i++ {
		// Create a config
		nodeMap[i] = &synchsconfig.NodeDataConfig{}

		// Create Address and set it in the next loop
		addrMap[i] = &config.Address{}
		addrMap[i].IP = defaultIP
		addrMap[i].Port = fmt.Sprintf("%d", basePort+uint32(i))

		cliMap[i] = &config.Address{}
		cliMap[i].IP = defaultIP
		cliMap[i].Port = fmt.Sprintf("%d", clientBasePort+uint32(i))
		nodeMap[i].ClientPort = cliMap[i].Port

		// Generate keypairs
		pvtKey, pubkey := alg.KeyGen()

		nodeMap[i].CryptoCon = &config.CryptoConfig{}
		nodeMap[i].CryptoCon.KeyType = alg.Type()
		nodeMap[i].CryptoCon.PvtKey, err = pvtKey.Raw()
		check(err)

		// Set it in the next loop
		pubKeyMap[i], err = pubkey.Raw()
		check(err)

		// Setup Protocol Configuration
		nodeMap[i].ProtConfig = &synchsconfig.SyncHSConfig{}
		nodeMap[i].ProtConfig.Id = i
		nodeMap[i].ProtConfig.Delta = delta
		nodeMap[i].ProtConfig.Info = &synchsconfig.ProtoInfo{}
		nodeMap[i].ProtConfig.Info.NodeSize = nReplicas
		nodeMap[i].ProtConfig.Info.Faults = nFaulty
		nodeMap[i].ProtConfig.Info.BlockSize = blkSize

	}

	for i := uint64(0); i < nReplicas; i++ {
		nodeMap[i].NetConfig = &config.NetConfig{}
		nodeMap[i].NetConfig.NodeAddressMap = addrMap

		nodeMap[i].CryptoCon.NodeKeyMap = pubKeyMap
	}

	// Write Node Configs
	for i := uint64(0); i < nReplicas; i++ {
		// Open File
		fmt.Println("Processing Node:", i)
		fname := fmt.Sprintf(outDir+nodeConfigFile, i)
		// Serialize NodeConfig and Write to file
		nc := synchsconfig.NewNodeConfig(nodeMap[i])
		io.WriteToFile(nc, fname)
	}

	// Write a config for any client
	clientConfig := &synchsconfig.ClientDataConfig{}

	// Setup cryptographic configurations for the client
	clientConfig.CryptoCon = &config.CryptoConfig{}
	clientConfig.CryptoCon.KeyType = alg.Type()
	pvtKey, _ := alg.KeyGen()
	clientConfig.CryptoCon.PvtKey, err = pvtKey.Raw()
	check(err)
	clientConfig.CryptoCon.NodeKeyMap = pubKeyMap

	// Setup networking configurations for the client
	clientConfig.NetConfig = &config.NetConfig{}
	clientConfig.NetConfig.NodeAddressMap = cliMap

	// Setup Protocol Configurations
	clientConfig.Info = &synchsconfig.ProtoInfo{}
	clientConfig.Info.NodeSize = nReplicas
	clientConfig.Info.BlockSize = blkSize

	fname := fmt.Sprintf(outDir + clientFile)
	cc := synchsconfig.NewClientConfig(clientConfig)
	io.WriteToFile(cc, fname)
}
