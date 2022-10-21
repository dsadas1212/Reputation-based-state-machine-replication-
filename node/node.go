package main

import (
	"flag"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"

	"github.com/adithyabhatkajake/libchatter/io"
	"github.com/adithyabhatkajake/libchatter/log"
	"github.com/adithyabhatkajake/libchatter/net"
	config "github.com/adithyabhatkajake/libsynchs/config"
	synchs "github.com/adithyabhatkajake/libsynchs/consensus"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")
var logLevelPtr = flag.Uint64("loglevel", uint64(log.DebugLevel),
	"Loglevels are one of \n0 - PanicLevel\n1 - FatalLevel\n2 - ErrorLevel\n3 - WarnLevel\n4 - InfoLevel\n5 - DebugLevel\n6 - TraceLevel")
var configFileStrPtr = flag.String("conf", "", "Path to config file")
var numCPU = flag.Int("cpu", 2, "number of cpu cores to use for the node")

func main() {
	// Parse flags
	flag.Parse()

	// Set number of cores
	runtime.GOMAXPROCS(*numCPU)

	logLevel := log.InfoLevel

	switch uint32(*logLevelPtr) {
	case 0:
		logLevel = log.PanicLevel
	case 1:
		logLevel = log.FatalLevel
	case 2:
		logLevel = log.ErrorLevel
	case 3:
		logLevel = log.WarnLevel
	case 4:
		logLevel = log.InfoLevel
	case 5:
		logLevel = log.DebugLevel
	case 6:
		logLevel = log.TraceLevel
	}

	// Log Settings
	log.SetLevel(logLevel)

	if *cpuprofile != "" {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		go func() {
			_ = <-sigs
			pprof.StopCPUProfile()
			f.Close() // error handling omitted for example
			os.Exit(0)
		}()
	}

	log.Info("I am the replica.")
	Config := &config.NodeConfig{}

	io.ReadFromFile(Config, *configFileStrPtr)
	log.Debug("Finished reading the config file", os.Args[1])

	// Setup connections
	netw := net.Setup(Config, Config, Config)

	// Connect and send a test message
	netw.Connect()
	log.Debug("Finished connection to all the nodes")

	// Configure E2C protocol
	shs := &synchs.SyncHS{}
	shs.Init(Config)
	shs.Setup(netw)

	// Start E2C
	shs.Start()

	// Disconnect
	netw.ShutDown()
}
