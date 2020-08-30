package config

import (
	"github.com/adithyabhatkajake/libchatter/crypto"
	pb "github.com/golang/protobuf/proto"
	jsonpb "google.golang.org/protobuf/encoding/protojson"
)

// NodeConfig is an aggregation of all configs in one place
type NodeConfig struct {
	// All Configs
	Config *NodeDataConfig
	// PKI Algorithm
	alg crypto.PKIAlgo
	// My private key
	PvtKey crypto.PrivKey
	// Mapping between nodes and their Public keys
	NodeKeyMap map[uint64]crypto.PubKey
}

// MarshalBinary implements Serializable interface
func (nc NodeConfig) MarshalBinary() ([]byte, error) {
	return pb.Marshal(nc.Config)
}

// UnmarshalBinary implements Deserializable interface
func (nc *NodeConfig) UnmarshalBinary(inpBytes []byte) error {
	data := &NodeDataConfig{}
	err := pb.Unmarshal(inpBytes, data)
	if err != nil {
		return nil
	}
	nc.Config = data
	nc.init()
	return err
}

// MarshalJSON implements JSON Marshaller interface
// https://talks.golang.org/2015/json.slide#19
func (nc NodeConfig) MarshalJSON() ([]byte, error) {
	return jsonpb.Marshal(nc.Config)
}

// UnmarshalJSON implements Unmarshaller JSON interface
// https://talks.golang.org/2015/json.slide#19
func (nc *NodeConfig) UnmarshalJSON(bytes []byte) error {
	err := jsonpb.Unmarshal(bytes, nc.Config)
	if err != nil {
		return err
	}
	nc.init()
	return err
}

// init function initializes the structure assuming the config has been set
func (nc *NodeConfig) init() {
	alg, exists := crypto.GetAlgo(nc.Config.CryptoCon.KeyType)
	if exists == false {
		panic("Unknown key type")
	}
	nc.alg = alg
	nc.PvtKey = alg.PrivKeyFromBytes(nc.Config.CryptoCon.PvtKey)
	nc.NodeKeyMap = make(map[uint64]crypto.PubKey)
	for idx, pubkeyBytes := range nc.Config.CryptoCon.NodeKeyMap {
		nc.NodeKeyMap[idx] = alg.PubKeyFromBytes(pubkeyBytes)
	}
}

// NewNodeConfig creates a NodeConfig object from NodeDataConfig
func NewNodeConfig(con *NodeDataConfig) *NodeConfig {
	nc := &NodeConfig{}
	nc.Config = con
	nc.init()
	return nc
}
