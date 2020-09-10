package chain

import (
	"sync/atomic"

	"github.com/adithyabhatkajake/libchatter/crypto"
	pb "github.com/golang/protobuf/proto"
)

var (
	EmptyHash = crypto.Hash{}
)

// ExtBlock implements the block interface
// ExtBlock is the block, we will use in our databases and during communication
type ExtBlock struct {
	*ExtHeader
	*ExtBody

	// Cache
	hash  atomic.Value
	proto atomic.Value
}

// ExtHeader is the interface we will use in our databases and during communication
type ExtHeader struct {
	*ProtoHeader
}

// ExtBody is the implementation of the Block Body interface
type ExtBody struct {
	*ProtoBody
}

// ====================================
// Implementing ExtBlock as chain.Block
// ====================================

// GetHeader returns the header for the block
func (eb *ExtBlock) GetHeader() Header {
	return eb.ExtHeader
}

// GetBody returns the body of the block
func (eb *ExtBlock) GetBody() Body {
	return eb.ExtBody
}

// GetSize returns the number of transactions present in the block
func (eb *ExtBlock) GetSize() uint64 {
	return uint64(len(eb.GetTxs()))
}

// ToProto converts ExtBlock into a protocol buffer message
func (eb *ExtBlock) ToProto() *ProtoBlock {
	data := eb.proto.Load()
	proto := data.(*ProtoBlock)
	if proto == nil {
		proto = &ProtoBlock{
			Header:    eb.ProtoHeader,
			Body:      eb.ProtoBody,
			BlockHash: eb.GetBlockHash().GetBytes(),
		}
		eb.proto.Store(proto)
	}
	return proto
}

// FromProto builds an extBlock from the protocol buffer message
func (eb *ExtBlock) FromProto(data *ProtoBlock) {
	eb.ExtHeader = &ExtHeader{
		ProtoHeader: data.Header,
	}
	eb.ExtBody = &ExtBody{
		ProtoBody: data.Body,
	}
	// Update cache
	eb.proto.Store(data)
	eb.hash.Store(crypto.ToHash(data.BlockHash))
}

// GetBlockHash returns the hash of the block this header is referring to
func (eb ExtBlock) GetBlockHash() crypto.Hash {
	var val interface{}
	var hash crypto.Hash
	defer func() {
		eb.hash.Store(hash)
	}()
	if val = eb.hash.Load(); val == nil {
		data, _ := pb.Marshal(eb.ExtHeader.ToProto())
		hash = crypto.ToHash(data)
	} else if hash = val.(crypto.Hash); hash == EmptyHash {
		data, _ := pb.Marshal(eb.ExtHeader.ToProto())
		hash = crypto.ToHash(data)
	}
	return hash
}

func (eb *ExtBlock) IsValid() bool {
	// Check if the hash provided in the block and the hash
	return true
}

// ======================================
// Implementing ExtHeader as chain.Header
// ======================================

// GetExtradata returns extra data from the block
func (eh ExtHeader) GetExtradata() []byte {
	return eh.GetExtra()
}

// GetParentHash returns the hash of the parent block for this block
func (eh ExtHeader) GetParentHash() crypto.Hash {
	return crypto.ToHash(eh.ProtoHeader.GetParentHash())
}

// ToProto converts the struct into a protocol buffer
func (eh *ExtHeader) ToProto() *ProtoHeader {
	return eh.ProtoHeader
}

// FromProto builds a ExtHeader from a protocol buffer header
func (eh *ExtHeader) FromProto(data *ProtoHeader) {
	eh.ProtoHeader = data
}

// GetTxHash returns the TxHash of all the transactions in the body
func (eh *ExtHeader) GetTxHash() crypto.Hash {
	return crypto.ToHash(eh.ProtoHeader.GetTxHash())
}

// ==================================
// Implementing ExtBody as chain.Body
// ==================================

// GetTransactions returns all the transactions in the body
func (eB *ExtBody) GetTransactions() [][]byte {
	return eB.ProtoBody.GetTxs()
	// 	for _, tx := range txs {
	// 		Tx := Transaction(tx)
	// 	}
	// 	return Txs
}

// FromProto builds an ExtBody from ProtoBody
func (eB *ExtBody) FromProto(data *ProtoBody) {
	eB.ProtoBody = data
}
