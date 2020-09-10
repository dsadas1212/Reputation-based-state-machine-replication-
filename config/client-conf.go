package config

import (
	"fmt"

	"github.com/adithyabhatkajake/libchatter/crypto"
	pb "github.com/golang/protobuf/proto"
	peerstore "github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
	jsonpb "google.golang.org/protobuf/encoding/protojson"
)

// ClientConfig is an aggregation of all configs for the client in one place
// We overload the protobuf Client Config Here
type ClientConfig struct {
	// Client Config Protocol Buffer - Read Only
	*ProtoInfo
	*NetConfig
	*CryptoConfig
	// PKI Algorithm
	alg crypto.PKIAlgo
	// My private key
	pvtKey crypto.PrivKey
	// Mapping between nodes and their Public keys
	nodeKeyMap map[uint64]crypto.PubKey
	// cache of client data config
	proto *ClientDataConfig
}

// ToProto converts this into a Protocol Buffer Data Structure
func (cc *ClientConfig) ToProto() *ClientDataConfig {
	if cc.proto == nil {
		cc.proto = &ClientDataConfig{
			Info:      cc.ProtoInfo,
			NetConfig: cc.NetConfig,
			CryptoCon: cc.CryptoConfig,
		}
	}
	return cc.proto
}

// FromProto updates the current ClientConfig from the protocol buffer data
func (cc *ClientConfig) FromProto(data *ClientDataConfig) {
	cc.ProtoInfo = data.Info
	cc.NetConfig = data.NetConfig
	cc.CryptoConfig = data.CryptoCon
}

// MarshalBinary implements Serializable interface
func (cc ClientConfig) MarshalBinary() ([]byte, error) {
	return pb.Marshal(cc.ToProto())
}

// UnmarshalBinary implements Deserializable interface
func (cc *ClientConfig) UnmarshalBinary(inpBytes []byte) error {
	data := &ClientDataConfig{}
	err := pb.Unmarshal(inpBytes, data)
	if err != nil {
		return nil
	}
	cc.FromProto(data)
	cc.init()
	return err
}

// MarshalJSON implements JSON Marshaller interface
// https://talks.golang.org/2015/json.slide#19
func (cc ClientConfig) MarshalJSON() ([]byte, error) {
	return jsonpb.Marshal(cc.ToProto())
}

// UnmarshalJSON implements Unmarshaller JSON interface
// https://talks.golang.org/2015/json.slide#19
func (cc *ClientConfig) UnmarshalJSON(bytes []byte) error {
	err := jsonpb.Unmarshal(bytes, cc.ToProto())
	if err != nil {
		return err
	}
	cc.init()
	return err
}

// init function initializes the structure assuming the config has been set
func (cc *ClientConfig) init() {
	// alg, exists := crypto.GetAlgo(cc.Config.CryptoCon.KeyType)
	alg, exists := crypto.GetAlgo(cc.GetKeyType())
	if exists == false {
		panic("Unknown key type")
	}
	cc.alg = alg
	cc.pvtKey = alg.PrivKeyFromBytes(cc.GetPvtKey())
	cc.nodeKeyMap = make(map[uint64]crypto.PubKey)
	for idx, pubkeyBytes := range cc.GetNodeKeyMap() {
		cc.nodeKeyMap[idx] = alg.PubKeyFromBytes(pubkeyBytes)
	}
}

// NewClientConfig creates a ClientConfig object from ClientDataConfig
func NewClientConfig(con *ClientDataConfig) *ClientConfig {
	cc := &ClientConfig{}
	cc.FromProto(con)
	cc.init()
	return cc
}

// GetPeerFromID returns libp2p peerInfo from the config
func (cc *ClientConfig) GetPeerFromID(nid uint64) peerstore.AddrInfo {
	pID, err := peerstore.IDFromPublicKey(cc.GetPubKeyFromID(nid))
	if err != nil {
		panic(err)
	}
	addr, err := ma.NewMultiaddr(cc.GetP2PAddrFromID(nid))
	if err != nil {
		panic(err)
	}
	pInfo := peerstore.AddrInfo{
		ID:    pID,
		Addrs: []ma.Multiaddr{addr},
	}
	return pInfo
}

// GetNumNodes returns the protocol size
func (cc *ClientConfig) GetNumNodes() uint64 {
	return cc.GetNodeSize()
}

// GetP2PAddrFromID gets the P2P address of the node rid
func (cc *ClientConfig) GetP2PAddrFromID(rid uint64) string {
	address := cc.GetNodeAddressMap()[rid]
	addr := fmt.Sprintf("/ip4/%s/tcp/%s", address.IP, address.Port)
	return addr
}

// GetMyKey returns the private key of this instance
func (cc *ClientConfig) GetMyKey() crypto.PrivKey {
	return cc.pvtKey
}

// GetPubKeyFromID returns the Public key of node whose ID is nid
func (cc *ClientConfig) GetPubKeyFromID(nid uint64) crypto.PubKey {
	return cc.nodeKeyMap[nid]
}
