package msg

import (
	"sync/atomic"

	"github.com/adithyabhatkajake/libchatter/crypto"
)

// Vote contains the
// 1. header: VoteData, and
// 2. body: VoteBody
type Vote struct {
	*VoteData
	*VoteBody

	// Cache
	proto atomic.Value
}

// VoteData contains the data on which we have a signature
type VoteData struct {
	*ProtoVoteData
}

// VoteBody contains the Voter and the signature from the voter
type VoteBody struct {
	*ProtoVoteBody
}

// PartCert is a partial certificate consisting of multiple signatures on a fixed data
type PartCert interface {
	GetNumSigners() uint64            // Return the number of signatures contained in the certificate
	AddSignature(uint64, []byte)      // Add signer's signature to the certificate
	SetData([]byte)                   // Set this as the data for signing
	GetData() []byte                  // Returns the Data on which we have signatures
	GetSignatureFromID(uint64) []byte // Return the signature for signer
	GetSigners() []uint64
}

// BlockCert is a special type of PartCert used for block certification
type BlockCert interface {
	AddVote(Vote) // Add this vote to the certificate
	GetNumSigners() uint64
	SetBlockInfo(crypto.Hash, uint64)
	GetBlockInfo() (crypto.Hash, uint64) // Returns the hash of the block, and the view for which we have a certificate
}
