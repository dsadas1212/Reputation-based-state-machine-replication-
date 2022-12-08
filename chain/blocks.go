package chain

import (
	"github.com/adithyabhatkajake/libchatter/crypto"
	pb "google.golang.org/protobuf/proto"
)

// ComputeHash computes the hash for the block (i.e. the header)
func (b *ProtoBlock) ComputeHash() crypto.Hash {
	data, _ := pb.Marshal(b)
	return crypto.DoHash(data)
}
