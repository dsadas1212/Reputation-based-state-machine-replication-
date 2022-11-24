package consensus

import (
	"math/big"

	"github.com/adithyabhatkajake/libchatter/crypto"
	"github.com/adithyabhatkajake/libchatter/log"
	"github.com/adithyabhatkajake/libsynchs/chain"
	msg "github.com/adithyabhatkajake/libsynchs/msg"
)

// How to create and validate certificates(we need convert it to reputation-based)

// NewCert creates a certificate
func NewCert(certMap map[uint64]*msg.Vote, blockhash crypto.Hash, view uint64) *msg.BlockCertificate {
	bc := &msg.BlockCertificate{}
	bc.SetBlockInfo(blockhash, view)
	bc.Init()
	for _, v := range certMap {
		bc.AddVote(*v)
	}
	return bc
}

// IsCertValid checks if the certificate is valid for the data
func (n *SyncHS) IsCertValid(bc *msg.BlockCertificate) bool {
	// log.Debug("Received a block certificate -")
	h, _ := bc.GetBlockInfo()
	// Certificate for genesis is always correct
	if h == chain.EmptyHash {
		return true
	}
	exEmptyBlk := n.bc.BlocksByHash[h]
	ex := exEmptyBlk.ExtHeader.GetExtra()

	//Certificate for emptycmdblock is always correct
	if len(ex) == 1 {
		return true
	}
	// if bc.GetNumSigners() <= n.GetNumberOfFaultyNodes() {
	// 	log.Error("The certificate has <= f signatures")
	// 	return false
	// }
	benchmark := n.GetCertBenchMark(n.view - 1)
	totalRepInCert := new(big.Float).SetFloat64(0)
	for _, id := range bc.GetSigners() {
		sig := bc.GetSignatureFromID(id)
		if sig == nil {
			log.Error("Signature for ID not found")
			return false
		}
		_, err := n.GetPubKeyFromID(id).Verify(bc.GetData(), sig)
		if err != nil {
			log.Error("Certificate signature verification error")
			return false
		}
		// if !sigOk {
		// 	log.Error("Certificate signature is invalid for idx", idx)
		// 	return false
		// }
		totalRepInCert = totalRepInCert.Add(totalRepInCert, n.reputationMap[n.view-1][id])
	}
	if totalRepInCert.Cmp(benchmark) == -1 || totalRepInCert.Cmp(benchmark) == 0 {
		log.Error("invalid cert because lacking reputation")
		return false
	}
	return true
}

func (n *SyncHS) addCert(bc *msg.BlockCertificate, blockNum uint64) {
	log.Debug("Adding certificate to block ", blockNum)
	n.certMapLock.Lock()
	n.certMap[blockNum] = bc
	n.certMapLock.Unlock()
}

func (n *SyncHS) getCertForBlockIndex(idx uint64) (*msg.BlockCertificate, bool) {
	n.certMapLock.Lock()
	defer n.certMapLock.Unlock()
	blk, exists := n.certMap[idx]
	return blk, exists
}
