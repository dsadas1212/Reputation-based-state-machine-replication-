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

// 检查当前生成的法定人数证数是否有效
func (n *SyncHS) IsCertValid(bc *msg.BlockCertificate) bool {
	// log.Debug("Received a block certificate -")
	h, _ := bc.GetBlockInfo()
	// 创世区块的法定人数证数一直有效
	if h == chain.EmptyHash {
		return true
	}
	exEmptyBlk := n.bc.BlocksByHash[h]
	ex := exEmptyBlk.ExtHeader.GetExtra()
	//空块的法定人数证书一定有效
	if len(ex) == 1 {
		return true
	}
	//设定法定人数证数评判标准
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
		totalRepInCert = totalRepInCert.Add(totalRepInCert, n.reputationMap[n.view-1][id])
	}
	//检查当前法定证数上的签名所代表的节点们的信誉度是否大于整体信誉度的一半
	if totalRepInCert.Cmp(benchmark) == -1 || totalRepInCert.Cmp(benchmark) == 0 {
		log.Error("invalid cert because lacking reputation")
		return false
	}
	return true
}

func (n *SyncHS) addCert(bc *msg.BlockCertificate, blockNum uint64) {
	log.Debug(n.GetID(), "Adding certificate to block ", blockNum)
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
