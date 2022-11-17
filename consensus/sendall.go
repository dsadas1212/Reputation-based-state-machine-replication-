package consensus

import (
	"github.com/adithyabhatkajake/libchatter/log"
	msg "github.com/adithyabhatkajake/libsynchs/msg"
	pb "github.com/golang/protobuf/proto"
)

// Broadcast broadcasts a protocol message to all the nodes!!
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

// Equivocating block (broadcast)
func (n *SyncHS) EquivocatingBroadcast(m1 *msg.SyncHSMsg, m2 *msg.SyncHSMsg) (e1, e2 error) {
	n.netMutex.Lock()
	defer n.netMutex.Unlock()
	data1, err1 := pb.Marshal(m1)
	data2, err2 := pb.Marshal(m2)
	if err1 != nil || err2 != nil {
		return err1, err2
	}
	// If we fail to send a message to someone, continue
	for idx, s := range n.streamMap {
		if idx%2 == 0 {
			_, err1 = s.Write(data1)
			if err1 != nil {
				log.Error("Error while sending to node", idx)
				log.Error("Error:", err1)
				continue
			}
			err1 = s.Flush()
			if err1 != nil {
				log.Error("Error while sending to node", idx)
				log.Error("Error:", err1)
			}
		} else {
			_, err2 = s.Write(data2)
			if err2 != nil {
				log.Error("Error while sending to node", idx)
				log.Error("Error:", err2)
				continue
			}
			err2 = s.Flush()
			if err2 != nil {
				log.Error("Error while sending to node", idx)
				log.Error("Error:", err2)
			}

		}

	}
	return nil, nil

}
