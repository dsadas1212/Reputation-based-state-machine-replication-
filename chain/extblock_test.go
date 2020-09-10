package chain_test

import (
	"fmt"
	"testing"

	"github.com/adithyabhatkajake/libchatter/crypto"
	"github.com/adithyabhatkajake/libsynchs/chain"
)

var (
	testTxs = make([][]byte, 30)
)

func init() {
	for i := 0; i < 30; i++ {
		testTxs[i] = crypto.DoHash([]byte("hello")).GetBytes()
	}
}
func TestBlockBodyCodec(t *testing.T) {
	pbody := &chain.ProtoBody{
		Txs: testTxs,
	}
	ebody := &chain.ExtBody{}
	ebody.FromProto(pbody)
}

func TestBlockHeaderCodec(t *testing.T) {
	phdr := &chain.ProtoHeader{
		ParentHash: crypto.DoHash([]byte("TestString")).GetBytes(),
		TxHash:     crypto.DoHash([]byte("TestString")).GetBytes(),
		Height:     0,
		Extra:      nil,
	}
	ehdr := &chain.ExtHeader{}
	ehdr.FromProto(phdr)
}

func TestBlockCodec(t *testing.T) {
	phdr := &chain.ProtoHeader{
		ParentHash: crypto.DoHash([]byte("TestString")).GetBytes(),
		TxHash:     crypto.DoHash([]byte("TestString")).GetBytes(),
		Height:     0,
		Extra:      nil,
	}
	ehdr := &chain.ExtHeader{}
	ehdr.FromProto(phdr)
	pbody := &chain.ProtoBody{
		Txs: testTxs,
	}
	ebody := &chain.ExtBody{}
	ebody.FromProto(pbody)
	extb := &chain.ExtBlock{
		ExtHeader: ehdr,
		ExtBody:   ebody,
	}
	fmt.Println(extb.GetBlockHash())
	fmt.Println(extb.GetBlockHash())
	fmt.Println(extb.GetBlockHash())
}
