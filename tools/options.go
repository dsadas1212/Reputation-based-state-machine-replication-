package main

import (
	"os"

	"github.com/pborman/getopt"
)

var (
	nReplicas      uint64  = 10
	nFaulty        uint64  = 4
	blkSize        uint64  = 1
	delta          float64 = 1 // in seconds
	basePort       uint32  = 10000
	clientBasePort uint32  = 20000
	outDir         string  = "testData/"
	defaultIP      string  = "127.0.0.1"
)

var (
	generalConfigFile = "nodes.txt"
	nodeConfigFile    = "nodes-%d.txt"
	clientFile        = "client.txt"
	optNumReplica     = getopt.Uint64('n', nReplicas, "", "Number of Replicas(n)")
	optNumFaulty      = getopt.Uint64('f', nFaulty, "", "Numer of Faulty Replicas(f) [Cannot exceed n-1/2]")
	optBlockSize      = getopt.Uint64('b', blkSize, "", "Number of commands per block(b)")
	optDelay          = getopt.Uint64('d', 10000, "", "Network Delay(d) [in milliseconds]")
	optBasePort       = getopt.Uint32('p', basePort, "", "Base port for repicas. The nodes will use ports starting from this port number.")
	optClientBasePort = getopt.Uint32('c', clientBasePort, "", "Base port for clients. The clients will use these ports to talk to the nodes.")
	help              = getopt.BoolLong("help", 'h', "Show this help message and exit")
	optOutDir         = getopt.String('o', outDir, "Output Directory for the config files")
)

// ParseOptions for generating config files
func ParseOptions() {
	getopt.Parse()
	if *help {
		getopt.Usage()
		os.Exit(0)
	}

	if nReplicas == *optNumReplica {
		nFaulty = *optNumFaulty
	} else {
		nReplicas = *optNumReplica
		tempFaulty := uint64((nReplicas - 1) / 2)
		if *optNumFaulty < tempFaulty {
			nFaulty = *optNumFaulty
		} else {
			nFaulty = tempFaulty
		}
	}
	blkSize = *optBlockSize
	delta = float64(*optDelay) / 1000.0
	basePort = *optBasePort
	clientBasePort = *optClientBasePort
	outDir = *optOutDir
}
