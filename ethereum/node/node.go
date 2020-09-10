package node

// import (
// 	"net/http"
// 	"os"
// 	"path/filepath"

// 	"github.com/adithyabhatkajake/libsynchs/ethereum/config"
// 	"github.com/ethereum/go-ethereum/rpc"

// 	"github.com/prometheus/tsdb/fileutil"
// )

// type EthNode struct {
// 	dirLock fileutil.Releaser
// 	http    *http.Server
// 	ws      *http.Server
// 	rpc     *rpc.Server
// }

// func NewNode() *EthNode {
// 	e := &EthNode{}
// 	if err := e.CreateDataDir(); err != nil {
// 		panic(err)
// 	}
// 	e.rpc = rpc.NewServer()
// 	e.http = newHTTPServer(node.log, conf.HTTPTimeouts)
// 	e.ws = newHTTPServer(node.log, rpc.DefaultHTTPTimeouts)
// 	e.ipc = newIPCServer(node.log, conf.IPCEndpoint())
// 	return e
// }

// func (e *EthNode) CreateDataDir() error {
// 	instdir := filepath.Join(config.DefaultDataDir(), "")
// 	if err := os.MkdirAll(instdir, 0700); err != nil {
// 		return err
// 	}
// 	// Lock the instance directory to prevent concurrent use by another instance as well as
// 	// accidental use of the instance directory as a database.
// 	release, _, err := fileutil.Flock(filepath.Join(instdir, "LOCK"))
// 	if err != nil {
// 		return err
// 	}
// 	e.dirLock = release
// 	return nil
// }
