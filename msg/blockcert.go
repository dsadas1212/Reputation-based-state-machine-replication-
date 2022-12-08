package msg

import (
	"sync/atomic"

	"github.com/adithyabhatkajake/libsynchs/chain"
	"github.com/ethereum/go-ethereum/log"

	"github.com/adithyabhatkajake/libchatter/crypto"
	pb "google.golang.org/protobuf/proto"
)

// BlockCertificate implements PartCert and BlockCert
type BlockCertificate struct {
	PartialCertificate

	// Data
	data ProtoVoteData

	// Cache
	valid atomic.Value
	proto atomic.Value
}

var (
	// GenesisCert is the certificate for the genesis block
	GenesisCert = BlockCertificate{
		data: ProtoVoteData{
			BlockHash: chain.EmptyHash.GetBytes(),
			View:      1,
		},
	}
)

// GetBlockInfo returns the Block Certificate information
// i.e. block hash and view number
func (bc *BlockCertificate) GetBlockInfo() (crypto.Hash, uint64) {
	log.Debug(" blockhash in cert is", bc.data.BlockHash)
	return crypto.ToHash(bc.data.GetBlockHash()), bc.data.GetView()
}

// ToProto converts a block certificate into a protocol buffer for communication
func (bc *BlockCertificate) ToProto() *Certificate {
	val := bc.proto.Load()
	valid := bc.valid.Load()
	isValid := false
	if valid == nil {
		isValid = false
	} else {
		isValid = valid.(bool)
	}
	if val == nil || isValid == false {
		data, _ := pb.Marshal(&bc.data)
		c := &Certificate{
			Data:       data, // Store marshalled ProtoVoteData
			Ids:        bc.ids,
			Signatures: bc.sigs,
		}
		bc.proto.Store(c)
		bc.valid.Store(true)
		return c
	}
	c := val.(*Certificate)
	// Check if the cache is valid
	if len(c.GetIds()) != len(bc.ids) {
		bc.valid.Store(false)
		return bc.ToProto()
	}
	return c
}

// FromProto updates the block certificate from contents of a protocol buffer
func (bc *BlockCertificate) FromProto(data *Certificate) {
	bc.Init()
	pb.Unmarshal(data.GetData(), &bc.data)
	bc.ids = data.GetIds()
	bc.sigs = data.GetSignatures()
	bc.PartialCertificate.data = data.GetData()
	// Update votemap
	for idx, id := range data.GetIds() {
		bc.voteMap[id] = bc.sigs[idx]
	}
	// Update Cache
	bc.proto.Store(data)
}

// AddVote adds the vote<Voter,Signature> to the BlockCertificate
func (bc *BlockCertificate) AddVote(v Vote) {
	bc.PartialCertificate.AddSignature(v.GetVoter(), v.GetSignature())
	// Invalidate the cache
	bc.valid.Store(false)
}

// SetBlockInfo sets the hash and view for the block certificate
func (bc *BlockCertificate) SetBlockInfo(hash crypto.Hash, view uint64) {
	bc.data.BlockHash = hash.GetBytes()
	bc.data.View = view
	// Invalidate the cache
	bc.valid.Store(false)
}
