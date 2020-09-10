package consensus_test

import (
	"testing"

	"github.com/adithyabhatkajake/libchatter/crypto/secp256k1"
	"github.com/stretchr/testify/require"

	"github.com/adithyabhatkajake/libchatter/crypto"
	"github.com/adithyabhatkajake/libchatter/log"
	"github.com/adithyabhatkajake/libsynchs/chain"
	msg "github.com/adithyabhatkajake/libsynchs/msg"
)

func TestProposal(t *testing.T) {
	sk, _ := secp256k1.Secp256k1Context.KeyGen()
	cmds := make([]*chain.Command, 1)
	cmds[0] = &chain.Command{}
	cmds[0].Cmd = []byte("Test Command")
	prop := &msg.Proposal{}
	prop.Cert = &msg.BlockCertificate{}
	prop.Cert.Init()
	prop.ProposedBlock = &chain.Block{}
	prop.ProposedBlock.Proposer = 1
	prop.ProposedBlock.Data = &chain.BlockData{}
	prop.ProposedBlock.Data.Index = 0 + 1
	prop.ProposedBlock.Data.Cmds = cmds
	// Set previous hash to the current head
	prop.ProposedBlock.Data.PrevHash = crypto.DoHash([]byte("")).GetBytes()
	// Set Hash
	bhash := prop.ProposedBlock.GetHash()
	prop.ProposedBlock.BlockHash = bhash.GetBytes()
	sig, err := prop.ProposedBlock.Sign(sk)
	if err != nil {
		log.Error("Error in signing a block during proposal")
		panic(err)
	}
	prop.ProposedBlock.Signature = sig
	prop.View = 1
	relayMsg := &msg.SyncHSMsg{}
	relayMsg.Msg = &msg.SyncHSMsg_Prop{Prop: prop}

	rprop := relayMsg.GetProp()
	require.Equal(t, true, rprop.ProposedBlock.IsValid())
}
