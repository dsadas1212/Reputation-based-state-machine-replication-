package consensus

import (
	"github.com/adithyabhatkajake/libchatter/log"
	msg "github.com/adithyabhatkajake/libsynchs/msg"
	pb "github.com/golang/protobuf/proto"
)

// Broadcast broadcasts a protocol message to all the nodes
func (n *SyncHS) Broadcast(m *msg.SyncHSMsg) error {
	n.netMutex.Lock()
	defer n.netMutex.Unlock()
	data, err := pb.Marshal(m)
	if err != nil {
		return err
	}
	// If we fail to send a message to someone, continue
	for idx, s := range n.streamMap {
		_, err = s.Write(data)
		if err != nil {
			log.Error("Error while sending to node", idx)
			log.Error("Error:", err)
			continue
		}
		err = s.Flush()
		if err != nil {
			log.Error("Error while sending to node", idx)
			log.Error("Error:", err)
		}
	}
	return nil
}
