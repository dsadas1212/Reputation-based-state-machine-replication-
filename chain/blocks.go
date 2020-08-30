package chain

import (
	"bytes"

	"github.com/adithyabhatkajake/libchatter/log"
	"github.com/adithyabhatkajake/libchatter/util"

	"github.com/adithyabhatkajake/libchatter/crypto"
	pb "github.com/golang/protobuf/proto"
)

// GetHash computes the hash from the block data
func (b *Block) GetHash() crypto.Hash {
	data, err := pb.Marshal(b.Data)
	if err != nil {
		panic(err)
	}
	return crypto.DoHash(data)
}

// GetHashBytes returns crypto.Hash from b.BlockHash
func (b *Block) GetHashBytes() crypto.Hash {
	var x crypto.Hash
	copy(x[:], b.BlockHash)
	return x
}

// IsValid checks if the block is valid
func (b *Block) IsValid() bool {
	bhash := b.GetHash()
	// Check if the hash is correctly computed
	if !bytes.Equal(bhash.GetBytes(), b.BlockHash) {
		log.Debug("Have Hash:", util.HashToString(b.GetHashBytes()))
		log.Debug("Computed Hash:", util.HashToString(bhash))
		return false
	}
	return true
}

// Sign computes and sets the signature for the block or returns an error
func (b *Block) Sign(sk crypto.PrivKey) ([]byte, error) {
	data, err := pb.Marshal(b.Data)
	log.Debug("Marshalled Block while signing")
	log.Debug(util.BytesToHexString(data))
	if err != nil {
		log.Error("Error in marshalling block data during proposal")
		panic(err)
	}
	return sk.Sign(data)
}
