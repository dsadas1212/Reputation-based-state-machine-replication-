package msg

import (
	chain "github.com/adithyabhatkajake/libsynchs/chain"
	pb "github.com/golang/protobuf/proto"
)

type ExtProposal struct {
	*Proposal
	chain.ExtBlock
	BlockCertificate
	// ProposalByheight map[int64]*Proposal
}

func (ep *ExtProposal) FromProto(data *Proposal) {
	ep.Proposal = data
	ep.ExtBlock.FromProto(data.Block)
	bc := &Certificate{}
	pb.Unmarshal(data.GetProposeEvidence(), bc)
	ep.BlockCertificate.FromProto(bc)
}

func (ep *ExtProposal) ToProto() *Proposal {
	return ep.Proposal
}
