package msg

import "sync/atomic"

const (
	// MaxMsgSize defines the biggest message to be ever recived in the system
	MaxMsgSize = 10000
)

// PartialCertificate implements PartCert
type PartialCertificate struct {
	data    []byte
	ids     []uint64
	sigs    [][]byte
	voteMap map[uint64][]byte // Duplicate for quick lookups

	// Cache
	proto atomic.Value
	valid atomic.Value // is the cache valid?
}

// ToProto returns the protocol buffer form of the certificate
func (pc PartialCertificate) ToProto() *Certificate {
	val := pc.proto.Load()
	valid := pc.valid.Load()
	isValid := false
	if valid == nil {
		isValid = false
	} else {
		isValid = valid.(bool)
	}
	if val == nil || isValid == false {
		c := &Certificate{
			Data:       pc.data,
			Ids:        pc.ids,
			Signatures: pc.sigs,
		}
		pc.proto.Store(c)
		pc.valid.Store(true)
		return c
	}
	c := val.(*Certificate)
	// We have an invalid cache
	if len(c.GetIds()) != len(pc.ids) {
		pc.valid.Store(false) // Flush
		return pc.ToProto()   // and return latest proto
	}
	return c
}

// FromProto sets the current certificate to data from protocol buffers
func (pc *PartialCertificate) FromProto(data *Certificate) {
	pc.data = data.Data
	pc.ids = data.Ids
	pc.sigs = data.Signatures
}

// GetNumSigners returns the number of signatures we have in the certificate so far
func (pc *PartialCertificate) GetNumSigners() uint64 {
	return uint64(len(pc.ids))
}

// AddSignature adds signer's signature to the certificate
func (pc *PartialCertificate) AddSignature(signer uint64, signature []byte) {
	pc.ids = append(pc.ids, signer)
	pc.sigs = append(pc.sigs, signature)
	pc.voteMap[signer] = signature
	// Invalidate cache of protobuf
	pc.valid.Store(false)
}

// SetData sets the data for which we have the certificate
func (pc *PartialCertificate) SetData(data []byte) {
	pc.data = data
	// Invalidate the cache of protobuf
	pc.valid.Store(false)
}

// GetSignatureFromID returns the signature of node ID
func (pc *PartialCertificate) GetSignatureFromID(id uint64) []byte {
	sig, ok := pc.voteMap[id]
	if !ok {
		return nil
	}
	return sig
}

// GetSigners returns all the signers in this certificate
func (pc *PartialCertificate) GetSigners() []uint64 {
	return pc.ids
}

// GetData returns the data on which we have collected signatures
func (pc *PartialCertificate) GetData() []byte {
	return pc.data
}

// Init initializes the partial certificate
func (pc *PartialCertificate) Init() {
	pc.voteMap = make(map[uint64][]byte)
}
