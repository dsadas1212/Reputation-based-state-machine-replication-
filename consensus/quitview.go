package consensus

import (
	"github.com/adithyabhatkajake/libchatter/log"
	"github.com/adithyabhatkajake/libsynchs/msg"
)

// QuitView quits the view
func (n *SyncHS) QuitView() {
	log.Info("Quitting view ", n.view)
	cert := &msg.Certificate{}
	var ids []uint64
	var sigs [][]byte // An array of byte arrays
	idx := uint64(0)
	n.blLock.RLock()
	ids = make([]uint64, len(n.blameMap[n.view]))
	sigs = make([][]byte, len(n.blameMap[n.view]))
	for origin, bl := range n.blameMap[n.view] {
		ids[idx] = origin
		sigs[idx] = bl.Signature
	}
	n.blLock.RUnlock()
	cert.Ids = ids
	cert.Signatures = sigs
	qv := &msg.QuitView{}
	qv.BlCert = cert
	m := &msg.SyncHSMsg{}
	m.Msg = &msg.SyncHSMsg_QV{QV: qv}
	n.Broadcast(m)
	log.Debug("Finished Quitting view", n.view)
}
